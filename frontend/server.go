package frontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"chronicler/downloader"
	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/storage"
	"chronicler/webdriver"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
)

type DeleteRecordResponse struct {
	Id      string
	Deleted bool
	Error   error
}

type WebServer struct {
	storage storage.Storage
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
	rs, _ := ws.storage.ListRecordSets().Get()
	rs = records.SortRecordSets(rs)

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
	f, err := ws.storage.GetFile(id, filename).Get()
	if err != nil {
		ws.Error(w, err.Error(), 500)
		return
	}
	if filename == "record.json" {
		rs := records.NewRecordSet(&rpb.RecordSet{})
		err = json.Unmarshal(f, rs)
		if err != nil {
			ws.Error(w, err.Error(), 500)
			return
		}
		rs.Records = records.SortRecords(rs.Records)
		ws.writeJson(w, rs)
		return
	}
	w.Write(f)
}

func (ws *WebServer) handleDeleteRecord(p PathParams, w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	ws.logger.Infof("Request [delete]: %s", queryParams)

	result := []*DeleteRecordResponse{}
	for _, r := range strings.Split(queryParams.Get("ids"), ",") {
		err := ws.storage.DeleteRecordSet(r)
		if err != nil {
			ws.logger.Warningf("Could not delete record %q: %s", r, err)
		}
		result = append(result, &DeleteRecordResponse{
			Id:      r,
			Deleted: err == nil,
			Error:   err,
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

func NewServer(port int, storageRoot string, staticFiles string) *http.Server {
	server := &WebServer{
		logger:  cm.NewLogger("frontend"),
		storage: storage.NewStorage(storageRoot, webdriver.NewFakeBrowser(nil), downloader.NewNoopDownloader()),
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
