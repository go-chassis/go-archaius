package benchmark

import (
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"testing"
)

var f1Content = `
age0: 1
age1: 1
age2: 1
age3: 1
age4: 1
age5: 1
age6: 1
age7: 1
age8: 1
age9: 1
age10: 1
age11: 1
age12: 1
age13: 1
age14: 1
age15: 1
age16: 1
age17: 1
age18: 1
age19: 1
age20: 1
age21: 1
age22: 1
age23: 1
age24: 1
age25: 1
age26: 1
age27: 1
age28: 1
age29: 1
age30: 1
age31: 1
age32: 1
age33: 1
age34: 1
age35: 1
age36: 1
age37: 1
age38: 1
age39: 1
age40: 1
age41: 1
age42: 1
age43: 1
age44: 1
age45: 1
age46: 1
age47: 1
age48: 1
age49: 1
age50: 1
age51: 1
age52: 1
age53: 1
age54: 1
age55: 1
age56: 1
age57: 1
age58: 1
age59: 1
age60: 1
age61: 1
age62: 1
age63: 1
age64: 1
age65: 1
age66: 1
age67: 1
age68: 1
age69: 1
age70: 1
age71: 1
age72: 1
age73: 1
age74: 1
age75: 1
age76: 1
age77: 1
age78: 1
age79: 1
age80: 1
age81: 1
age82: 1
age83: 1
age84: 1
age85: 1
age86: 1
age87: 1
age88: 1
age89: 1
age90: 1
age91: 1
age92: 1
age93: 1
age94: 1
age95: 1
age96: 1
age97: 1
age98: 1
age99: 0
`

func createFile(content string, name string, dir string) string {
	filename := filepath.Join(dir, name)
	f1, _ := os.Create(filename)
	_, _ = io.WriteString(f1, content)
	f1.Close()
	return filename
}

func TestMain(m *testing.M) {
	filename := createFile(f1Content, "f1.yaml", "/tmp")
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
