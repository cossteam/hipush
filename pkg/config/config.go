package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func NewConfig() *viper.Viper {
	envConf := os.Getenv("APP_CONF")
	if envConf == "" {
		flag.StringVar(&envConf, "config", "config/local.yml", "config path, eg: -config config/local.yml")
		flag.Parse()
	}
	if envConf == "" {
		envConf = "config/local.yml"
	}
	fmt.Println("load conf file:", envConf)
	return getConfig(envConf)
}

func getConfig(path string) *viper.Viper {
	conf := viper.New()
	conf.SetConfigFile(path)
	err := conf.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return conf
}

func Load() *Config {
	envConf := os.Getenv("APP_CONF")
	if envConf == "" {
		flag.StringVar(&envConf, "config", "config/local.yml", "config path, eg: -config config/local.yml")
		flag.Parse()
	}
	if envConf == "" {
		envConf = "config/local.yml"
	}
	fmt.Println("load conf file:", envConf)
	data, err := os.ReadFile(envConf)
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}

type Config struct {
	Env      string              `yaml:"env"`
	HTTP     HTTPConfig          `yaml:"http"`
	Security Security            `yaml:"security"`
	Data     DataConfig          `yaml:"data"`
	Log      LogConfig           `yaml:"log"`
	Apps     []ProvisionerConfig `yaml:"apps"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type Security struct {
	APISign APISign `yaml:"apiSign"`
	JWT     JWT     `yaml:"jwt"`
}

type APISign struct {
	AppKey      string `yaml:"appKey"`
	AppSecurity string `yaml:"appSecurity"`
}

type JWT struct {
	Key string `yaml:"key"`
}

type DataConfig struct {
	MySQL MySQLConfig `yaml:"mysql"`
	Redis RedisConfig `yaml:"redis"`
}

type MySQLConfig struct {
	User string `yaml:"user"`
}

type RedisConfig struct {
	Addr         string        `yaml:"addr"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

type LogConfig struct {
	LogLevel    string `yaml:"logLevel"`
	Encoding    string `yaml:"encoding"`
	LogFileName string `yaml:"logFileName"`
	MaxBackups  int    `yaml:"maxBackups"`
	MaxAge      int    `yaml:"maxAge"`
	MaxSize     int    `yaml:"maxSize"`
	Compress    bool   `yaml:"compress"`
}

type ProvisionerConfig struct {
	Enabled   bool              `yaml:"enabled"`
	Platform  string            `yaml:"platform"`
	AppName   string            `yaml:"appName"`
	AppConfig map[string]string `yaml:"appConfig"`
}
