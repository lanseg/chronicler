package storage

import (
	"io"
)

type PutRequest struct {
	Url             string
	SaveOnOverwrite bool
}

type GetRequest struct {
	Url string
}

type ListRequest struct {
	WithSnapshots bool
	Url           []string
}

type StorageItem struct {
	Url      string
	Versions []string
}

type ListResponse struct {
	Items []StorageItem
}

type Storage interface {
	Put(put *PutRequest) (io.WriteCloser, error)
	Get(get *GetRequest) (io.ReadCloser, error)
	List(list *ListRequest) (*ListResponse, error)
}
