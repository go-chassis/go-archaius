package consistencytest

import (
	"fmt"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/source"
	"github.com/go-chassis/go-archaius/source/mem"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type TestSource struct {
	Name string
	source.ConfigSource
}

func (t *TestSource) GetSourceName() string {
	return t.Name
}

func TestFinalConsistencyAfterConcurrentUpdateDelete(t *testing.T) {
	err := archaius.Init()
	assert.NoError(t, err)

	var ms []source.ConfigSource = make([]source.ConfigSource, 3)
	for i:=0;i<len(ms);i++ {
		ms[i] = &TestSource{
			fmt.Sprintf("testSource%d", i),
			mem.NewMemoryConfigurationSource(),
		}
		ms[i].SetPriority(i)
		err = archaius.AddSource(ms[i])
		assert.NoError(t, err)
	}
	
	assert.NoError(t, err)
	type configOp func (src source.ConfigSource, i int)
	var ops []configOp
	ops = append(ops, func (src source.ConfigSource, i int) {
		src.Set("age", i)
	})
	ops = append(ops, func (src source.ConfigSource, i int) {
		src.Delete("age")
	})
	oneTestRun := func () bool {
		wg := sync.WaitGroup{}
		wg.Add(len(ms))
		for _, src := range ms {
			go func() {
				for i:=0;i<10;i++ {
					ops[rand.Int()%len(ops)](src, i)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		time.Sleep(10*time.Microsecond)
		var finalVal interface{}
		for _, src := range ms {
			tmp, err := src.GetConfigurationByKey("age")
			if err == nil {
				finalVal = tmp
				break
			}
		}
		val := archaius.Get("age")
		return assert.Equal(t, finalVal, val)
	}
	for i:=0;i<10000;i++ {
		t.Logf("test run %05d", i)
		if !oneTestRun() {
			break
		}
	}
}
