package frontend

import (
	"fmt"
	"os"
	"strings"

	"chronicler/util"
	"net/http"
	"path/filepath"

	rpb "chronicler/proto/records"
)

type WebServer struct {
	storageRoot string
	server      *http.Server
	logger      *util.Logger
}

func (ws *WebServer) Error(w http.ResponseWriter, msg string, code int) {
	ws.logger.Warningf("HTTP %d: %s", code, msg)
	http.Error(w, msg, code)
}

func (ws *WebServer) handleRecordListRequest(w http.ResponseWriter, r *http.Request) {
	ws.logger.Infof("Requesting record list: %s", r.URL.String())
	// params := r.URL.Query()
	// sourceType := rpb.SourceType(rpb.SourceType_value[strings.ToUpper(params.Get("type"))])

}

func (ws *WebServer) handleRecordRequest(w http.ResponseWriter, r *http.Request) {
	ws.logger.Infof("Requesting record: %s", r.URL.String())
	params := r.URL.Query()
	recordId := params.Get("id")
	fname := "record.json"
	sourceType := rpb.SourceType(rpb.SourceType_value[strings.ToUpper(params.Get("type"))])
	if params.Has("file") {
		fname = params.Get("file")
	}
	b, err := os.ReadFile(filepath.Join(ws.storageRoot,
		fmt.Sprintf("%s", sourceType), recordId, fname))
	if err != nil {
		ws.Error(w, fmt.Sprintf("File %s/%s/%s not found", sourceType, recordId, fname), 500)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		ws.Error(w, err.Error(), 500)
	}
}

func (ws *WebServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	sourceType := rpb.SourceType(rpb.SourceType_value[strings.ToUpper(params.Get("type"))])
	if sourceType == rpb.SourceType_UNKNOWN_TYPE {
		ws.Error(w, fmt.Sprintf("Unknown source type: \"%s\"", params.Get("type")), 500)
		return
	}
	if !params.Has("id") {
		ws.handleRecordListRequest(w, r)
		return
	}
	ws.handleRecordRequest(w, r)
}

func (ws *WebServer) Start() {
	if err := ws.server.ListenAndServe(); err != nil {
		ws.logger.Errorf("Failed to start server: %s", err)
	}
}

func NewServer(port int, storageRoot string, staticFiles string) *WebServer {
	server := &WebServer{
		logger:      util.NewLogger("frontend"),
		storageRoot: storageRoot,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chronicler", server.handleRequest)
	mux.Handle("/", http.FileServer(http.Dir(staticFiles)))

	server.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	return server
}
