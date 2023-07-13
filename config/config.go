package config

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

var baseAddrRegexp = regexp.MustCompile("^:[0-9]{1,}$")

type Config struct {
	FlagRunAddr    string
	FlagBaseAddr   string
	FlagLogLevel   string
	FlagPathToFile string
	FlagSaveToFile bool
}

func ParseConfigAndFlags() Config {
	var conf Config

	flag.StringVar(&conf.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&conf.FlagBaseAddr, "b", "http://localhost:8080", "base address for urls")
	flag.StringVar(&conf.FlagLogLevel, "l", "info", "log level")
	flag.StringVar(&conf.FlagPathToFile, "f", "/tmp/short-url-db.json", "file to save short urls")

	flag.Parse()

	defaultHost := fmt.Sprintf("http://localhost:%s", conf.FlagBaseAddr)
	if baseAddrRegexp.MatchString(conf.FlagBaseAddr) {
		conf.FlagBaseAddr = defaultHost
	}

	if val, ok := os.LookupEnv("BASE_URL"); ok {
		conf.FlagBaseAddr = val
	}

	if val, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		conf.FlagRunAddr = val
	}

	if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
		conf.FlagLogLevel = val
	}

	if val, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		conf.FlagPathToFile = val
	}

	if conf.FlagPathToFile != "" {
		conf.FlagSaveToFile = true
	} 

	fmt.Println("FlagRunAddr = ", conf.FlagRunAddr)
	fmt.Println("FlagBaseAddr = ", conf.FlagBaseAddr)
	fmt.Println("FlagLogLevel = ", conf.FlagLogLevel)
	fmt.Println("FlagFileStorage = ", conf.FlagPathToFile)
	fmt.Println("FlagSaveToFile = ", conf.FlagSaveToFile)

	return conf
}
