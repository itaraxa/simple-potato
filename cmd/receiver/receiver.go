package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/itaraxa/simple-potato/internal/session"
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

	//var Sessions map[uint32]*session.Session
	Sessions := make(map[uint32]*session.Session)

	fmt.Print("|    ID    |  T  |    SIZE    |      bytes |       |       |          |\n")

	for {
		var buf [1024]byte
		_, _, err := socket.ReadFromUDP(buf[:])
		if err != nil {
			errorLog.Printf("Error reading data: %s", err)
			continue
		}

		// Пропускаем пакет, если он не наш
		if string(buf[:4]) != "RASU" {
			continue
		}

		// Получаем ID сессии
		temp, err := binary.ReadVarint(bytes.NewBuffer(buf[4:14]))
		if err != nil {
			errorLog.Printf("error getting Session ID: %s", err)
		}
		SessionID := uint32(temp)
		fmt.Printf(">>> SessionID: %d\n", SessionID)

		// Проверяем тип пакета
		msgType := int(buf[14])
		// fmt.Printf("Type: %d, rAddr: %v, pocketSize: %d bytes\n", msgType, rAddr.String(), n)
		fmt.Printf(">>> MSG TYPE: %d\n", msgType)

		switch msgType {
		case 1:
			{
				// Обработка пакета с метаданными
				// Обновление объекта с хранящейся сессией

				// ZipMode := buf[15]
				fileNameLength, err := binary.ReadVarint(bytes.NewBuffer(buf[16:26]))
				if err != nil {
					errorLog.Printf("Error parse metadata pocket: %s", err)
				}
				fileSize, err := binary.ReadVarint(bytes.NewBuffer(buf[26:36]))
				if err != nil {
					errorLog.Printf("Error parse metadata pocket: %s", err)
				}
				zipFileSize, err := binary.ReadVarint(bytes.NewBuffer(buf[36:46]))
				if err != nil {
					errorLog.Printf("Error parse metadata pocket: %s", err)
				}
				fileMd5 := buf[46:62]
				zipFileMd5 := buf[62:78]
				fileName := string(buf[78 : 78+fileNameLength])

				err = Sessions[SessionID].AddMetaData(fileName, uint32(fileSize), uint32(zipFileSize), fileMd5, zipFileMd5)
				if err != nil {
					errorLog.Printf("Error write metadata: %s", err)
				}

				fmt.Printf("| ID: %4d | %4d | %4d bytes | %4d bytes | 0x%x | 0x%x | %s |\n", SessionID, msgType, fileSize, zipFileSize, fileMd5, zipFileMd5, fileName)

				// DEBUG
				fmt.Printf(">>> FilenameLength: %d\n", fileNameLength)
				fmt.Printf(">>> FileSize: %d\n", fileSize)
				fmt.Printf(">>> ZipFileSize: %d\n", zipFileSize)
				fmt.Printf(">>> md5: 0x%x, 0x%x\n", fileMd5, zipFileMd5)
				fmt.Printf(">>> Filename: %s (0x%x)\n", fileName, buf[78:78+fileNameLength])
			}
		case 2:
			{
				// Обработка пакета с данными
				chankID, err := binary.ReadVarint(bytes.NewBuffer(buf[15:25]))
				if err != nil {
					errorLog.Printf("Error parse data pocket: %s", err)
				}
				chankSize, err := binary.ReadVarint(bytes.NewBuffer(buf[25:35]))
				if err != nil {
					errorLog.Printf("Error parse data pocket: %s", err)
				}
				data := buf[35 : 35+chankSize]

				_ = Sessions[SessionID].AddData(uint32(chankID), uint32(chankSize), data)

				fmt.Printf("| ID: %4d | %4d | %4d | %4d bytes | %4d bytes |\n", SessionID, msgType, chankID, chankSize, len(data))
			}
		case 4:
			{
				// Обработка пакета с управляющим сообщением
				command := int(buf[15])
				dataLength, err := binary.ReadVarint(bytes.NewBuffer(buf[16:26]))
				if err != nil {
					errorLog.Printf("Error parse control pocket: %s", err)
				}
				// data := buf[26:26+dataLength]

				switch command {
				case 1:
					{
						// Начало передачачи файла
						// Создание новой сессии
						Sessions[SessionID] = session.NewSession(SessionID)
					}
				case 2:
					{
						// Конец передачи файла
						err = Sessions[SessionID].Flash()
						if err != nil {
							errorLog.Printf("Error getting file: %s", err)
						}
					}
				case 4:
					{
						// Завершение работы sender
						continue
					}
				}

				fmt.Printf("| ID: %4d | %4d | b%4d | %d bytes |\n", SessionID, msgType, command, dataLength)
			}
		}

		// var buf [1024]byte
		// n, _, err := socket.ReadFromUDP(buf[:])
		// if err != nil {
		// 	errorLog.Printf("Error reading data: %s", err)
		// 	continue
		// }
		// msgLogo := string(buf[:4])
		// msgType, err := networkOperation.ParseMsgType(buf[4])
		// if err != nil {
		// 	errorLog.Fatalf("Error parsing message type: %s", err)
		// }

		// fmt.Printf(">>> Readed %4d bytes: [ 0x%x ],  ", n, buf[:5])
		// switch msgType {
		// case "META":
		// 	{
		// 		// Сообщение с метаданными
		// 		dataCompression := buf[5]
		// 		SessionID, err := binary.ReadVarint(bytes.NewBuffer(buf[6:16]))
		// 		if err != nil {
		// 			errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[6:16], err)
		// 		}
		// 		filenameLength, err := binary.ReadVarint(bytes.NewBuffer(buf[16:26]))
		// 		if err != nil {
		// 			errorLog.Printf("Error convert bytes to filename length: 0x%x : %s", buf[16:26], err)
		// 		}
		// 		compressedFilenameSize, err := binary.ReadVarint(bytes.NewBuffer(buf[26:36]))
		// 		if err != nil {
		// 			errorLog.Printf("Error convert bytes to compressed file size: 0x%x : %s", buf[26:36], err)
		// 		}
		// 		fileName := string(buf[n-2-int(filenameLength) : n-2])
		// 		fmt.Printf("META>>> %s Size=%4d bytes SSID=%4d ZIP=0x%x filename=%s fileSize=%d bytes\n", msgLogo, n, SessionID, dataCompression, fileName, compressedFilenameSize)
		// 	}
		// case "DATA":
		// 	{
		// 		// Сообщение с данными файла
		// 		SessionID, err := binary.ReadVarint(bytes.NewBuffer(buf[5:15]))
		// 		if err != nil {
		// 			errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[5:15], err)
		// 		}
		// 		ChankID, err := binary.ReadVarint(bytes.NewBuffer(buf[15:25]))
		// 		if err != nil {
		// 			errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[15:25], err)
		// 		}
		// 		ChankSize, err := binary.ReadVarint(bytes.NewBuffer(buf[25:35]))
		// 		if err != nil {
		// 			errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[25:35], err)
		// 		}
		// 		data := buf[35 : ChankSize+35]
		// 		err = fileOperation.WriteFileToDisk(fmt.Sprintf("test/Receiver/downloaded/%d.txt.zip", SessionID), data)
		// 		if err != nil {
		// 			errorLog.Printf("Error writing file to disk: %s", err)
		// 		}
		// 		fmt.Printf("DATA>>> %s Size=%4d bytes SSID=%4d ChankID=%4d ChankSize=%4d bytes\n", msgLogo, n, SessionID, ChankID, ChankSize)
		// 	}
		// case "CMD":
		// 	{
		// 		// Управляющее сообщение
		// 		SessionID, err := binary.ReadVarint(bytes.NewBuffer(buf[5:15]))
		// 		if err != nil {
		// 			errorLog.Printf("Error convert bytes to session id: 0x%x : %s", buf[5:15], err)
		// 		}
		// 		command := buf[16]
		// 		fmt.Printf("CMD >>> %s Size=%4d bytes SSID=%4d Command=0x%4x\n", msgLogo, n, SessionID, command)
		// 	}
		// default:
		// 	{
		// 		fmt.Printf("????>>> %s\t\n", msgLogo)
		// 	}
		// }

		// time.Sleep(500 * time.Millisecond)
	}

	// infoLog.Println("END PROGRAMM")

}
