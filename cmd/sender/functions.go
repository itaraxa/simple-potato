package main

import (
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

/* Функция для поиска новых файлов
 */
func searchNewFiles(infoLog *log.Logger, errorLog *log.Logger, config *senderConfig) ([]string, error) {

	// fileNames, err := fileOperation.ScanDir(config.DirectoryForNewFiles)
	fileNames, err := fileOperation.ScanDir(".")
	if err != nil {
		errorLog.Println("scanning directory error: ", err)
		return nil, err
	}

	infoLog.Print("Filter allowed file types")
	fileNames, _ = fileOperation.FilterFiles(fileNames, config.AllowedFileTypes)

	return fileNames, nil
}

/* Функция проверки доступности файла - проверяет что файл уже полностью загружен на диск
 */
func checkFile(infoLog *log.Logger, errorLog *log.Logger, config *senderConfig, file string, SessionID uint32) error {

	return nil
}

/* Функция для отправки файла
 */
func sendFile(infoLog *log.Logger, errorLog *log.Logger, config *senderConfig, conn *net.UDPConn, file string, SessionID uint32) error {
	s := session.NewSession(SessionID)
	err := s.ReadFile(file)
	if err != nil {
		errorLog.Printf("error read file: %s : %s\n", file, err)
		return fmt.Errorf("error read file: %s : %s", file, err)
	}
	err = s.SendFile(conn)
	if err != nil {
		errorLog.Printf("error send file: %s : %s\n", file, err)
		return fmt.Errorf("error send file: %s : %s", file, err)
	}

	time.Sleep(1000 * time.Millisecond)

	return nil
}

/* Функция для перемещния отправленных файлов
 */
func moveFile(infoLog *log.Logger, errorLog *log.Logger, config *senderConfig, file string, SessionID uint32) error {
	sep := string(os.PathSeparator)
	path := strings.Split(file, sep)
	fileName := path[len(path)-1]
	fullDstDir := config.DirectoryForUploadedFiles + sep + strings.Join(path[1:len(path)-1], sep)
	infoLog.Printf("Try move file %s -> %s\n", file, fullDstDir)

	if _, err := os.Stat(fullDstDir); os.IsNotExist(err) {
		err2 := os.MkdirAll(fullDstDir, 0760)
		if err2 != nil {
			errorLog.Printf("cannot create directory: %s\n", err2)
			return fmt.Errorf("cannot create directory: %s", err2)
		}
	}

	source, err := os.Open(file)
	if err != nil {
		errorLog.Printf("cannot open src file: %s: %s\n", file, err)
		return fmt.Errorf("cannot open src file: %s: %s", file, err)
	}
	defer source.Close()

	dest, err := os.Create(fullDstDir + sep + fileName)
	if err != nil {
		errorLog.Printf("cannot open dst file: %s: %s\n", fullDstDir+sep+fileName, err)
		return fmt.Errorf("cannot open dst file: %s: %s", fullDstDir+sep+fileName, err)
	}
	defer dest.Close()

	buf := make([]byte, 4*1024)

	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			errorLog.Printf("error reading file: %s : %s\n", file, err)
			return fmt.Errorf("error reading file: %s : %s", file, err)
		}
		if n == 0 {
			break
		}
		if _, err := dest.Write(buf[:n]); err != nil {
			errorLog.Printf("error writing file: %s : %s\n", fullDstDir+sep+fileName, err)
			return fmt.Errorf("error writing file: %s : %s", fullDstDir+sep+fileName, err)
		}
	}
	source.Close()
	infoLog.Printf("File %s copied to %s\n", file, fullDstDir+sep+fileName)

	err = os.Remove(file)
	if err != nil {
		errorLog.Printf("error removing file: %s : %s\n", file, err)
		return fmt.Errorf("error removing file: %s : %s", file, err)
	}
	infoLog.Printf("File %s was removed\n", file)

	return nil
}
