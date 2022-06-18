package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/itaraxa/simple-potato/internal/fileOperation"
	"github.com/itaraxa/simple-potato/internal/networkOperation"
)

func main() {

	// Initialize logging
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	infoLog.Println("START PROGRAMM")

	// Read configuration file
	configFile := flag.String("config", "receiver.json", "Configuration file for receiver")
	flag.Parse()

	config := new(receiverConfig)
	err := config.loadConfig(*configFile)
	if err != nil {
		errorLog.Fatalln("Error read configuration file: ", err)
	}
	err = config.checkConfig()
	if err != nil {
		errorLog.Fatalln("Error in configuration file: ", err)
	}
	infoLog.Printf("Configuration file %s was openned", *configFile)

	lAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", "0.0.0.0", config.LocalPort))
	if err != nil {
		errorLog.Fatalf("Cannot resolve local address: %s\n", fmt.Sprintf(config.LocalAddress, config.LocalPort))
	}
	// rAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", config.RemoteAddress, "0"))
	// if err != nil {
	// 	errorLog.Fatalf("Cannot resolve remote address: %s\n", fmt.Sprintf("%s:%s", config.RemoteAddress, "0"))
	// }

	fmt.Printf("Local Port = %s\n", config.LocalPort)
	socket, err := net.ListenUDP("udp4", lAddr)
	if err != nil {
		errorLog.Fatalf("Cannot init listen UDP-connection: %s", err)
	}
	defer socket.Close()
	infoLog.Printf("Listening local address: %s", fmt.Sprintf("%s:%s", "0.0.0.0", config.LocalPort))

	for {
		var buf [1024]byte
		n, _, err := socket.ReadFromUDP(buf[:])
		if err != nil {
			errorLog.Printf("Error reading data: %s", err)
			continue
		}
		msgLogo := string(buf[:4])
		msgType, err := networkOperation.ParseMsgType(buf[4])
		if err != nil {
			errorLog.Fatalf("Error parsing message type: %s", err)
		}

		fmt.Printf(">>> Readed %4d bytes: [ 0x%x ],  ", n, buf[:5])
		switch msgType {
		case "META":
			{
				// Сообщение с метаданными
				dataCompression := buf[5]
				SessionID, err := binary.ReadVarint(bytes.NewBuffer(buf[6:16]))
				if err != nil {
					errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[6:16], err)
				}
				filenameLength, err := binary.ReadVarint(bytes.NewBuffer(buf[16:26]))
				if err != nil {
					errorLog.Printf("Error convert bytes to filename length: 0x%x : %s", buf[16:26], err)
				}
				compressedFilenameSize, err := binary.ReadVarint(bytes.NewBuffer(buf[26:36]))
				if err != nil {
					errorLog.Printf("Error convert bytes to compressed file size: 0x%x : %s", buf[26:36], err)
				}
				fileName := string(buf[n-2-int(filenameLength) : n-2])
				fmt.Printf("META>>> %s Size=%4d bytes SSID=%4d ZIP=0x%x filename=%s fileSize=%d bytes\n", msgLogo, n, SessionID, dataCompression, fileName, compressedFilenameSize)
			}
		case "DATA":
			{
				// Сообщение с данными файла
				SessionID, err := binary.ReadVarint(bytes.NewBuffer(buf[5:15]))
				if err != nil {
					errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[5:15], err)
				}
				ChankID, err := binary.ReadVarint(bytes.NewBuffer(buf[15:25]))
				if err != nil {
					errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[15:25], err)
				}
				ChankSize, err := binary.ReadVarint(bytes.NewBuffer(buf[25:35]))
				if err != nil {
					errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[25:35], err)
				}
				data := buf[35 : ChankSize+35]
				err = fileOperation.WriteFileToDisk(fmt.Sprintf("test/Receiver/downloaded/%d.txt.zip", SessionID), data)
				if err != nil {
					errorLog.Printf("Error writing file to disk: %s", err)
				}
				fmt.Printf("DATA>>> %s Size=%4d bytes SSID=%4d ChankID=%4d ChankSize=%4d bytes\n", msgLogo, n, SessionID, ChankID, ChankSize)
			}
		case "CMD":
			{
				// Управляющее сообщение
				SessionID, err := binary.ReadVarint(bytes.NewBuffer(buf[5:15]))
				if err != nil {
					errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[5:15], err)
				}
				command := buf[16]
				fmt.Printf("CMD >>> %s Size=%4d bytes SSID=%4d Command=0x%4x\n", msgLogo, n, SessionID, command)
			}
		default:
			{
				fmt.Printf("????>>> %s\t\n", msgLogo)
			}
		}

		// time.Sleep(500 * time.Millisecond)
	}

	// infoLog.Println("END PROGRAMM")

}
