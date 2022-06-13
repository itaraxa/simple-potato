package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	configFile := flag.String("config", "receiver.json", "Configuration file in json")
	flag.Parse()

	config := new(receiverConfig)
	err := config.loadConfig(*configFile)
	if err != nil {
		log.Fatalln("Error read configuration file: ", err)
	}
	err = config.checkConfig()
	if err != nil {
		log.Fatalln("Error in configuration file: ", err)
	}
	fmt.Println(config.Log_File)
	fmt.Println(config.Log_Level)
}
