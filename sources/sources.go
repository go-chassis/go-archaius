package sources

import (
	"github.com/servicecomb/go-archaius/core"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	core.DefaultSources = append(core.DefaultSources, NewEnvConfigurationSource())
	yamlSource, _ := NewYamlConfigurationSource("Default", getCurrentDirectory()+"/config.yaml")
	core.DefaultSources = append(core.DefaultSources, yamlSource)
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		//log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
