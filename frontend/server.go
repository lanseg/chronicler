package frontend

import (
	"encoding/json"
	"fmt"
	"strings"

	"chronicler/storage"
	"chronicler/util"
	"net/http"
	"net/url"

	rpb "chronicler/proto/records"
)

type DataRequest struct {
	sourceType rpb.SourceType
	id         string
	filename   string
}

func (d DataRequest) String() string {
	return fmt.Sprintf("DataRequest {sourceType: \"%s\", id: \"%s\", file: \"%s\"}",
		d.sourceType, d.id, d.filename)
}

func parseUrlRequest(link *url.URL) (*DataRequest, error) {
	path := link.Path
	params := DataRequest{}
	for i, param := range strings.Split(strings.TrimPrefix(path, "/chronicler/"), "/") {
		switch i {
		case 0:
			params.sourceType = rpb.SourceType(rpb.SourceType_value[strings.ToUpper(param)])
		case 1:
			params.id = param
		case 2:
			params.filename = param
		default:
			return nil, fmt.Errorf("Unsupported path parameter #%d: %s", i, param)
		}
	}
	return &params, nil
}

type WebServer struct {
	http.Handler

	staticFileServer http.Handler
	storage          storage.Storage
	server           *http.Server
	logger           *util.Logger
}

func (ws *WebServer) Error(w http.ResponseWriter, msg string, code int) {
	ws.logger.Warningf("HTTP %d: %s", code, msg)
	http.Error(w, msg, code)
}

func (ws *WebServer) handleRecordRequest(w http.ResponseWriter, r *http.Request) {
	ws.logger.Infof("Requesting record: %s", r.URL.String())
}

func (ws *WebServer) writeJson(w http.ResponseWriter, data any) {
	bytes, err := json.Marshal(data)
	if err != nil {
		ws.Error(w, fmt.Sprintf("Marshalling error: %s", err.Error()), 500)
		return
	}
	w.Write(bytes)
}

func (ws *WebServer) responseSourceTypes(w http.ResponseWriter) {
	values := []string{}
	for name, i := range rpb.SourceType_value {
		if i == 0 {
			continue
		}
		values = append(values, name)
	}
	ws.writeJson(w, values)
}

func (ws *WebServer) responseIdsForSource(w http.ResponseWriter, srcType rpb.SourceType) {
	records, err := ws.storage.ListRecords()
	if err != nil {
		ws.Error(w, fmt.Sprintf("Cannot enumerate records for %s", srcType), 500)
		return
	}
	for _, r := range records {
		if len(r.Records) == 0 {
			continue
		}
		w.Write([]byte(fmt.Sprintf("%s\n", r.Records[0])))
	}
}

func (ws *WebServer) handleApiRequest(w http.ResponseWriter, r *http.Request) {
	params, err := parseUrlRequest(r.URL)
	ws.logger.Infof("Request [api]: %s (%s)", r.URL.String(), params)
	if err != nil {
		ws.Error(w, err.Error(), 422)
		return
	}

	if params.sourceType == 0 && params.id == "" && params.filename == "" {
		ws.responseSourceTypes(w)
		return
	} else if params.id == "" && params.filename == "" {
		ws.responseIdsForSource(w, params.sourceType)
		return
	}
	w.Write([]byte(fmt.Sprintf("%s", params)))
}

func (ws *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/chronicler") {
		ws.handleApiRequest(w, r)
		return
	}
	ws.logger.Infof("Request [static]: %s", r.URL.Path)
	ws.staticFileServer.ServeHTTP(w, r)
}

func NewServer(port int, storageRoot string, staticFiles string) *http.Server {
	server := &WebServer{
		logger:           util.NewLogger("frontend"),
		storage:          storage.NewStorage(storageRoot),
		staticFileServer: http.FileServer(http.Dir(staticFiles)),
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server,
	}
}
