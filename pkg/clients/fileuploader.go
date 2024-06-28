package clients

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type FileUploader struct {
	Url string
}

type FileContent struct {
	Filename string
	Filetype string
	Data     []byte
}

func NewFileUploader(url string) *FileUploader {
	return &FileUploader{
		Url: url,
	}
}
func (f *FileUploader) Upload(files ...FileContent) error {
	var (
		buf = new(bytes.Buffer)
		w   = multipart.NewWriter(buf)
	)

	for _, f := range files {
		part, err := w.CreateFormFile(f.Filetype, filepath.Base(f.Filename))
		if err != nil {
			return err
		}

		_, err = part.Write(f.Data)
		if err != nil {
			return err
		}
	}

	err := w.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", f.Url, buf)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
