package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
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

	infoLog.Printf("Switch to temporary folder: %s", config.DirectoryForNewFiles)
	err = os.Chdir(config.DirectoryForNewFiles)
	if err != nil {
		errorLog.Fatalf("Cannot switch to temporary folder: %s", config.DirectoryForNewFiles)
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

		err = MoveFile(fileName, config.DirectoryForUploadedFiles)
		if err != nil {
			errorLog.Printf("error moving file %s : %s", fileName, err)
		}
		time.Sleep(1000 * time.Millisecond)

	}

	infoLog.Printf("End connection to %s", fmt.Sprintf("%s:%s", config.SendToAddress, config.SendToPort))

	infoLog.Println("END PROGRAMM")
}

func MoveFile(src, dstDir string) error {
	sep := string(os.PathSeparator)
	path := strings.Split(src, sep)
	fileName := path[len(path)-1]
	fullDstDir := dstDir + sep + strings.Join(path[1:len(path)-1], sep)

	if _, err := os.Stat(fullDstDir); os.IsNotExist(err) {
		err2 := os.MkdirAll(fullDstDir, 0760)
		if err2 != nil {
			return fmt.Errorf("cannot create directory: %s", err2)
		}
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(fullDstDir + sep + fileName)
	if err != nil {
		return err
	}
	defer dest.Close()

	buf := make([]byte, 4*1024)

	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := dest.Write(buf[:n]); err != nil {
			return err
		}
	}

	source.Close()
	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("cannot remove file %s: %s", src, err)
	}

	// fmt.Printf("Directory: %s -> %s\n", src, fullDstDir)
	return nil
}
