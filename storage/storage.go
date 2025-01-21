package storage

import (
	"io"
)

type PutRequest struct {
	Url string
}

type GetRequest struct {
	Url string
}

type ListRequest struct {
}

type ListResponse struct {
}

type Storage interface {
	Put(put *PutRequest) (io.WriteCloser, error)
	Get(get *GetRequest) (io.ReadCloser, error)
	List(list *ListRequest) (*ListResponse, error)
}
