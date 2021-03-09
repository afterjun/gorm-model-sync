package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Db Db `yaml:"db"`
}
type Db struct {
	Address string `yaml:"address"`
}

var Conf = &Config{}

func init() {
	data, err := ioutil.ReadFile("./config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("读取 'config.yaml' 失败: %v\n", err))
	}
	yaml.Unmarshal(data, &Conf)
}
