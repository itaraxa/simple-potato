package session

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
