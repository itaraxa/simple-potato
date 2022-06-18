package main

import (
	"flag"
	"log"
	"os"

	"github.com/itaraxa/simple-potato/internal/fileOperation"
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

	// infoLog.Print("Try to send file")
	// err = networkOperation.SendFile("test/tmp/2.txt2", config.SendToAddress, config.SendToPort)
	// if err != nil {
	// 	errorLog.Printf("Error sending file: %s", err)
	// }

	infoLog.Println("Search files in current work directory")
	fileNames, err := fileOperation.ScanDir(".")
	if err != nil {
		errorLog.Println("scanning directory error: ", err)
	}

	// fmt.Print("\nСписок файлов до фильтра:")
	// for _, fileName := range fileNames {
	// 	fmt.Println(fileName)
	// }

	infoLog.Print("Filter allowed file types")
	fileNames, _ = fileOperation.FilterFiles(fileNames, config.AllowedFileTypes)

	// fmt.Println("\nСписок файлов после фильтра:")
	// for _, fileName := range fileNames {
	// 	fmt.Println(fileName)
	// }

	// Sending files
	for _, fileName := range fileNames {
		t := new(fileOperation.MetaFile)
		if err := t.Init(fileName); err != nil {
			errorLog.Printf("Error reading file: %s", err)
		}
		t.PrettyOut()

	}

	// for _, file := range fileNames {
	// 	infoLog.Printf("Move file SRC=%s DST=%s", file, config.DirectoryForUploadedFiles+string(os.PathSeparator)+file)

	// 	err = fileOperation.MoveFile(config.DirectoryForTemporaryFiles+string(os.PathSeparator)+file, config.DirectoryForUploadedFiles+string(os.PathSeparator)+file)
	// 	if err != nil {
	// 		errorLog.Printf("error moving file: %s : %s", file, err)
	// 	}
	// }
	infoLog.Println("END PROGRAMM")
}
