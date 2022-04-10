package common

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"sync"
)

type ServerConfig struct {
	NsqdAddr       string            `yaml:"nsqd"`
	EtcdAddr       string            `yaml:"etcd"`
	SupportTopics  []string          //不对应yaml字段，手动处理
	TorchServeAddr map[string]string `yaml:"ts"`
}

type WebConfig struct {
	WebAddr  string            `yaml:"web"`
	EtcdAddr string            `yaml:"etcd"`
	NsqdAddr string            `yaml:"nsqd"`
	Strategy map[string]string `yaml:"strategy"`
}

var sconfig ServerConfig
var wconfig WebConfig
var once sync.Once

func GetServerConfig(path string) *ServerConfig {
	once.Do(func() {
		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
		}
		err = yaml.Unmarshal(yamlFile, &sconfig)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
		// 处理SupportTopics字段
		for k, _ := range sconfig.TorchServeAddr {
			sconfig.SupportTopics = append(sconfig.SupportTopics, k)
		}
	})
	return &sconfig
}

func GetWebConfig(path string) *WebConfig {
	once.Do(func() {
		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
		}
		err = yaml.Unmarshal(yamlFile, &wconfig)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
	})
	return &wconfig
}

func ParseYaml(path string) map[string]string {
	m := make(map[string]string)
	data := ReadFile(path)
	err := yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil
	}
	return m
}
