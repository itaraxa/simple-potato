/* Описание используемой модели файла
 */

package model

//File data and metadata
type File struct {
	ID           uint32
	FullFileName string
	fileSize     uint32
	zipFileSize  uint32
	fileMD5      [16]byte
	zipFileMD5   [16]byte
	data         []byte
	zipData      []byte
}

type FileRepo interface {
	Send() error
	Receive() error
	Read() error
	Write() error
}

type SessionRepo interface {
	NewSession() error
}
