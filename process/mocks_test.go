package process

import (
	"errors"
	"io"
	"os"
)

type mockStorage struct{}

func (m mockStorage) Kind() string {
	return "mock"
}

func (m mockStorage) CollectionRef() string {
	return "mock-collection"
}

// UploadFile simulates an upload and saves the file to ./testdata/upload/{reference}.
func (m mockStorage) UploadFile(filename, reference string) error {

	switch reference {
	case "phpcompatuploaderror-phpcs_phpcompatibility-parsed.json":
		fallthrough
	case "uploaderrorchecksum-phpcs_wordpress-raw.json":
		return errors.New("upload error")
	}

	file, err := fileOpen(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	out, _ := os.Create("./testdata/upload/" + reference)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return err
	}
	return nil
}

func (m mockStorage) DownloadFile(reference, filename string) error {
	return nil
}
