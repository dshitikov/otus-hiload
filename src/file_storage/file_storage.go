package file_storage

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

type fileStorage struct {
	storageDir string
}

type IFileStorage interface {
	SaveFile(file multipart.File, fileName string) (string, error)
	DeleteFile(fileName string)
}

func NewFileStorage(storageDir string) IFileStorage {
	return &fileStorage{storageDir: storageDir}
}

func (s *fileStorage) DeleteFile(fileName string) {
	if len(fileName) == 0 {
		return
	}

	fullPath := path.Join(s.storageDir, fileName)

	var _, err = os.Stat(fullPath)
	if err != nil {
		log.Printf("deleteFile Stat error: %s", err.Error())
		return
	}

	err = os.Remove(fullPath)
	if err != nil {
		log.Printf("deleteFile remove error: %s", err.Error())
	}
}

func (s *fileStorage) checkFileType(reader io.Reader) error {
	buff := make([]byte, 512)
	_, err := reader.Read(buff)

	if err != nil {
		return err
	}

	fileType := http.DetectContentType(buff)

	switch fileType {
	case "image/jpeg", "image/jpg":
	case "image/gif":
	case "image/png":
	default:
		return fmt.Errorf("unknown file type uploaded: %s", fileType)
	}

	return nil
}

func (s *fileStorage) SaveFile(file multipart.File, fileName string) (string, error) {
	err := s.checkFileType(file)
	if err != nil {
		log.Printf("saveFile checkFileType: %s", err.Error())
		return "", err
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Printf("saveFile Seek: %s", err.Error())
		return "", err
	}

	ext := path.Ext(fileName)
	uid := uuid.NewV4()
	fName := uid.String() + ext
	targetFileName := path.Join(s.storageDir, fName)
	log.Printf("targetFileName=%s", targetFileName)

	f, err := os.OpenFile(targetFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("saveFile OpenFile: %s", err.Error())
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		log.Printf("saveFile Copy: %s", err.Error())
		return "", err
	}

	return fName, nil
}
