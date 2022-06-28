package session

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"time"
)

/* Сохранение полученного файла и удаление сессии из RAM
 */
func (s *Session) Flash() error {
	err := s.combineZipData()
	if err != nil {
		return fmt.Errorf("error combinig data from chanks: %s", err)
	}

	if s.checkZipData() {
		err = s.unzipFile()
		if err != nil {
			return fmt.Errorf("error unzip data: %s", err)
		}

		// DEBUG
		// fmt.Printf(">>> Getted zip-data - OK: md5=0x%x\n", s.zipFileMd5)
	} else {
		return fmt.Errorf("incorrect getted zip-data md5 sum\n>>> Data: 0x%x", s.zipData)
	}

	// if !s.checkData() {
	// 	return fmt.Errorf("incorrect getted data md5 sum: 0x%x", s.fileMd5)
	// }

	err = s.writeDataToFile()
	if err != nil {
		return fmt.Errorf("error writing data to file: %s", err)
	}

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

	s.directory, s.fileName, _ = s.splitDirFileName()

	return nil
}

/* Отправка файла
1. Отправка сообщения о начале передачи
2. Отправка сообщения с метаданными
3. Отправка сообщений с данными
4. Отправка сообщения о завершении передачи
*/
func (s *Session) SendFile(conn net.Conn) error {
	err := s.zipFile()
	if err != nil {
		return err
	}

	err = s.sendCommandMsg(conn, "START", []byte{})
	if err != nil {
		return err
	}

	// Для того что бы receiver успел создать объект
	time.Sleep(500 * time.Microsecond)

	err = s.sendMetadataMsg(conn)
	if err != nil {
		return err
	}

	err = s.sendDataMsg(conn)
	if err != nil {
		return err
	}

	// Для того что бы этот пакет не обогнал пакет с данными
	time.Sleep(10 * time.Millisecond)

	err = s.sendCommandMsg(conn, "STOP", []byte{})
	if err != nil {
		return err
	}

	return nil
}

/* Сборка полученных chanks в []byte
TODO: добавить сортировку чанков по chankID и проверку отсутсвующих
*/
func (s *Session) combineZipData() error {
	// Сортировка чанков по chankID
	sort.Slice(s.chanks, func(i, j int) bool {
		return s.chanks[i].chankID < s.chanks[j].chankID
	})

	for k := 0; k < len(s.chanks); k++ {
		// if k != int(s.chanks[k].chankID) {
		// 	return fmt.Errorf("error in chank order")
		// }
		s.zipData = append(s.zipData, s.chanks[k].data...)
	}
	return nil
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
	// fmt.Printf(">>> CMD_MSG: %d\n0x%x\n", s.ID, msg)

	return nil
}

/* Отправка сообщения с метаданными
 */
func (s *Session) sendMetadataMsg(con net.Conn) error {
	msg := make([]byte, 0, 128)
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
	// fmt.Printf(">>> len(buf)= %d bytes, buf= 0x%x\n", len(buf), buf)
	// fmt.Printf(">>> filename: %s\n", s.fullFileName)
	// fmt.Printf(">>> MTD_MSG: %d\n0x%x\n", s.ID, msg)

	return nil
}

func (s *Session) sendDataMsg(con net.Conn) error {
	for chankId := 0; chankId <= int(s.zipFileSize)/int(POCKETSIZE); chankId++ {
		msg := make([]byte, 0, 64)
		msg = append(msg, s.createMsgHeader()...)

		// указываем тип сообщения
		msg = append(msg, byte(2))

		// указываем номер chankId
		buf := make([]byte, binary.MaxVarintLen64)
		_ = binary.PutVarint(buf, int64(chankId))
		msg = append(msg, buf...)

		// вычисляем и указываем размер chank
		a := uint32(chankId) * POCKETSIZE
		b := (uint32(chankId) + 1) * POCKETSIZE
		if b > s.zipFileSize {
			b = s.zipFileSize
		}
		chankSize := b - a
		_ = binary.PutVarint(buf, int64(chankSize))
		msg = append(msg, buf...)

		// Записываем данные в сообщение
		msg = append(msg, s.zipData[a:b]...)

		msg = append(msg, []byte("RASU")...)

		n, err := con.Write(msg)
		if err != nil {
			return fmt.Errorf("error sending data: send %d/%d bytes : %s", n, len(msg), err)
		}

		time.Sleep(100 * time.Microsecond)
	}

	return nil
}

func (s *Session) createMsgHeader() []byte {
	msgHeader := make([]byte, 0, 14)
	msgHeader = append(msgHeader, []byte("RASU")...)

	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(buf, int64(s.ID))
	msgHeader = append(msgHeader, buf...)

	return msgHeader
}

func (s *Session) createCommandMsg(command byte, data []byte) []byte {
	msgBody := make([]byte, 0, 64)
	msgBody = append(msgBody, command)
	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(buf, int64(len(data)))
	msgBody = append(msgBody, buf...)
	msgBody = append(msgBody, data...)
	msgBody = append(msgBody, []byte("\r\n")...)

	return msgBody
}
