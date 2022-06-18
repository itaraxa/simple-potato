package main

import (
	"flag"
	"log"
	"os"
)

func main() {

	// Initialize logging
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	infoLog.Println("START PROGRAMM")

	// Read configuration file
	configFile := flag.String("config", "receiver.json", "Configuration file for receiver")
	flag.Parse()
	infoLog.Printf("Open configuration file: %s", *configFile)
	config := new(receiverConfig)
	err := config.loadConfig(*configFile)
	if err != nil {
		errorLog.Fatalln("Error read configuration file: ", err)
	}
	err = config.checkConfig()
	if err != nil {
		errorLog.Fatalln("Error in configuration file: ", err)
	}

	infoLog.Println("END PROGRAMM")

}
