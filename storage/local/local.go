package local

import (
	"io"
	"os"
)

var (
	fileCreate = os.Create
	fileOpen   = os.Open
)

// Provider is a local storage provider.
type Provider struct {
	serverPath string
	localPath  string
}

// Kind returns the kind of provider.
func (p Provider) Kind() string {
	return "local"
}

// CollectionRef returns the path for the provider.
func (p Provider) CollectionRef() string {
	return p.localPath
}

// UploadFile copies the file to a destination.
func (p Provider) UploadFile(filename, reference string) error {
	// Copy to "uploads" folder.
	dest := p.serverPath + "/" + reference
	return copyFile(filename, dest)
}

// DownloadFile copies the file from the storage provider.
func (p Provider) DownloadFile(reference, filename string) error {
	// Copy from "uploads" folder.
	src := p.serverPath + "/" + reference
	return copyFile(src, filename)
}

// NewLocalStorage returns a local storage provider.
func NewLocalStorage(storagePath string, localPath string) *Provider {
	return &Provider{
		storagePath,
		localPath,
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
