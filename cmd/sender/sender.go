package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/itaraxa/simple-potato/internal/fileOperation"
	"github.com/itaraxa/simple-potato/internal/session"
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
	var SessionID uint32 = 007

	remoteAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", config.SendToAddress, config.SendToPort))
	if err != nil {
		errorLog.Fatalf("Incorrect destination address: %s", fmt.Sprintf("%s:%s", config.SendToAddress, config.SendToPort))
	}
	localAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", config.SendFromAddress, "0"))
	if err != nil {
		errorLog.Fatalf("Incorrect local address: %s", fmt.Sprintf("%s:%s", config.SendFromAddress, "0"))
	}
	con, err := net.DialUDP("udp4", localAddr, remoteAddr)
	if err != nil {
		errorLog.Fatalf("Error init connection: %s", err)
	}
	defer con.Close()

	for _, fileName := range fileNames {
		infoLog.Printf("Start send file: %s", fileName)

		SessionID += 1

		s := session.NewSession(SessionID)

		err = s.ReadFile(fileName)
		if err != nil {
			errorLog.Printf("Error reading file: %s : %s", fileName, err)
			continue
		}

		err = s.SendFile(con)
		if err != nil {
			errorLog.Printf("Error sending file %s to %s : %s", fileName, con.RemoteAddr().String(), err)
			continue
		}

		time.Sleep(1000 * time.Millisecond)

	}

	infoLog.Printf("End connection to %s", fmt.Sprintf("%s:%s", config.SendToAddress, config.SendToPort))

	infoLog.Println("END PROGRAMM")
}
