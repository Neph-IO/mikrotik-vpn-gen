package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type configStruct struct {
	Mikrotik   mikrotik   `yaml:"routeros"`
	GlobalConf globalConf `yaml:"globalconf"`
	Vpncreator vpncreator `yaml:"vpncreator"`
}

type mikrotik struct {
	Debug    bool   `yaml:"debug"`
	Address  string `yaml:"address"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Async    bool   `yaml:"async"`
	Tls      bool   `yaml:"tls"`
	CertName string `yaml:"certname"`
}

type globalConf struct {
	AllowedOrigin string `yaml:"allowedOrigin"`
	ApiPort       string `yaml:"apiport"`
}

type vpncreator struct {
	Nginxfolder string            `yaml:"nginxfolder"`
	CaName      string            `yaml:"caname"`
	ProfileMap  map[string]string `yaml:"profilemap"`
	ValidTime   string            `yaml:"validtime"`
	Keysize     string            `yaml:"keysize"`
	Countrycode string            `yaml:"countrycode"`
}

var Conf configStruct

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Error reading config file : %w", err)
	}

	if err := yaml.Unmarshal(data, &Conf); err != nil {
		return fmt.Errorf("error parsing YAML : %w", err)
	}

	return nil
}
