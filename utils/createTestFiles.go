package main

import (
	"log"
	"os"
)

type testConfig struct {
	DirectoryForTemporaryFiles string `json:"Directory for temporary files"`
	DirectoryForUploadedFiles  string `json:"Directory for uploaded files"`
	SendToAddress              string `json:"Send to Address"`
	SendToPort                 string `json:"Send to Port"`
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	config := new(testConfig)
	config.DirectoryForTemporaryFiles = "test/tmp"

	files := make([]string, 0, 10)
	files = append(files, "test.txt")
	files = append(files, "text.tst")

	for _, file := range files {
		infoLog.Println(file)
		t, _ := os.Create(config.DirectoryForTemporaryFiles + string(os.PathSeparator) + file)
		defer t.Close()

		if _, err := t.Write([]byte("123456")); err != nil {
			infoLog.Print(err)
		}
	}
}
