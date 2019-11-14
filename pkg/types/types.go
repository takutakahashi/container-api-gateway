package types

func NewConfig() Config {
	return Config{
		Endpoints: []Endpoint{},
	}
}

type Config struct {
	Endpoints []Endpoint `yaml:"endpoints"`
}
type Endpoint struct {
	Path      string    `yaml:"path"`
	Method    string    `yaml:"method"`
	Async     bool      `yaml:"async"`
	Params    Params    `yaml:"params"`
	Container Container `yaml:"container"`
}

type Container struct {
	Image   string   `yaml:"image"`
	Command []string `yaml:"command"` // contains go template

}

type Param struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Value    string
	Optional bool `yaml:"optional"`
}

type Params []Param

func (ps *Params) BuildCommand(base []string) []string {
	return base
}

// BuildCommand build command with params
func (e *Endpoint) BuildCommand() []string {
	return e.Container.Command
}
