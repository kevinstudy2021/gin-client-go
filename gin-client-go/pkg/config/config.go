package config

import (
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	"strconv"
)

var keyMap map[KeyName]string

type Config struct {
	Server Server
}

type Server struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func init() {
	var config Config
	yamlFile, err := ioutil.ReadFile("./.gin-client-go.yaml")
	if err != nil {
		klog.Fatal(err)
		return
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		klog.Fatal(err)
		return
	}
	keyMap = make(map[KeyName]string)
	keyMap[ServerName] = config.Server.Name
	keyMap[ServerHost] = config.Server.Host
	keyMap[SercerPort] = config.Server.Port

}

func GetString(keyName KeyName) string {
	return keyMap[keyName]
}

func GetInt(keyName KeyName) int {
	intstr := keyMap[keyName]
	if intstr == "" {
		klog.Fatal("GetInt not read config ====>" +keyName)
		return -1
	}
	v, err := strconv.Atoi(intstr)
	if intstr == ""{
		klog.Fatal(err)
		return -1
	}
	return v
}