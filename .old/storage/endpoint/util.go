package endpoint

import (
	"fmt"
	"io"

	ep "chronicler/storage/endpoint_go_proto"
)

type FileData struct {
	Data  []byte
	Error string
}

func (f *FileData) String() string {
	return fmt.Sprintf("FileData { Data: %q, Error: %q }", f.Data, f.Error)
}

func WriteAll(server ep.Storage_GetFileServer, data io.ReadCloser, id int, chunkSize int) error {
	defer data.Close()
	buf := make([]byte, chunkSize)
	chunkId := 0
	for {
		end, err := data.Read(buf)
		if err != nil {
			return err
		}
		if err := server.Send(&ep.GetFileResponse{
			Part: &ep.FilePart{
				FileId: int32(id),
				Data: &ep.FilePart_Chunk_{
					Chunk: &ep.FilePart_Chunk{
						ChunkId: int32(chunkId),
						Data:    buf[:end],
					},
				},
			},
		}); err != nil {
			return err
		}
		chunkId++
	}
	return nil
}

func ReadAll(client ep.Storage_GetFileClient, err error) ([]*FileData, error) {
	if err != nil {
		return nil, err
	}
	receivedData := map[int]*FileData{}
	maxId := 0
	for {
		part, err := client.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		chunk := part.Part
		fileId := int(chunk.FileId)
		if fileId > maxId {
			maxId = fileId
		}
		if receivedData[fileId] == nil {
			receivedData[fileId] = &FileData{}
		}
		if chErr := chunk.GetError(); chErr != nil {
			receivedData[fileId].Error = chErr.Error
			continue
		}
		receivedData[fileId].Data = append(receivedData[fileId].Data, chunk.GetChunk().Data...)
	}
	result := make([]*FileData, maxId+1)
	for i := 0; i <= maxId; i++ {
		if data, ok := receivedData[i]; ok {
			result[i] = data
		}
	}
	return result, nil
}
