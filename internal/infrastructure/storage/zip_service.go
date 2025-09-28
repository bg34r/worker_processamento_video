package storage

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

type ZipService interface {
	CreateZipFile(files []string, zipPath string) error
}

type zipService struct{}

func NewZipService() ZipService {
	return &zipService{}
}

func (z *zipService) CreateZipFile(files []string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		err := z.addFileToZip(zipWriter, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (z *zipService) addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
