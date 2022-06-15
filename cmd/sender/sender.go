package main

import (
	"flag"
	"log"
	"os"

	"github.com/itaraxa/simple-potato/internal/networkOperation"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	infoLog.Println("START PROGRAMM")

	configFile := flag.String("config", "sender.json", "Configuration file for sender")
	flag.Parse()
	infoLog.Printf("Open configuration file: %s", *configFile)
	config := new(senderConfig)
	err := config.loadConfig(*configFile)
	if err != nil {
		errorLog.Fatalln("Error read configuration file: ", err)
	}
	err = config.checkConfig()
	if err != nil {
		errorLog.Fatalln("INcorrct parametr in configuration file: ", err)
	}

	infoLog.Print("Try to send file")
	err = networkOperation.SendFile("test/tmp/2.txt2", config.SendToAddress, config.SendToPort)
	if err != nil {
		errorLog.Printf("Error sending file: %s", err)
	}
}
