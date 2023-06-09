package config

import "flag"

var FlagRunAddr string
var FlagBaseAddr string

func ParseFlags() {

	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagBaseAddr, "b", ":8080", "base address for urls")

	flag.Parse()
}
