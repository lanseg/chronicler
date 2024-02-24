package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/storage/endpoint"
	ep "chronicler/storage/endpoint_go_proto"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
)

type DeleteRecordResponse struct {
	Id      string
	Deleted bool
	Error   error
}

type WebServer struct {
	storage ep.StorageClient
	logger  *cm.Logger
}

func (ws *WebServer) Error(w http.ResponseWriter, msg string, code int) {
	ws.logger.Warningf("HTTP %d: %s", code, msg)
	http.Error(w, msg, code)
}

func (ws *WebServer) writeJson(w http.ResponseWriter, data any) {
	bytes, err := json.Marshal(data)
	if err != nil {
		ws.Error(w, fmt.Sprintf("Marshalling error: %s", err.Error()), 500)
		return
	}
	w.Write(bytes)
}

func (ws *WebServer) handleRecordSetList(p PathParams, w http.ResponseWriter, r *http.Request) {
	recv, err := ws.storage.List(context.Background(), &ep.ListRequest{})
	if err != nil {
		ws.Error(w, fmt.Sprintf("Cannot get RecordSets: %s", err.Error()), 500)
		return
	}
	sets := []*rpb.RecordSet{}
	for {
		rs, err := recv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			ws.Error(w, fmt.Sprintf("Error while getting RecordSets: %s", err), 500)
			return
		}
		if err == io.EOF {
			break
		}
		sets = append(sets, rs.RecordSet)
	}

	rs := records.SortRecordSets(sets)

	userById := map[string]*rpb.UserMetadata{}
	result := &rpb.RecordListResponse{}
	for _, r := range rs {
		result.RecordSets = append(result.RecordSets, records.CreatePreview(r))
		for _, data := range r.UserMetadata {
			userById[data.Id] = data
		}
	}
	result.UserMetadata = collections.Values(userById)
	ws.writeJson(w, result)
}

func (ws *WebServer) responseFile(w http.ResponseWriter, id string, filename string) {
	files, err := endpoint.ReadAll(ws.storage.GetFile(context.Background(), &ep.GetFileRequest{
		File: []*ep.GetFileRequest_FileDef{
			{RecordSetId: id, Filename: filename},
		},
	}))
	if err != nil {
		ws.Error(w, err.Error(), 500)
		return
	}
	if files == nil || len(files) == 0 || files[0] == nil {
		ws.Error(w, fmt.Sprintf("File %s/%s not found", id, filename), 404)
		return
	}

	if filename == "record.json" {
		rs := records.NewRecordSet(&rpb.RecordSet{})
		err = json.Unmarshal(files[0].Data, rs)
		if err != nil {
			ws.Error(w, err.Error(), 500)
			return
		}
		rs.Records = records.SortRecords(rs.Records)
		ws.writeJson(w, rs)
		return
	}
	w.Write(files[0].Data)
}

func (ws *WebServer) handleDeleteRecord(p PathParams, w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	ws.logger.Infof("Request [delete]: %s", queryParams)

	idsToDelete := strings.Split(queryParams.Get("ids"), ",")
	_, err := ws.storage.Delete(context.Background(), &ep.DeleteRequest{
		RecordSetIds: idsToDelete,
	})
	if err != nil {
		ws.Error(w, err.Error(), 500)
		return
	}

	result := []*DeleteRecordResponse{}
	for _, r := range idsToDelete {
		result = append(result, &DeleteRecordResponse{
			Id:      r,
			Deleted: true,
			Error:   nil,
		})
	}
	ws.writeJson(w, result)
}

func (ws *WebServer) handleRecord(p PathParams, w http.ResponseWriter, r *http.Request) {
	ws.logger.Infof("Request [api]: %s", p)
	queryParams := r.URL.Query()
	filename := "record.json"
	if queryParams["file"] != nil {
		filename = queryParams.Get("file")
	}
	ws.responseFile(w, p["recordId"], filename)
}

func NewServer(port int, staticFiles string, storageClient ep.StorageClient) *http.Server {
	server := &WebServer{
		logger:  cm.NewLogger("frontend"),
		storage: storageClient,
	}

	handler := &PathParamHandler{
		elseHandler: http.FileServer(http.Dir(staticFiles)),
	}

	handler.Handle("/chronicler/records/delete", server.handleDeleteRecord)
	handler.Handle("/chronicler/records/{recordId}", server.handleRecord)
	handler.Handle("/chronicler/records", server.handleRecordSetList)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
}
