package networkOperation

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

/* Функция отправки файла по сети, используя udp
 */
func SendFile(fileName string, destIP string, destPort string) (err error) {
	target := fmt.Sprintf("%s:%s", destIP, destPort)
	conn, err := net.Dial("udp4", target)
	if err != nil {
		return fmt.Errorf("error init connection to %s:%s : %s", destIP, destPort, err)
	}
	defer conn.Close()

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error open file %s : %s", fileName, err)
	}
	defer file.Close()

	fileData := make([]byte, 0, 1024*100)
	fileSize := 0
	chank := make([]byte, 1024*4)
	for {
		readTotal, err := file.Read(chank)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error reading data from file: %s", err)
			}
			break
		}
		fileSize += readTotal
		fileData = append(fileData, chank[:readTotal]...)
	}
	// fileSize, err := file.Read(fileData)
	// fmt.Printf("fileSize (%s) = %d bytes\nFile data: %s", fileName, fileSize, string(fileData))
	// if err != nil {
	// 	return fmt.Errorf("error readinf file %s : %s", fileName, err)
	// }
	// buf := make([]byte, 1024)

	var header [512]byte
	// Задаем тип сообщения
	header[0] = 1

	// Записываем имя файла
	// дополняем 0 до 256
	for i := 0; i < 256; i++ {
		if i < len(fileName) {
			header[i+8] = []byte(fileName)[i]
		} else {
			header[i+8] = 0
		}
	}

	// Определяем и записываем размер файла
	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(buf, int64(fileSize))
	for j := 0; j < binary.MaxVarintLen64; j++ {
		header[j+264] = buf[j]
	}

	// Записываем md5-хэш сумму файла
	fileMD5 := md5.Sum(fileData)
	for k := 0; k < 16; k++ {
		header[k+296] = fileMD5[k]
	}

	if _, err := conn.Write(header[:]); err != nil {
		return fmt.Errorf("error writing header: %s", err)
	}

	if _, err := conn.Write(fileData); err != nil {
		return fmt.Errorf("error writing data: %s", err)
	}

	return nil
}

/* Функция получения файла по сети
 */
func ReceiveFile(listenPort string) (err error) {
	buf := make([]byte, 1024)

	conn, err := net.Dial("udp4", fmt.Sprintf(":%s", listenPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
	}

	return
}
