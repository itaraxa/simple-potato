package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/itaraxa/simple-potato/internal/session"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Настройка логирования
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.WithTime(time.Now()).Info("Receiver started")

	// Protect from Panic
	defer func() {
		if err := recover(); err != nil {
			log.WithFields(log.Fields{"Error": err}).Fatal("Fatal error -> Receiver closed")
		}
	}()

	// Read configuration file
	configFile := flag.String("config", "receiver.json", "Configuration file for receiver")
	flag.Parse()

	config := new(receiverConfig)
	err := config.loadConfig(*configFile)
	if err != nil {
		log.WithFields(log.Fields{"Error": err,
			"Configuration file": *configFile}).Fatal("Error reading configuration file")
	}
	err = config.checkConfig()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Error in configuration file")
	}
	log.WithFields(log.Fields{"Configuration file": *configFile,
		"LPort":                    config.LocalPort,
		"LAddress":                 config.LocalAddress,
		"RAddress":                 config.RemoteAddress,
		"Temp files directory":     config.DirectoryForTemporaryFiles,
		"Download files directory": config.DirectoryForDownloadedFiles}).Info("Configuration file was openned")

	// Смена рабочей директории
	err = os.Chdir(config.DirectoryForDownloadedFiles)
	oldWorkDIr, _ := os.Getwd()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Error changing current work directory")
	}
	newWorkDir, _ := os.Getwd()
	log.WithFields(log.Fields{"Old work directory": oldWorkDIr,
		"New work directory": newWorkDir}).Info("Workdir changed")

	lAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", "0.0.0.0", config.LocalPort))
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "LAddress": config.LocalAddress}).Fatal("Cannot resolve local address")
	}

	socket, err := net.ListenUDP("udp4", lAddr)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "LPort": config.LocalPort}).Fatal("Cannot init listening UDP")
	}
	defer socket.Close()
	log.WithFields(log.Fields{"LAddress": config.LocalAddress,
		"LPort": config.LocalPort}).Info("Ready to listen")

	Sessions := make(map[uint32]*session.Session)
	var SessionsMutex sync.Mutex
	// a := new(sync.Map)

STOPMAINLOOP:
	for {
		var buf [1024]byte
		_, _, err := socket.ReadFromUDP(buf[:])
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Error reading data")
			continue
		}

		// Пропускаем пакет, если он не наш
		if string(buf[:4]) != "RASU" {
			log.WithFields(log.Fields{"Error": err, "Package header": string(buf[:4])}).Error("Get strange pocket")
			continue
		}

		// Получаем ID сессии
		temp, err := binary.ReadVarint(bytes.NewBuffer(buf[4:14]))
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Error getting session ID")
		}
		SessionID := uint32(temp)
		// infoLog.Printf("get sessionID from pocket: %d", SessionID)

		// Проверяем тип пакета
		msgType := int(buf[14])

		// DEBUG
		// fmt.Printf("\n>>> SessionID: %d\n", SessionID)
		// fmt.Printf(">>> MSG TYPE: %d\n", msgType)

		switch msgType {
		case 1:
			{
				// Обработка пакета с метаданными
				// Обновление объекта с хранящейся сессией

				// ZipMode := buf[15]
				fileNameLength, err := binary.ReadVarint(bytes.NewBuffer(buf[16:26]))
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Error parse metadata pocket")
				}
				fileSize, err := binary.ReadVarint(bytes.NewBuffer(buf[26:36]))
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Error parse metadata pocket")
				}
				zipFileSize, err := binary.ReadVarint(bytes.NewBuffer(buf[36:46]))
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Error parse metadata pocket")
				}
				fileMd5 := buf[46:62]
				zipFileMd5 := buf[62:78]
				fileName := string(buf[78 : 78+fileNameLength])

				SessionsMutex.Lock()
				err = Sessions[SessionID].AddMetaData(fileName, uint32(fileSize), uint32(zipFileSize), fileMd5, zipFileMd5)
				SessionsMutex.Unlock()

				log.WithFields(log.Fields{"SessionID": SessionID,
					"File name": fileName,
					"MD5 summ":  fmt.Sprintf("0x%x", fileMd5)}).Info("Get metadata pocket")

				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Error write metadata")
				}

				// DEBUG
				// fmt.Printf("| ID: %4d | %4d | %4d bytes | %4d bytes | 0x%x | 0x%x | %s |\n", SessionID, msgType, fileSize, zipFileSize, fileMd5, zipFileMd5, fileName)
				// fmt.Printf(">>> FilenameLength: %d\n", fileNameLength)
				// fmt.Printf(">>> FileSize: %d\n", fileSize)
				// fmt.Printf(">>> ZipFileSize: %d\n", zipFileSize)
				// fmt.Printf(">>> md5: 0x%x, 0x%x\n", fileMd5, zipFileMd5)
				// fmt.Printf(">>> Filename: %s (0x%x)\n", fileName, buf[78:78+fileNameLength])
			}
		case 2:
			{
				// Обработка пакета с данными
				chankID, err := binary.ReadVarint(bytes.NewBuffer(buf[15:25]))
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Error parse data pocket")
				}
				chankSize, err := binary.ReadVarint(bytes.NewBuffer(buf[25:35]))
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Error parse data pocket")
				}
				data := buf[35 : 35+chankSize]

				SessionsMutex.Lock()
				_ = Sessions[SessionID].AddData(uint32(chankID), uint32(chankSize), data)
				SessionsMutex.Unlock()

				// DEBUG
				// infoLog.Printf("Get data for %d session\n", SessionID)
				// fmt.Printf("| ID: %4d | %4d | %4d | %4d bytes | %4d bytes |\n", SessionID, msgType, chankID, chankSize, len(data))
				// fmt.Printf(">>> ChankID: %d, ChankSize: %d bytes\n", chankID, chankSize)
				// fmt.Printf(">>> Data: 0x%x\n", data)
			}
		case 4:
			{
				// Обработка пакета с управляющим сообщением
				command := int(buf[15])
				_, err := binary.ReadVarint(bytes.NewBuffer(buf[16:26]))
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Error parse control pocket")
				}
				// data := buf[26:26+dataLength]

				switch command {
				case 1:
					{
						// Начало передачи файла
						// Создание новой сессии
						Sessions[SessionID] = session.NewSession(SessionID)
						// infoLog.Printf("Start new session: %d\n", SessionID)
					}
				case 2:
					{
						// Конец передачи файла
						go func(SessionID uint32) {
							// Откладывем сохранение данных
							time.Sleep(500 * time.Millisecond)

							SessionsMutex.Lock()
							err = Sessions[SessionID].Flash()
							if err != nil {
								log.WithFields(log.Fields{"Error": err}).Error("Error getting file")
							}
							log.WithFields(log.Fields{"File name": Sessions[SessionID].FullFileName}).Info("File saved")
							delete(Sessions, SessionID)
							SessionsMutex.Unlock()
						}(SessionID)

					}
				case 4:
					{
						// Завершение работы sender
						break STOPMAINLOOP
					}
				}

				// DEBUG
				// fmt.Printf("| ID: %4d | %4d | b%4d | %d bytes |\n", SessionID, msgType, command, dataLength)
				// fmt.Printf(">>> Command: %d\n", command)
				// fmt.Printf(">>> DataLength: %d bytes\n", dataLength)

			}
		}
	}

	log.WithTime(time.Now()).Info("Receiver closed")
}
