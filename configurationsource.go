package goarchaius

type ConfigurationSource interface {
	//poll(bool initial, interface{} checkPoint) string
	GetConfiguration() map[string]interface{}
	AddDispatcher(dispatcher *Dispatcher)
}
