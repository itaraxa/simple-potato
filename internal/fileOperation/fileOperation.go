package fileOperation

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
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

/* Функция для рекурсивного поиска файлов в директории
 */
func ScanDir(dirPath string) ([]string, error) {
	tempFileList := make([]string, 0, 10)

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return tempFileList, fmt.Errorf("cannot scan dir: %s : %s", dirPath, err)
	}

	for _, file := range files {
		if file.IsDir() {
			fileList, err := ScanDir(dirPath + string(os.PathSeparator) + file.Name())
			if err != nil {
				return fileList, err
			}
			tempFileList = append(tempFileList, fileList...)
		} else {
			tempFileList = append(tempFileList, dirPath+string(os.PathSeparator)+file.Name())
		}
	}

	return tempFileList, nil
}

/* Функция фильтр файлов по расширению
 */
func FilterFiles(files []string) (filesAccept, filesReject []string) {
	allowedTypes := strings.Split("txt|pdf", "|")

	for _, file := range files {
		flag := false
		for _, alloallowedType := range allowedTypes {
			if file[len(file)-len(alloallowedType):] == alloallowedType {
				flag = true
				break
			}
		}
		if flag {
			filesAccept = append(filesAccept, file)
			// fmt.Println("Accepted: ", file)
		} else {
			filesReject = append(filesReject, file)
			// fmt.Println("Rejected: ", file)
		}
	}
	return
}

/* Функция для удаления префикса/"временной директории" из имен файла
 */
func PathCleaner(fileNames []string, prefix string) error {
	for j, fileName := range fileNames {
		fileNames[j] = fileName[len(prefix)+1:]
	}
	return nil
}

/* Функция для перемещения файла из временной директории в директорию с загруженными и переданными файлами
 */
func MoveFile(sourceFile string, destFile string) (err error) {
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	// Создает целевую директорию для сохранения файла
	err = os.MkdirAll(strings.Join(strings.Split(destFile, string(os.PathSeparator))[:strings.Count(destFile, string(os.PathSeparator))], string(os.PathSeparator)), 0760)
	if err != nil {
		return fmt.Errorf("cannot create destination directory for %s: %s", destFile, err)
	}

	dest, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	buf := make([]byte, 1024)

	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := dest.Write(buf[:n]); err != nil {
			return err
		}
	}
	// Delete src file
	source.Close()
	err = os.Remove(sourceFile)
	if err != nil {
		return fmt.Errorf("cannot remove file %s: %s", sourceFile, err)
	}
	return nil
}
