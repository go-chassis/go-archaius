package archaius

import "github.com/go-chassis/go-archaius/core"

//Options hold options
type Options struct {
	RequiredFiles    []string
	OptionalFiles    []string
	ConfigCenterInfo ConfigCenterInfo
	EventListeners   []core.EventListener
	UseCLISource     bool
	UseENVSource     bool
}

//Option is a func
type Option func(options *Options)

//WithRequiredFiles tell archaius to manage files, if not exist will return error
func WithRequiredFiles(f []string) Option {
	return func(options *Options) {
		options.RequiredFiles = f
	}
}

//WithOptionalFiles tell archaius to manage files, if not exist will not return error
func WithOptionalFiles(f []string) Option {
	return func(options *Options) {
		options.OptionalFiles = f
	}
}

//WithConfigCenter accept the information for initiating a config center client and archaius config source
func WithConfigCenter(cci ConfigCenterInfo) Option {
	return func(options *Options) {
		options.ConfigCenterInfo = cci
	}
}

//WithCommandLineSource enable cmd line source
func WithCommandLineSource() Option {
	return func(options *Options) {
		options.UseCLISource = true
	}
}

//WithENVSource enable env source
func WithENVSource() Option {
	return func(options *Options) {
		options.UseENVSource = true
	}
}

//WithEventListeners will register listeners to archaius runtime
func WithEventListeners(ls ...core.EventListener) Option {
	return func(options *Options) {
		options.EventListeners = ls
	}
}
