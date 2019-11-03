package local

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strings"
)

const (
	defaultLogPath = "./conf.yaml"
)

type LocalConf struct {
	Discover struct {
		Name       string `yaml:"Name"`
		ListenHost string `yaml:"ListenHost"`
		ListenPort int    `yaml:"ListenPort"`

		Etcd struct {
			Name      string   `yaml:"Name"`
			Endpoints []string `yaml:"Endpoints"`
			Username  string   `yaml:"Username"`
			Password  string   `yaml:"Password"`

			AutoSyncInterval     int `yaml:"AutoSyncInterval"`
			DialTimeout          int `yaml:"DialTimeout"`
			DialKeepAliveTime    int `yaml:"DialKeepAliveTime"`
			DialKeepAliveTimeout int `yaml:"DialKeepAliveTimeout"`
		} `yaml:"Etcd"`
	} `yaml:"Discover"`

	Server struct {
		Name       string `yaml:"Name"`
		ListenHost string `yaml:"ListenHost"`
		ListenPort int    `yaml:"ListenPort"`
		Mysql      struct {
			User     string `yaml:"User"`
			Password string `yaml:"Password"`
			Addr     string `yaml:"Addr"`
			DBName   string `yaml:"DBName"`
		} `yaml:"Mysql"`
	} `yaml:"Server"`
}

var (
	Conf *LocalConf
)

func ReadConfig(path ...string) *LocalConf {
	var f *os.File
	var err error
	var config LocalConf
	var isProductionEnv bool
	var configFilePath string

	// env IS_PRODUCTION
	if os.Getenv("IS_PRODUCTION") == "1" {
		isProductionEnv = true
	} else {
		isProductionEnv = false
	}

	// env CONFIG_PATH
	if len(path) == 1 {
		configFilePath = path[0]
	} else if len(path) == 0 {
		configFilePath = os.Getenv("CONFIG_PATH")
		if configFilePath == "" {
			configFilePath = defaultLogPath
		}
	} else {
		log.Fatal("only one path could be passed in")
	}

	if !isProductionEnv {
		fp := strings.Split(configFilePath, ".")
		fpLen := len(fp)
		if fpLen > 1 {
			fp[fpLen-2] += "_test"
			configFilePath = strings.Join(fp, ".")
		} else if fpLen == 1 {
			configFilePath += "_test"
		} else {
			log.Fatal(fmt.Sprintf("lack of config file path, %s", configFilePath))
		}
	}
	f, err = os.OpenFile(configFilePath, os.O_RDONLY, 0666)
	defer func(fd *os.File) {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}(f)
	if err != nil {
		log.Fatal(err)
	}

	length, _ := f.Seek(0, 2)
	conf := make([]byte, length)
	if _, err := f.Seek(0, 0); err != nil {
		panic(err)
	}
	_, err = f.Read(conf)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(conf, &config)
	if err != nil {
		log.Fatal(err)
	}
	return &config
}

func init() {
	Conf = ReadConfig()
}
