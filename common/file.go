package common

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"io"
	"os"
)

func GetContent(c context.Context, filename string) (*[]byte, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		log.Errorf(c, "Error opening file: %v", err)
		return nil, err
	}
	log.Infof(c, "FILE FOUND : "+filename+" \n")
	buffer := make([]byte, 10*1024*1024)
	n, err := file.Read(buffer)
	if (err == nil) || (err == io.EOF) {
		content := buffer[:n]
		return &content, nil
	}

	log.Errorf(c, "Error reading file: %v", err)
	return nil, err

}
