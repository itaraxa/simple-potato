package main

import (
	"flag"
	"log"

	"github.com/itaraxa/simple-potato/pkg/fileOperation"
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

	myFile := new(fileOperation.MyFile)
	err = myFile.Init("test/file1.txt")
	if err != nil {
		log.Println("Error opening file: ", err)
	}
	myFile.PrettyOut()

}
