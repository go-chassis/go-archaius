package benchmark

import (
	"bytes"
	"fmt"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func createFile(content string, name string, dir string) string {
	filename := filepath.Join(dir, name)
	f1, _ := os.Create(filename)
	_, _ = io.WriteString(f1, content)
	f1.Close()
	return filename
}

func TestMain(m *testing.M) {
	var buff bytes.Buffer
	for i:=0;i<99;i++ {
		buff.WriteString(fmt.Sprintf("age%d: 1\n", i))
	}
	buff.WriteString(fmt.Sprintf("age99: 0\n"))
	filename := createFile(buff.String(), "f1.yaml", "/tmp")
	defer os.Remove(filename)

	err := archaius.Init(archaius.WithOptionalFiles([]string{filename}))
	if err != nil {
		os.Exit(-1)
	}
	os.Exit(m.Run())
}

func BenchmarkFileSource(b *testing.B) {
	s := 0
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s += archaius.GetInt("age99", 1)
	}
	assert.Zero(b, s)
}

func BenchmarkFileSourceParallelism(b *testing.B) {
	s := 0
	b.ResetTimer()
	b.SetParallelism(200)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s += archaius.GetInt("age99", 1)
		}
	})
	assert.Zero(b, s)
}
