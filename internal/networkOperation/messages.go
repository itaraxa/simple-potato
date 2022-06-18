package networkOperation

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/itaraxa/simple-potato/internal/fileOperation"
)

type Msg interface {
	Init() error
	PayLoad() error
	Send() error
}

/* Коммандное сообщение
Команды:
1 - начало отправки сообщения
2 - конец отправки сообщения
3 - ?
*/
type ControlMsg struct {
	pocket []byte
}

type MetadataMsg struct {
	pocket []byte
}

type DataMsg struct {
	pockets [][]byte
}

/* Задание полезное загрузки
 */
func (cm *ControlMsg) PayLoad(SessionID uint32, command byte, data []byte) error {
	cm.pocket = append(cm.pocket, []byte("RASU")...)
	cm.pocket = append(cm.pocket, 4) // Тип сообщения "4"

	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(buf, int64(SessionID))
	cm.pocket = append(cm.pocket, buf...)
	cm.pocket = append(cm.pocket, command)
	cm.pocket = append(cm.pocket, data...)
	cm.pocket = append(cm.pocket, []byte("\r\n")...)
	return nil
}

/* Передача данных в установленное подключение
 */
func (cm *ControlMsg) Send(con net.Conn) error {
	// if _, err := con.Write(cm.pocket); err != nil {
	// 	return fmt.Errorf("error send data: %s", err)
	// }
	n, err := con.Write(cm.pocket)
	if err != nil {
		fmt.Printf(">>> Send %d/%d bytes\n", n, len(cm.pocket))
		return fmt.Errorf("error send data: %s", err)
	}
	fmt.Printf(">>> Send %d/%d bytes\n", n, len(cm.pocket))
	return nil
}

/* Задание полезное загрузки
 */
func (mm *MetadataMsg) PayLoad(SessionID uint32, compressing byte, t fileOperation.MetaFile) error {
	mm.pocket = append(mm.pocket, []byte("RASU")...)
	mm.pocket = append(mm.pocket, 1)           // Тип сообщения "1"
	mm.pocket = append(mm.pocket, compressing) // Указание сжатия
	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(buf, int64(SessionID)) // Запись ID сессии
	mm.pocket = append(mm.pocket, buf...)
	_ = binary.PutVarint(buf, int64(t.NameLength))
	mm.pocket = append(mm.pocket, buf...) // Запись длины имени файла
	_ = binary.PutVarint(buf, int64(t.CompressedFileSize))
	mm.pocket = append(mm.pocket, buf...) // Запись размера сжатого файла
	mm.pocket = append(mm.pocket, t.FileMD5[:]...)
	mm.pocket = append(mm.pocket, t.CompressedFileMD5[:]...)
	mm.pocket = append(mm.pocket, []byte(t.Name)...)
	mm.pocket = append(mm.pocket, []byte("\r\n")...)

	return nil
}

/* Создание подключения и передача данных
 */
func (mm *MetadataMsg) Send(con net.Conn) error {
	// if _, err := con.Write(mm.pocket); err != nil {
	// 	return fmt.Errorf("error send data: %s", err)
	// }
	n, err := con.Write(mm.pocket)
	if err != nil {
		fmt.Printf(">>> Send %d/%d bytes\n", n, len(mm.pocket))
		return fmt.Errorf("error send data: %s", err)
	}
	fmt.Printf(">>> Send %d/%d bytes\n", n, len(mm.pocket))
	return nil
}

/* Задание полезное загрузки
 */
func (dm *DataMsg) PayLoad(SessionID uint32, compressing byte, t fileOperation.MetaFile) error {
	pocketSize := 500
	for i := 0; i <= int(t.CompressedFileSize)/pocketSize; i++ {
		temp := make([]byte, 0, pocketSize)
		temp = append(temp, []byte("RASU")...)
		temp = append(temp, 2)
		buf := make([]byte, binary.MaxVarintLen64)
		_ = binary.PutVarint(buf, int64(SessionID))
		temp = append(temp, buf...)
		_ = binary.PutVarint(buf, int64(i))
		temp = append(temp, buf...)

		a := i
		b := i + pocketSize
		if b > int(t.CompressedFileSize) {
			b = int(t.CompressedFileSize)
		}
		chankSize := b - a
		_ = binary.PutVarint(buf, int64(chankSize))
		temp = append(temp, buf...)
		temp = append(temp, t.CompressedData[a:b]...)
		temp = append(temp, []byte("\r\n")...)

		dm.pockets = append(dm.pockets, temp)
		// fmt.Printf("Chank count = %d\n", len(dm.pockets))
	}

	return nil
}

/* Создание подключения и передача данных
 */
func (dm *DataMsg) Send(con net.Conn) error {
	for _, chank := range dm.pockets {
		// if _, err := con.Write(chank); err != nil {
		// 	return fmt.Errorf("error send data: %s", err)
		// }
		// n, err := con.Write(chank)
		// fmt.Printf(">>> Data package send: %d ERR: %s\n", n, err)

		n, err := con.Write(chank)
		if err != nil {
			fmt.Printf(">>> Send %d/%d bytes\n", n, len(chank))
			return fmt.Errorf("error send data: %s", err)
		}
		fmt.Printf(">>> Send %d/%d bytes\n", n, len(chank))
	}

	return nil
}

func ParseMsgType(input byte) (string, error) {
	if input == byte(1) {
		return "META", nil
	} else if input == byte(2) {
		return "DATA", nil
	} else if input == byte(4) {
		return "CMD", nil
	}
	return "????", fmt.Errorf("unknown message type: 0x%x", input)
}
