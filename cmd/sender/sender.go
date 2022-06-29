/* Реализация отправщика файлов с многопоточностью
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

func main() {
	runtime.GOMAXPROCS(2)

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

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

	err = os.Chdir(config.DirectoryForNewFiles)
	if err != nil {
		errorLog.Fatalf("Cannot switch to temporary folder: %s", config.DirectoryForNewFiles)
	}

	remoteAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", config.SendToAddress, config.SendToPort))
	if err != nil {
		errorLog.Fatalf("Incorrect destination address: %s", fmt.Sprintf("%s:%s", config.SendToAddress, config.SendToPort))
	}
	localAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", config.SendFromAddress, "0"))
	if err != nil {
		errorLog.Fatalf("Incorrect local address: %s", fmt.Sprintf("%s:%s", config.SendFromAddress, "0"))
	}
	conn, err := net.DialUDP("udp4", localAddr, remoteAddr)
	if err != nil {
		errorLog.Fatalf("Error init connection: %s", err)
	}
	defer conn.Close()

	wg := new(sync.WaitGroup)
	var SessionID uint32 = 0

	// Бесконечный цикл, который запускает следующие горутины
	// 1. Таймер в соответствии с параметром CycleTimeMs конфигурации
	// 2. Функция сканер, которая ищет новые файлы. По найденным файлам запускается новый цикл, который:
	// 2.1 Проверяет доступен ли файл
	// 2.2 Отправляет файл по сети
	// 2.3 Перемещает файл в директорию с отправленными
	for {
		// 1
		wg.Add(1)
		go func(infoLog *log.Logger, errorLog *log.Logger, wg *sync.WaitGroup, config *senderConfig) {
			infoLog.Printf("Start timer for new iteration at %v\n", time.Now())
			time.Sleep(time.Duration(config.CycleTimeMs) * time.Millisecond)
			infoLog.Printf("End timer for iteration at %v\n", time.Now())
			defer wg.Done()
		}(infoLog, errorLog, wg, config)

		// 2
		files, err := searchNewFiles(infoLog, errorLog, config)
		if err != nil {
			continue
		}

		for _, file := range files {
			SessionID++
			infoLog.Printf("Start new session %d for file: %s\n", SessionID, file)
			wg.Add(1)
			relativePath := "." + file[len(config.DirectoryForNewFiles):]
			go func(infoLog *log.Logger, errorLog *log.Logger, wg *sync.WaitGroup, config *senderConfig, SessionID uint32, fileName string) {
				defer wg.Done()
				defer infoLog.Printf("Session %d was ended\n", SessionID)

				// 2.1
				err := checkFile(infoLog, errorLog, config, fileName, SessionID)
				if err != nil {
					return
				}
				infoLog.Printf("File: %s(SessionID=%d) was accepted\n", fileName, SessionID)
				// 2.2
				err = sendFile(infoLog, errorLog, config, conn, fileName, SessionID)
				if err != nil {
					return
				}
				infoLog.Printf("File: %s(SessionID=%d) was sended\n", fileName, SessionID)
				// 2.3
				err = moveFile(infoLog, errorLog, config, fileName, SessionID)
				if err != nil {
					return
				}
				infoLog.Printf("File: %s(SessionID=%d) was moved\n", fileName, SessionID)
			}(infoLog, errorLog, wg, config, SessionID, relativePath)
		}
		wg.Wait()
		// DEBUG
		// break
	}

}
