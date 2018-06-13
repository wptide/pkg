package storage

// Provider interface describes the methods required to upload or download files from a storage provider.
type Provider interface {
	Kind() string
	CollectionRef() string
	UploadFile(filename, reference string) error
	DownloadFile(reference, filename string) error
}
