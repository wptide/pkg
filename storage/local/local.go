package local

import (
	"os"
	"io"
)

var (
	fileCreate = os.Create
	fileOpen   = os.Open
)

type Provider struct {
	path string
}

func (p Provider) Kind() string {
	return "local"
}

func (p Provider) CollectionRef() string {
	return p.path
}

func (p Provider) UploadFile(filename, reference string) error {
	// Copy to "uploads" folder.
	dest := p.path + "/" + reference
	return copyFile(filename, dest)
}

func (p Provider) DownloadFile(reference, filename string) error {
	// Copy from "uploads" folder.
	src := p.path + "/" + reference
	return copyFile(src, filename)
}

func NewLocalStorage(storagePath string) *Provider {
	return &Provider{
		storagePath,
	}
}

func copyFile(src, dst string) error {

	// Open source file to copy from.
	sourceFile, err := fileOpen(src)

	// Error if file cannot be opened.
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination file for writing.
	destFile, err := fileCreate(dst)

	// Error if file cannot be created.
	if err != nil {
		return err
	}
	defer destFile.Close()

	io.Copy(destFile, sourceFile)

	return nil
}
