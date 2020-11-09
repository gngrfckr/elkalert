package config

//Config структура конфига
type Config struct {
	BindAddr                string `toml:"bind_addr"`
	RulesPath               string `toml:"rules_path"`
}

// NewConfig возвращает структуру конфиг с дефолтными значениями
func NewConfig() *Config {
	return &Config{
		BindAddr:                ":9000",
		RulesPath:               "rules.json",
	}
}
