package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/itaraxa/simple-potato/internal/fileOperation"
	"github.com/itaraxa/simple-potato/internal/networkOperation"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
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
		errorLog.Fatalln("Incorrect parameter in configuration file: ", err)
	}

	infoLog.Printf("Switch to temporary folder: %s", config.DirectoryForTemporaryFiles)
	err = os.Chdir(config.DirectoryForTemporaryFiles)
	if err != nil {
		errorLog.Fatalf("Cannot switch to temporary folder: %s", config.DirectoryForTemporaryFiles)
	}

	infoLog.Println("Search files in current work directory")
	fileNames, err := fileOperation.ScanDir(".")
	if err != nil {
		errorLog.Println("scanning directory error: ", err)
	}

	infoLog.Print("Filter allowed file types")
	fileNames, _ = fileOperation.FilterFiles(fileNames, config.AllowedFileTypes)

	// Sending files
	var SessionID uint32
	SessionID = 0

	infoLog.Printf("Init connection to %s", fmt.Sprintf("%s:%s", config.SendToAddress, config.SendToPort))
	con, err := net.Dial("udp4", fmt.Sprintf("%s:%s", config.SendToAddress, config.SendToPort))
	if err != nil {
		errorLog.Panicf("Error init connection: %s", err)
	}
	defer con.Close()

	for _, fileName := range fileNames {
		infoLog.Printf("Start send file: %s", fileName)
		t := new(fileOperation.MetaFile)
		if err := t.Init(fileName); err != nil {
			errorLog.Printf("Error reading file: %s", err)
		}

		SessionID += 1

		infoLog.Printf("Start sending command packet: SessionID = %x", SessionID)
		cm := new(networkOperation.ControlMsg)
		if err := cm.PayLoad(SessionID, 1, []byte(fileName)); err != nil {
			errorLog.Printf("Error create payload for message: %s", err)
		}
		if err := cm.Send(con); err != nil {
			errorLog.Printf("Error send control message: %s", err)
		}
		infoLog.Printf("End sending command packet: SessionID = %x", SessionID)

		infoLog.Printf("Start sending metadata packet: SessionID = %x, File = %s", SessionID, fileName)
		mm := new(networkOperation.MetadataMsg)
		if err := mm.PayLoad(SessionID, 9, *t); err != nil {
			errorLog.Printf("Error create payload for message: %s", err)
		}
		if err := mm.Send(con); err != nil {
			errorLog.Printf("Error send metadata message: %s", err)
		}
		infoLog.Printf("End sending metadata packet: SessionID = %x, File = %s", SessionID, fileName)

		infoLog.Printf("Start sending data packets: SessionID = %x, File = %s", SessionID, fileName)
		dm := new(networkOperation.DataMsg)
		if err := dm.PayLoad(SessionID, 9, *t); err != nil {
			errorLog.Printf("Error create payload for message: %s", err)
		}
		if err := dm.Send(con); err != nil {
			errorLog.Printf("Error send data message: %s", err)
		}
		infoLog.Printf("End sending data packets: SessionID = %x, File = %s", SessionID, fileName)

		infoLog.Printf("Start sending command packet: SessionID = %x", SessionID)
		cm2 := new(networkOperation.ControlMsg)
		if err := cm2.PayLoad(SessionID, 2, []byte("")); err != nil {
			errorLog.Printf("Error create payload for message: %s", err)
		}
		if err := cm2.Send(con); err != nil {
			errorLog.Printf("Error send control message: %s", err)
		}
		infoLog.Printf("End sending command packet: SessionID = %x", SessionID)

	}
	infoLog.Printf("End connection to %s", fmt.Sprintf("%s:%s", config.SendToAddress, config.SendToPort))

	infoLog.Println("END PROGRAMM")
}
