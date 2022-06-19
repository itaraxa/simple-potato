package session

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	POCKETSIZE uint32 = 512
)

type chank struct {
	chankID   uint32
	chankSize uint32
	data      []byte
}

type Session struct {
	ID           uint32
	fullFileName string
	fileSize     uint32
	zipFileSize  uint32
	fileMd5      []byte
	zipFileMd5   []byte
	fileName     string
	directory    string
	chanks       []chank
	data         []byte
	zipData      []byte
}

/* Создание сессии
 */
func NewSession(ID uint32) *Session {
	t := new(Session)
	t.ID = ID
	t.chanks = make([]chank, 0, 5)
	return t
}

/* Сохранение полученного файла и удаление сессии из RAM
 */
func (s *Session) Flash() error {
	return nil
}

/* Получение и запись в структуру частей файла
 */
func (s *Session) AddData(chankID, chankSize uint32, data []byte) error {
	s.chanks = append(s.chanks, chank{chankID: chankID, chankSize: chankSize, data: data})
	return nil
}

/* Получение и запись в структуру метаданных файла
 */
func (s *Session) AddMetaData(fullFileName string, fileSize, zipFileSize uint32, fileMd5, zipFileMd5 []byte) error {
	s.fullFileName = fullFileName
	s.fileSize = fileSize
	s.zipFileSize = zipFileSize
	s.fileMd5 = fileMd5
	s.zipFileMd5 = zipFileMd5

	s.directory, s.fileName, _ = splitDirFileName(fullFileName)

	return nil
}

/* Отправка файла
1. Отправка сообщения о начале передачи
2. Отправка сообщения с метаданными
3. Отправка сообщений с данными
4. Отправка сообщения о завершении передачи
*/
func (s *Session) SendFile(conn net.Conn) error {
	err := s.compressFile()
	if err != nil {
		return err
	}

	err = s.sendCommandMsg(conn, "START", []byte{})
	if err != nil {
		return err
	}

	err = s.sendMetadataMsg(conn)
	if err != nil {
		return err
	}

	// err = s.sendDataMsg(conn)
	// if err != nil {
	// 	return err
	// }

	err = s.sendCommandMsg(conn, "STOP", []byte{})
	if err != nil {
		return err
	}

	return nil
}

/* Чтение файла и запись данных в структуру
 */
func (s *Session) ReadFile(fileName string) error {
	buf := make([]byte, 4*1024)

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error openning file: %s", err)
	}
	defer file.Close()

	fileSize := uint32(0)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error reading data from file: %s : %s", fileName, err)
			}
			break
		}
		fileSize += uint32(n)
		s.data = append(s.data, buf[:n]...)
	}
	s.fullFileName = fileName
	s.fileSize = fileSize

	return nil
}

/* Сжатие данных перед передачей
 */
func (s *Session) compressFile() error {
	buf := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("error creating gzip comressor: %s", err)
	}
	_, err = w.Write(s.data)
	if err != nil {
		return fmt.Errorf("error compressing data: %s", err)
	}
	w.Close()

	s.zipFileSize = 0
	buf2 := make([]byte, 1024*4)
	for {
		k, err := buf.Read(buf2)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error reading compressed data: %s", err)
			}
			break
		}
		s.zipFileSize += uint32(k)
		s.zipData = append(s.zipData, buf2[:k]...)
	}

	// if uint32(n) != s.zipFileSize {
	// 	return fmt.Errorf("file sizes not equal: %d != %d", uint32(n), s.zipFileSize)
	// }
	return nil
}

/* Рзаделение полного пути на директорию и имя файла
 */
func splitDirFileName(fullPath string) (dirName string, fileName string, err error) {
	sep := string(os.PathSeparator)
	t := strings.Split(fullPath, sep)

	dirName = strings.Join(t[:strings.Count(fullPath, sep)], sep)
	fileName = t[strings.Count(fullPath, sep)]

	return dirName, fileName, nil
}

/* Отправка командного сообщения
 */
func (s *Session) sendCommandMsg(con net.Conn, command string, data []byte) (err error) {
	msg := make([]byte, 0, 64)
	msg = append(msg, s.createMsgHeader()...)

	// Указание типа сообщения
	msg = append(msg, byte(4))

	switch command {
	case "START":
		{
			// Команда 1, блок данны []
			msg = append(msg, s.createCommandMsg(1, []byte{})...)
		}
	case "STOP":
		{
			// Команда 2, блок данных []
			msg = append(msg, s.createCommandMsg(2, []byte{})...)
		}
	case "SENDER_STOP":
		{
			// Команд 4, блок данных b"Sender stopped"
			data := []byte("Sender stopped")
			msg = append(msg, s.createCommandMsg(4, data)...)
		}
	default:
		{
			return fmt.Errorf("unknown command: %s", command)
		}
	}
	n, err := con.Write(msg)
	if err != nil {
		return fmt.Errorf("error sending data: send %d/%d bytes : %s", n, len(msg), err)
	}

	// DEBUG
	fmt.Printf(">>> CMD_MSG: %d\n0x%x\n", s.ID, msg)

	return nil
}

/* Отправка сообщения с метаданными
 */
func (s *Session) sendMetadataMsg(con net.Conn) error {
	msg := make([]byte, 0, 64)
	msg = append(msg, s.createMsgHeader()...)

	// указываем тип сообщения
	msg = append(msg, byte(1))

	// указывавем степень сжатия
	msg = append(msg, byte(9))

	// указываем длину имени файла
	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(buf, int64(len(s.fullFileName)))
	msg = append(msg, buf...)

	// указываем размер файла
	_ = binary.PutVarint(buf, int64(s.fileSize))
	msg = append(msg, buf...)

	// указываем размер сжатого файла
	_ = binary.PutVarint(buf, int64(s.zipFileSize))
	msg = append(msg, buf...)

	// записываем md5 суммы
	msg = append(msg, s.fileMd5...)
	msg = append(msg, s.zipFileMd5...)

	// записываем имя файла
	msg = append(msg, []byte(s.fullFileName)...)

	msg = append(msg, []byte("\r\n")...)

	n, err := con.Write(msg)
	if err != nil {
		return fmt.Errorf("error sending data: send %d/%d bytes : %s", n, len(msg), err)
	}

	// DEBUG
	fmt.Printf(">>> filename: %s", s.fullFileName)
	fmt.Printf(">>> MTD_MSG: %d\n0x%x\n", s.ID, msg)

	return nil
}

// func (s *Session) sendDataMsg(con net.Conn) error {
// 	for chankId := 0; chankId <= int(s.zipFileSize)/int(POCKETSIZE); chankId++ {
// 		msg := make([]byte, 0, 64)
// 		msg = append(msg, s.createMsgHeader()...)

// 		// указываем тип сообщения
// 		msg = append(msg, byte(2))

// 		// указываем номер chankId
// 		buf := make([]byte, binary.MaxVarintLen64)
// 		_ = binary.PutVarint(buf, int64(chankId))
// 		msg = append(msg, buf...)

// 		// вычисляем и указываем размер chank
// 		a := uint32(chankId) * POCKETSIZE
// 		b := (uint32(chankId) + 1) * POCKETSIZE
// 		if b > s.zipFileSize {
// 			b = s.zipFileSize
// 		}
// 		chankSize := b - a
// 		_ = binary.PutVarint(buf, int64(chankSize))
// 		msg = append(msg, buf...)

// 		// Записываем данные в сообщение
// 		msg = append(msg, s.zipData[a:b]...)

// 		msg = append(msg, []byte("RASU")...)

// 		n, err := con.Write(msg)
// 		if err != nil {
// 			return fmt.Errorf("error sending data: send %d/%d bytes : %s", n, len(msg), err)
// 		}
// 	}

// 	return nil
// }

func (s *Session) createMsgHeader() []byte {
	msgHeader := make([]byte, 0, 64)
	msgHeader = append(msgHeader, []byte("RASU")...)

	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(buf, int64(s.ID))
	msgHeader = append(msgHeader, buf...)

	return msgHeader
}

func (s *Session) createCommandMsg(command byte, data []byte) []byte {
	msgBody := make([]byte, 0, 64)
	msgBody = append(msgBody, byte(1))
	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(buf, int64(len(data)))
	msgBody = append(msgBody, buf...)
	msgBody = append(msgBody, data...)
	msgBody = append(msgBody, []byte("\r\n")...)

	return msgBody
}
