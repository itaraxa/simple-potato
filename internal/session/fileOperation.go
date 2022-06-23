package session

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

/* Чтение файла и запись данных в структуру
 */
func (s *Session) ReadFile(fileName string) error {
	buf := make([]byte, 4*1024)

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error openning file: %s", err)
	}
	defer file.Close()

	var fileSize uint32 = 0
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
	t := md5.Sum(s.data)
	s.fileMd5 = t[:]

	return nil
}

/* Сжатие данных перед передачей
 */
func (s *Session) zipFile() error {
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

	t := md5.Sum(s.zipData)
	s.zipFileMd5 = t[:]

	// if uint32(n) != s.zipFileSize {
	// 	return fmt.Errorf("file sizes not equal: %d != %d", uint32(n), s.zipFileSize)
	// }
	return nil
}

/* Распаковка сжатых данных
 */
func (s *Session) unzipFile() error {
	sliceReader := bytes.NewReader(s.zipData)
	r, err := gzip.NewReader(sliceReader)
	if err != nil {
		return fmt.Errorf("cannot create gzip.Reader: %s", err)
	}
	defer r.Close()

	buf := make([]byte, 4*1024)
	fileSize := 0
	for {
		n, err := r.Read(buf)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error read data from zipData: %s", err)
			}
		}
		fileSize += n
		s.data = append(s.data, buf[:n]...)

		if n == 0 {
			break
		}
	}

	// DEBUG

	fmt.Printf(">>>> File Size afer unzip: %d bytes, from metadata: %d bytes\n", fileSize, s.fileSize)
	time.Sleep(1000 * time.Millisecond)
	return nil
}

/* Сравнение md5 суммы полученных данных и md5 полученной в метаданных
 */
func (s *Session) checkZipData() bool {
	t := md5.Sum(s.zipData)
	if bytes.Equal(s.zipFileMd5, t[:]) {
		return true
	}
	// DEBUG

	fmt.Printf(">>>Getted zip-data md5: 0x%x 0x%x\n", s.zipFileMd5, t[:])

	return false
}

/* Сравнение md5 суммы полученных данных и md5 полученной в метаданных
 */
func (s *Session) checkData() bool {
	t := md5.Sum(s.data)
	if bytes.Equal(s.fileMd5, t[:]) {
		return true
	}
	// DEBUG

	fmt.Printf(">>>Getted data md5: 0x%x 0x%x\n", s.fileMd5, t[:])

	return false
}

/* Разделение полного пути на директорию и имя файла
 */
func (s *Session) splitDirFileName() (dirName string, fileName string, err error) {
	sep := string(os.PathSeparator)
	t := strings.Split(s.fullFileName, sep)

	dirName = strings.Join(t[:strings.Count(s.fullFileName, sep)], sep)
	fileName = t[strings.Count(s.fullFileName, sep)]

	return dirName, fileName, nil
}

/* Функция записи полученных данных в файл
 */
func (s *Session) writeDataToFile() error {
	err := s.createDir()
	if err != nil {
		return fmt.Errorf("cannot create directory: %s", err)
	}

	file, err := os.Create(s.fullFileName)
	if err != nil {
		return fmt.Errorf("cannot create file: %s", err)
	}
	defer file.Close()

	if _, err = file.Write(s.data); err != nil {
		return fmt.Errorf("cannot write data to file: %s", err)
	}

	return nil
}

/* Функция создания директории
Перед созданием директории проверяется ее существование.
При необходимости создается директория s.directory
*/
func (s *Session) createDir() error {
	if _, err := os.Stat(s.directory); os.IsNotExist(err) {
		err2 := os.MkdirAll(s.directory, 0760)
		if err2 != nil {
			return fmt.Errorf("cannot create directory: %s", err2)
		}
	}
	return nil
}
