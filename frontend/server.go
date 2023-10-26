package frontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/storage"
	"chronicler/webdriver"
	"web/htmlparser"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
)

const (
	textSampleSize = 512
)

type DataRequest struct {
	id       string
	filename string
}

func (d DataRequest) String() string {
	return fmt.Sprintf("DataRequest {id: \"%s\", file: \"%s\"}", d.id, d.filename)
}

func parseUrlRequest(link *url.URL) (*DataRequest, error) {
	path := link.Path
	query := link.Query()
	params := DataRequest{}
	for i, param := range strings.Split(strings.TrimPrefix(path, "/chronicler/"), "/") {
		switch i {
		case 0:
			params.id = param
		default:
			return nil, fmt.Errorf("Unsupported path parameter #%d: %s", i, param)
		}
	}
	if query["file"] != nil {
		params.filename = query.Get("file")
	}
	return &params, nil
}

type WebServer struct {
	http.Handler

	staticFileServer http.Handler
	storage          storage.Storage
	server           *http.Server
	logger           *cm.Logger
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

func (ws *WebServer) responseRecordList(w http.ResponseWriter) {
	rs, _ := ws.storage.ListRecords().Get()
	rs = records.SortRecordSets(rs)

	userById := map[string]*rpb.UserMetadata{}
	result := &rpb.RecordListResponse{}
	for _, r := range rs {
		desc := ""
		if len(r.Records) > 0 {
			desc = cm.IfEmpty(
				string(r.Records[0].RawContent),
				r.Records[0].TextContent)

			if r.Records[0].Source.Type == rpb.SourceType_WEB {
				desc = htmlparser.GetTitle(desc)
			}
		}
		if len(desc) > textSampleSize {
			desc = cm.Ellipsis(desc, textSampleSize, true)
		}
		set := &rpb.RecordListResponse_RecordSetInfo{
			Id:          r.Id,
			Description: desc,
			RecordCount: int32(len(r.Records)),
		}
		if len(r.Records) > 0 {
			set.RootRecord = r.Records[0]
		}
		result.RecordSets = append(result.RecordSets, set)
		for _, data := range r.UserMetadata {
			userById[data.Id] = data
		}
	}
	sort.Slice(result.RecordSets, func(i int, j int) bool {
		left := result.RecordSets[i].RootRecord
		right := result.RecordSets[j].RootRecord
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		return left.Time > right.Time
	})

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
		rs := &rpb.RecordSet{}
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

func (ws *WebServer) handleApiRequest(w http.ResponseWriter, r *http.Request) {
	params, err := parseUrlRequest(r.URL)
	ws.logger.Infof("Request [api]: %s (%s)", r.URL.String(), params)
	if err != nil {
		ws.Error(w, err.Error(), 422)
		return
	}
	if params.id == "" && params.filename == "" {
		ws.responseRecordList(w)
		return
	} else if params.filename == "" {
		ws.responseFile(w, params.id, "record.json")
		return
	} else {
		ws.responseFile(w, params.id, params.filename)
	}
	w.Write([]byte(":)"))
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
		logger:           cm.NewLogger("frontend"),
		storage:          storage.NewStorage(storageRoot, &webdriver.ExclusiveWebDriver{}),
		staticFileServer: http.FileServer(http.Dir(staticFiles)),
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server,
	}
}
