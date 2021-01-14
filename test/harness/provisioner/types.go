package provisioner

type Provisioner interface {
	Provision() error
}

type Config struct {
	Driver string                 `json:"use"`
	Config map[string]interface{} `json:"config"`
}
