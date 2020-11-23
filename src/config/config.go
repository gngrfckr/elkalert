package config

//Config структура конфига
type Config struct {
	BindAddr                string `toml:"bind_addr"`
	RulesPath               string `toml:"rules_path"`
	RequestSize             int    `toml:"es_request_size"`
}

// NewConfig возвращает структуру конфиг с дефолтными значениями
func NewConfig() *Config {
	return &Config{
		BindAddr:                ":9000",
		RulesPath:               "rules.json",
		RequestSize:             10000,
	}
}
