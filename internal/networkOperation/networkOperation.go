package networkOperation

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

func kek() error {
	return nil
}

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

	buf := make([]byte, 1024)

	var header [512]byte
	// Задаем тип сообщения
	header[0] = 1

	// Записываем имя файла
	// дополняем 0 до 256
	for i := 1; i < 256; i++ {
		if i < len(fileName) {
			header[i] = []byte(fileName)[i-1]
		} else {
			header[i] = 0
		}
	}

	// Определяем и записываем размер файла
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error get info about file: %s : %s", fileName, err)
	}
	bufLittle := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(bufLittle, info.Size())
	for j := 0; j < binary.MaxVarintLen64; j++ {
		header[257+j] = bufLittle[j]
	}
	// for k := 0; k < 8; k += 8 {
	// 	fmt.Printf("%v", header[k*8:k*8+7])
	// }
	if _, err := conn.Write(header[:]); err != nil {
		return err
	}

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := conn.Write(buf[:n]); err != nil {
			return err
		}
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
