package goarchaius

type Result struct {
	Complete    map[string]interface{}
	Added       map[string]interface{}
	Changed     map[string]interface{}
	Deleted     map[string]interface{}
	Incremental bool
}