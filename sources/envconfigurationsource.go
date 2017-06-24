package sources

import (
	. "github.com/servicecomb/go-archaius/core"
	"os"
	"strings"
)

type EnvConfigurationSource struct {
	configuration map[string]interface{}
}

func NewEnvConfigurationSource() *EnvConfigurationSource {
	configMap := make(map[string]interface{})
	environ := os.Environ()
	for _, value := range environ {
		rs := []rune(value)
		in := strings.Index(value, "=")
		configMap[string(rs[0:in])] = string(rs[in+1 : len(rs)])
	}
	for _, value := range os.Args {
		in := strings.Index(value, "=")
		if in <= 0 {
			continue
		}
		rs := []rune(value)
		configMap[string(rs[0:in])] = string(rs[in+1 : len(rs)])
	}
	return &EnvConfigurationSource{configuration: configMap}
}

func (this *EnvConfigurationSource) GetPriority() int {
	return 11
}

func (this *EnvConfigurationSource) GetSourceName() string {
	return "EnvironmentVariable"
}

func (this *EnvConfigurationSource) GetConfiguration() map[string]interface{} {
	return this.configuration
}

func (this *EnvConfigurationSource) AddDynamicConfigHandler(callback *ChangesCallback) {
	return
}
