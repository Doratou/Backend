package bootstrap

type Config struct {
	Constant    *Constant
	Environment *Environment
}

func Run() (*Config, error) {
	environment, err := NewEnvironment()
	if err != nil {
		return nil, err
	}
	return &Config{Constant: NewConstant(), Environment: environment}, nil
}
