package goarchaius

type ConfigurationManager struct {
	Sources       []ConfigurationSource
	Configuration map[string]interface{}
	dispatcher    *Dispatcher
}

func (this *ConfigurationManager) AddSource(s ConfigurationSource) {
	s.AddDispatcher(this.dispatcher)
	this.Sources = append(this.Sources, s)
	for k, v := range s.GetConfiguration() {
		this.Configuration[k] = v
	}
}

func (this *ConfigurationManager) Refresh() {
	this.Configuration = make(map[string]interface{})
	for _, s := range this.Sources {
		for k, v := range s.GetConfiguration() {
			this.Configuration[k] = v
		}
	}
}

func NewConfigurationManager() *ConfigurationManager {
	return &ConfigurationManager{Configuration: make(map[string]interface{}), dispatcher: NewDispatcher()}
}
