package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/itaraxa/simple-potato/internal/fileOperation"
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

	// Detect and filter files for transmition

	infoLog.Printf("Search files in temporary folder: %s", config.Directory_for_temporary_files)
	files, err := fileOperation.ScanDir(config.Directory_for_temporary_files)
	if err != nil {
		errorLog.Println("scanning directory error: ", err)
	}
	fileOperation.PathCleaner(files, config.Directory_for_temporary_files)

	files, _ = fileOperation.FilterFiles(files)

	for _, file := range files {
		infoLog.Printf("Move file SRC=%s DST=%s", file, config.Directory_for_downloaded_files+string(os.PathSeparator)+file)

		err = fileOperation.MoveFile(config.Directory_for_temporary_files+string(os.PathSeparator)+file, config.Directory_for_downloaded_files+string(os.PathSeparator)+file)
		if err != nil {
			errorLog.Printf("error moving file: %s : %s", file, err)
		}
		fmt.Printf("%s = %v\n", file, []byte(file))
	}

	// Передача файла по сети

	infoLog.Println("END PROGRAMM")

}
