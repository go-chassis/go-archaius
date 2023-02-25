package configcenter

// Instance is a struct
type Instance struct {
	Status      string   `json:"status"`
	ServiceName string   `json:"serviceName"`
	IsHTTPS     bool     `json:"isHttps"`
	EntryPoints []string `json:"endpoints"`
}

// Members is a struct
type Members struct {
	Instances []Instance `json:"instances"`
}
