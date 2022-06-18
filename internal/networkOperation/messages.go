package networkOperation

type Msg interface {
	Init() error
	PayLoad() error
	Send() error
}

type ConnectionParams struct {
	DestIP    string
	DestPort  string
	SessionID uint32
}

type ControlMsg struct {
	ConnectionParams
	Command string
	Data    []byte
}

type MetadataMsg struct {
	ConnectionParams
	Compressing        byte
	FileName           string
	FileNameLength     uint32
	FileSize           uint32
	CompressedFileSize uint32
	FileMD5            [16]byte
	CompressedFileMD5  [16]byte
}

type DataMsg struct {
	ConnectionParams
	ChankNumber uint32
	ChankSize   uint32
	Data        []byte
}

/* Задаем настройки подключения
 */
func (cm *ControlMsg) Init(DestIP, DestPort string, SessionID uint32) error {
	return nil
}

/* Задание полезное загрузки
 */
func (cm *ControlMsg) PayLoad() error {
	return nil
}

/* Создание подключения и передача данных
 */
func (cm *ControlMsg) Send() error {
	return nil
}

/* Задаем настройки подключения
 */
func (mm *MetadataMsg) Init(DestIP, DestPort string, SessionID uint32) error {
	return nil
}

/* Задание полезное загрузки
 */
func (mm *MetadataMsg) PayLoad() error {
	return nil
}

/* Создание подключения и передача данных
 */
func (mm *MetadataMsg) Send() error {
	return nil
}

/* Задаем настройки подключения
 */
func (dm *DataMsg) Init(DestIP, DestPort string, SessionID uint32) error {
	return nil
}

/* Задание полезное загрузки
 */
func (dm *DataMsg) PayLoad() error {
	return nil
}

/* Создание подключения и передача данных
 */
func (dm *DataMsg) Send() error {
	return nil
}
