package fileOperation

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

type MetaFile struct {
	Name               string
	NameLength         uint32
	NameBASE64         string
	NameBASE64Length   uint32
	FileSize           uint32
	CompressedFileSize uint32
	Data               []byte
	CompressedData     []byte
	FileMD5            [16]byte
	CompressedFileMD5  [16]byte
}

/* Инициализация файла
Открытие файла и заполнение всхе полей структуры.
Если файл неполучается открыть - возвращает ошибку.
*/
func (mf *MetaFile) Init(filePath string) error {
	mf.Name = filePath
	mf.NameLength = uint32(len(mf.Name))

	mf.encodeBASE64()

	if err := mf.readData(); err != nil {
		return fmt.Errorf("error reading file: %s : %s", mf.Name, err)
	}

	if err := mf.compressData(); err != nil {
		return fmt.Errorf("error compressing file: %s : %s", mf.Name, err)
	}

	mf.calcMD5()

	return nil
}

/* Расчет контрольной суммы
 */
func (mf *MetaFile) calcMD5() {
	mf.FileMD5 = md5.Sum(mf.Data)
	mf.CompressedFileMD5 = md5.Sum(mf.CompressedData)
}

/* Кодирование в BASE64
 */
func (mf *MetaFile) encodeBASE64() {
	mf.NameBASE64 = base64.StdEncoding.EncodeToString([]byte(mf.Name))
	mf.NameBASE64Length = uint32(len(mf.NameBASE64))
}

/* Чтение файла
 */
func (mf *MetaFile) readData() error {
	chank := make([]byte, 1024*4)

	file, err := os.Open(mf.Name)
	if err != nil {
		return fmt.Errorf("error openning file: %s", err)
	}
	defer file.Close()

	mf.FileSize = 0
	for {
		n, err := file.Read(chank)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error reading data from file: %s", err)
			}
			break
		}
		mf.FileSize += uint32(n)
		mf.Data = append(mf.Data, chank[:n]...)
	}

	return nil
}

/* Сжатие файла
 */
func (mf *MetaFile) compressData() error {
	var bb bytes.Buffer
	w, err := gzip.NewWriterLevel(&bb, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("error creating gzip comressor: %s", err)
	}
	w.Write(mf.Data)
	w.Close()

	chank := make([]byte, 1024*4)
	mf.CompressedFileSize = 0
	for {
		n, err := bb.Read(chank)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error reading compressed data: %s", err)
			}
			break
		}
		mf.CompressedFileSize += uint32(n)
		mf.CompressedData = append(mf.CompressedData, chank[:n]...)
	}

	return nil
}

/* Форматированный вывод содержимого структуры MetaFile
 */
func (mf *MetaFile) PrettyOut() {
	text := " =======================================================\n"
	text += fmt.Sprintf("             Filename: %s [%d symbols]\n", mf.Name, mf.NameLength)
	text += fmt.Sprintf("               BASE64: %s [%d symbols]\n", mf.NameBASE64, mf.NameBASE64Length)
	text += fmt.Sprintf("            File size: %d bytes\n", mf.FileSize)
	text += fmt.Sprintf("Commpressed file size: %d bytes\n", mf.CompressedFileSize)
	text += fmt.Sprintf("             File MD5: %x\n", mf.FileMD5[:])
	text += fmt.Sprintf("  Compressed file MD5: %x\n", mf.CompressedFileMD5[:])
	text += " =======================================================\n"
	fmt.Println(text)
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
func FilterFiles(files []string, allowedTypes string) (filesAccept, filesReject []string) {
	for _, file := range files {
		flag := false
		for _, alloallowedType := range strings.Split(allowedTypes, ",") {
			if strings.EqualFold(alloallowedType, file[len(file)-len(alloallowedType):]) {
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
