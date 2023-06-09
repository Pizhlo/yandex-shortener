package config

import (
	"flag"
	"fmt"
	"regexp"
)

var FlagRunAddr string
var FlagBaseAddr string
//var FlagBasePort string

var baseAddrRegexp = regexp.MustCompile(":[0-9]{1,}")

func ParseFlags() {

	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	//flag.StringVar(&FlagBasePort, "p", "8080", "port for urls")
	flag.StringVar(&FlagBaseAddr, "b", "http://localhost:8080", "base address for urls")

	flag.Parse()

	defaultHost := fmt.Sprintf("http://localhost:%s", FlagBaseAddr)

	if baseAddrRegexp.MatchString(FlagBaseAddr) {
		FlagBaseAddr = defaultHost
	}

	fmt.Println("FlagRunAddr = ", FlagRunAddr)
	fmt.Println("FlagBaseAddr = ", FlagBaseAddr)
	//fmt.Println("FlagBasePort = ", FlagBasePort)
}
