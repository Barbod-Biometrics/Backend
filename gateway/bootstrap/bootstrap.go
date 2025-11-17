package bootstrap

type Config struct {
	Constants *Constants
	Env       *Env
}

func Run() *Config {
	return &Config{
		Constants: NewConstants(),
		Env:       NewEnvironment(),
	}
}
