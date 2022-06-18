package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type senderConfig struct {
	SendToPort                 string `json:"Send to Port"`
	SendToAddress              string `json:"Send to Address"`
	SendFromAddress            string `json:"Send from Address"`
	DirectoryForTemporaryFiles string `json:"Directory for temporary files"`
	DirectoryForUploadedFiles  string `json:"Directory for uploaded files"`
	CycleTimeMs                string `json:"Cycle time (ms)"`
	Compression                string `json:"Compression"`
	AllowedFileTypes           string `json:"Allowed file types"`
}

/* Чтение параметров конфигурации их файла
 */
func (c *senderConfig) loadConfig(configFileName string) (err error) {
	inFile, err := os.Open(configFileName)
	if err != nil {
		return fmt.Errorf("error opening %v file: %v", configFileName, err)
	}
	defer inFile.Close()

	decoder := json.NewDecoder(inFile)
	err = decoder.Decode(c)
	if err != nil {
		return fmt.Errorf("error parsing %v file: %v", configFileName, err)
	}

	return nil
}

/* Проверка параметров, прочитанных из конфигурационного файла
 */
func (c *senderConfig) checkConfig() (err error) {
	return nil
}
