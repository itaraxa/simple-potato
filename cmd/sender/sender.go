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
		errorLog.Fatalln("Incorrect parameter in configuration file: ", err)
	}

	infoLog.Print("Try to send file")
	err = networkOperation.SendFile("test/tmp/2.txt2", config.SendToAddress, config.SendToPort)
	if err != nil {
		errorLog.Printf("Error sending file: %s", err)
	}

	// infoLog.Printf("Search files in temporary folder: %s", config.DirectoryForTemporaryFiles)
	// files, err := fileOperation.ScanDir(config.DirectoryForTemporaryFiles)
	// if err != nil {
	// 	errorLog.Println("scanning directory error: ", err)
	// }
	// fileOperation.PathCleaner(files, config.DirectoryForTemporaryFiles)

	// files, _ = fileOperation.FilterFiles(files)

	// for _, file := range files {

	// }

	// for _, file := range files {
	// 	infoLog.Printf("Move file SRC=%s DST=%s", file, config.DirectoryForUploadedFiles+string(os.PathSeparator)+file)

	// 	err = fileOperation.MoveFile(config.DirectoryForTemporaryFiles+string(os.PathSeparator)+file, config.DirectoryForUploadedFiles+string(os.PathSeparator)+file)
	// 	if err != nil {
	// 		errorLog.Printf("error moving file: %s : %s", file, err)
	// 	}
	// }

}
