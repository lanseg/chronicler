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
}

type ListResponse struct {
	Url []string
}

type Storage interface {
	Put(put *PutRequest) (io.WriteCloser, error)
	Get(get *GetRequest) (io.ReadCloser, error)
	List(list *ListRequest) (*ListResponse, error)
}
