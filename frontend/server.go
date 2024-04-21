package frontend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	col "github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
	opt "github.com/lanseg/golang-commons/optional"

	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/status"
	"chronicler/storage"
)

type DeleteRecordResponse struct {
	Id      string
	Deleted bool
	Error   error
}

type WebServer struct {
	data   storage.Storage
	logger *cm.Logger

	stats   status.StatusClient
	sorting *rpb.Sorting
}

func parseListRequest(v url.Values) *rpb.ListRecordsRequest {
	result := &rpb.ListRecordsRequest{
		Paging: &rpb.Paging{
			Offset: uint64(0),
			Size:   uint64(10),
		},
		Query:   v.Get("query"),
		Sorting: &rpb.Sorting{Field: rpb.Sorting_CREATE_TIME, Order: rpb.Sorting_DESC},
	}
	opt.OfError(strconv.ParseUint(v.Get("offset"), 10, 64)).IfPresent(cm.Set[uint64](&result.Paging.Offset))
	opt.OfError(strconv.ParseUint(v.Get("size"), 10, 64)).IfPresent(cm.Set[uint64](&result.Paging.Size))
	return result
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
	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

func (ws *WebServer) handleRecordSetList(p PathParams, w http.ResponseWriter, r *http.Request) {
	rs := records.SortRecordSets(
		ws.data.ListRecordSets(parseListRequest(r.URL.Query())).OrElse([]*rpb.RecordSet{}),
		&rpb.Sorting{Field: rpb.Sorting_CREATE_TIME, Order: rpb.Sorting_ASC})

	userById := map[string]*rpb.UserMetadata{}
	result := &rpb.RecordListResponse{}
	for _, r := range rs {
		result.RecordSets = append(result.RecordSets, records.CreatePreview(r))
		for _, data := range r.UserMetadata {
			userById[data.Id] = data
		}
	}
	records.SortPreviews(result.RecordSets, &rpb.Sorting{Field: rpb.Sorting_CREATE_TIME, Order: rpb.Sorting_DESC})
	result.UserMetadata = col.Values(userById)
	ws.writeJson(w, result)
}

func (ws *WebServer) responseFile(w http.ResponseWriter, id string, filename string) {
	data, err := opt.MapErr(ws.data.GetFile(id, filename), func(rc io.ReadCloser) ([]byte, error) {
		defer rc.Close()
		return io.ReadAll(rc)
	}).Get()

	if err != nil {
		ws.Error(w, err.Error(), 500)
		return
	}

	if filename == "record.json" {
		rs := &rpb.RecordSet{}
		err = json.Unmarshal(data, rs)
		if err != nil {
			ws.Error(w, err.Error(), 500)
			return
		}
		ws.writeJson(w, rs)
		return
	}
	w.Write(data)
}

func (ws *WebServer) handleDeleteRecord(p PathParams, w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	ws.logger.Infof("Request [delete]: %s", queryParams)

	idsToDelete := strings.Split(queryParams.Get("ids"), ",")
	var err error
	for _, id := range idsToDelete {
		err = ws.data.DeleteRecordSet(id)
	}
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

func (ws *WebServer) handleStatus(p PathParams, w http.ResponseWriter, r *http.Request) {
	stats, err := ws.stats.GetValues()
	if err != nil {
		ws.Error(w, err.Error(), 500)
		return
	}
	ws.writeJson(w, stats)
}

func NewServer(port int, staticFiles string, storage storage.Storage, stats status.StatusClient) *http.Server {
	server := &WebServer{
		logger: cm.NewLogger("frontend"),
		data:   storage,
		stats:  stats,
	}

	handler := &PathParamHandler{
		elseHandler: http.FileServer(http.Dir(staticFiles)),
	}

	handler.Handle("/chronicler/records/delete", server.handleDeleteRecord)
	handler.Handle("/chronicler/records/{recordId}", server.handleRecord)
	handler.Handle("/chronicler/records", server.handleRecordSetList)
	handler.Handle("/chronicler/status", server.handleStatus)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
}
