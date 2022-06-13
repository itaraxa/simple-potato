package fileOperation

import (
	"crypto/md5"
	"fmt"
)

type MyFile struct {
	Name string
	Size uint32
	Data []byte
	MD5  [16]byte
}

/* Инициализация файла
Открытие файла и заполнение всхе полей структуры.
Если файл неполучается открыть - возвращает ошибку.
*/
func (f *MyFile) Init(filePath string) error {
	f.Name = filePath
	f.Size = 10
	f.Data = []byte{71, 72, 73, 74}

	f.calcMD5()

	return nil
}

/* Расчет контрольной суммы
 */
func (f *MyFile) calcMD5() {
	f.MD5 = md5.Sum(f.Data)
}

/* Форматированный вывод содержимого структуры MyFile
 */
func (f *MyFile) PrettyOut() {
	fmt.Println("Filename: ", f.Name)
	fmt.Println("Size of file: ", f.Size, "bytes")
	fmt.Printf("md5: %x\n", f.MD5)
}
