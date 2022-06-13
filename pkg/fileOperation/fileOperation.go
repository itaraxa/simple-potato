package fileOperation

import "fmt"

type MyFile struct {
	Name string
	Size uint32
	Data []byte
	CRC  string
}

/* Инициализация файла
Открытие файла и заполнение всхе полей структуры.
Если файл неполучается открыть - возвращает ошибку.
*/
func (f *MyFile) Init(filePath string) error {
	f.Name = filePath
	f.Size = 10
	f.Data = []byte{1, 2, 3, 4}
	return nil
}

/* Расчет контрольной суммы
 */
func (f *MyFile) CalcCRC() (md5summ string, err error) {
	var data []byte
	data = append(data, []byte(f.Name)...)
	data = append(data, []byte(string(f.Size))...)
	data = append(data, f.Data...)

	f.CRC = string(data)
	fmt.Printf("Source DATA for CRC calculation is %v", data)
	return md5summ, nil
}
