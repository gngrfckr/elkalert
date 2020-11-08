package config

//Config структура конфига
type Config struct {
	BindAddr                string `toml:"bind_addr"`
	ElasticConnectionString string `toml:"elastic_connection_string"`
	RulesPath               string `toml:"rules_path"`
}

// NewConfig возвращает структуру конфиг с дефолтными значениями
func NewConfig() *Config {
	return &Config{
		BindAddr:                ":9000",
		ElasticConnectionString: "qa00knode01.ewp:32092",
		RulesPath:               "rules.json",
	}
}
