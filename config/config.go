package config

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

var FlagRunAddr string
var FlagBaseAddr string

var baseAddrRegexp = regexp.MustCompile("^:[0-9]{1,}$")

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func ParseConfigAndFlags() error {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagBaseAddr, "b", "http://localhost:8080", "base address for urls")

	flag.Parse()

	defaultHost := fmt.Sprintf("http://localhost:%s", FlagBaseAddr)
	if baseAddrRegexp.MatchString(FlagBaseAddr) {
		FlagBaseAddr = defaultHost
	}

	if val, ok := os.LookupEnv("BASE_URL"); ok {
		FlagBaseAddr = val
	}
	
	if val, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		FlagRunAddr = val
	}

	fmt.Println("FlagRunAddr = ", FlagRunAddr)
	fmt.Println("FlagBaseAddr = ", FlagBaseAddr)

	return nil
}
