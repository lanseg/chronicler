package storage

import (
	"bytes"
	"encoding/json"
	"io"
)

type BlockStorage struct {
	Storage
}

func (bs *BlockStorage) PutBytes(put *PutRequest, data []byte) (int64, error) {
	writer, err := bs.Put(put)
	if err != nil {
		return -1, err
	}
	defer writer.Close()
	return io.Copy(writer, bytes.NewReader(data))
}

func (bs *BlockStorage) GetBytes(get *GetRequest) ([]byte, error) {
	reader, err := bs.Get(get)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func (bs *BlockStorage) PutObject(put *PutRequest, data interface{}) (int64, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return -1, err
	}
	return bs.PutBytes(put, jsonBytes)
}

func (bs *BlockStorage) GetObject(get *GetRequest, data interface{}) error {
	jsonBytes, err := bs.GetBytes(get)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, data)
}
