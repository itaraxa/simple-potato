package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type receiverConfig struct {
	DirectoryForDownloadedFiles string `json:"Directory for downloaded files"`
	DirectoryForTemporaryFiles  string `json:"Directory for temporary files"`
	RemoteAddress               string `json:"Remote Address"`
	LocalPort                   string `json:"Local Port"`
	LocalAddress                string `json:"Local Address"`
}

/* Функция для чтения параметров конфигурации из файла, с указанным именем
 */
func (c *receiverConfig) loadConfig(configFileName string) error {
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

/* Проверки параметров, указанных в конфигурационном файле
 */
func (c *receiverConfig) checkConfig() error {
	// if c.Log_Level != "DEBUG" && c.Log_Level != "INFO" {
	// 	return fmt.Errorf("bad log level in configuration file : %v (should be DEBUG or INFO)", c.Log_Level)
	// }

	// Check directoryes exist
	if _, err := os.Stat(c.DirectoryForTemporaryFiles); os.IsNotExist(err) {
		return fmt.Errorf("directory %s not exist: %s", c.DirectoryForTemporaryFiles, err)
	}
	if _, err := os.Stat(c.DirectoryForDownloadedFiles); os.IsNotExist(err) {
		return fmt.Errorf("directory %s not exist: %s", c.DirectoryForDownloadedFiles, err)
	}

	return nil
}
