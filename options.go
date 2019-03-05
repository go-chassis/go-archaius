package archaius

import (
	"github.com/go-chassis/go-archaius/core"
	"github.com/go-chassis/go-archaius/sources/file-source"
	"github.com/go-chassis/go-cc-client"
)

//Options hold options
type Options struct {
	RequiredFiles    []string
	OptionalFiles    []string
	FileHandler      filesource.FileHandler
	ConfigCenterInfo ConfigCenterInfo
	ConfigClient     ccclient.ConfigClient
	UseCLISource     bool
	UseENVSource     bool
	ExternalSource   core.ConfigSource
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

//WithDefaultFileHandler let user custom handler
func WithDefaultFileHandler(handler filesource.FileHandler) Option {
	return func(options *Options) {
		options.FileHandler = handler
	}
}

//WithConfigCenterSource accept the information for initiating a config center source,
//ConfigCenterInfo is required if you want to use config center source
//client is optional,if client is nil, archaius will create one based on ConfigCenterInfo
//config client will be injected into config source as a client to interact with a config server
func WithConfigCenterSource(cci ConfigCenterInfo, ccc ccclient.ConfigClient) Option {
	return func(options *Options) {
		options.ConfigCenterInfo = cci
	}
}

//WithCommandLineSource enable cmd line source
//archaius will read command line params as key value
func WithCommandLineSource() Option {
	return func(options *Options) {
		options.UseCLISource = true
	}
}

//WithENVSource enable env source
//archaius will read ENV as key value
func WithENVSource() Option {
	return func(options *Options) {
		options.UseENVSource = true
	}
}

//WithExternalSource accept the information for initiating a External source
func WithExternalSource(e core.ConfigSource) Option {
	return func(options *Options) {
		options.ExternalSource = e
	}
}

//WithMemorySource accept the information for initiating a Memory source
func WithMemorySource(e core.ConfigSource) Option {
	return func(options *Options) {
		options.ExternalSource = e
	}
}

//FileOptions for AddFile func
type FileOptions struct {
	Handler filesource.FileHandler
}

//FileOption is a func
type FileOption func(options *FileOptions)

//WithFileHandler use custom handler
func WithFileHandler(h filesource.FileHandler) FileOption {
	return func(options *FileOptions) {
		options.Handler = h
	}

}
