package source

import (
	"os"
	"testing"
)

func Test_GetConfiguration(t *testing.T) {
	os.Setenv("zqtest", "a=b=c")
	s := NewEnvConfigurationSource()
	config := s.GetConfiguration()
	v := config["zqtest"]
	if v != "a=b=c" {
		t.Error("Failed to get config items from env var")
	}
}
