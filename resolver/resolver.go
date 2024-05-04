package resolver

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"sync/atomic"

	cm "github.com/lanseg/golang-commons/common"
	conc "github.com/lanseg/golang-commons/concurrent"
	"github.com/lanseg/golang-commons/optional"

	rpb "chronicler/records/proto"
	"chronicler/status"
	"chronicler/storage"
	"chronicler/webdriver"
)

const (
	workerCount = 10
)

type Resolver interface {
	Resolve(id string) error
}

type resolverImpl struct {
	Resolver

	activeWorkers *atomic.Uint32
	browser       webdriver.Browser
	logger        *cm.Logger
	pool          conc.Executor
	stats         status.StatusClient
	storage       storage.Storage
}

func NewResolver(browser webdriver.Browser, storage storage.Storage, stats status.StatusClient) Resolver {
	return &resolverImpl{
		activeWorkers: &atomic.Uint32{},
		browser:       browser,
		logger:        cm.NewLogger("Resolver"),
		pool:          cm.OrExit(conc.NewPoolExecutor(workerCount)),
		stats:         stats,
		storage:       storage,
	}
}

func newFile(id string, name string) *rpb.File {
	return &rpb.File{
		FileId:   id,
		LocalUrl: name,
	}
}

func (r *resolverImpl) addWorkerCount(i uint32) {
	r.stats.PutInt("resolver.workers.active", int64(r.activeWorkers.Add(i)))
}

func (r *resolverImpl) Resolve(id string) error {
	rs, err := r.storage.GetRecordSet(id).Get()
	if err != nil {
		return err
	}
	for _, rec := range rs.Records {
		for _, file := range rec.GetFiles() {
			r.pool.Execute(func() {
				r.addWorkerCount(uint32(1))
				downloader := NewHttpDownloader(&http.Client{}, r.storage, r.stats)
				r.logger.Infof("Started downloading %s", file)
				if err := downloader.Download(rs.Id, file.FileUrl); err != nil {
					r.logger.Warningf("Could not download file %s", file)
				}
				r.addWorkerCount(^uint32(0))
			})
		}

		if rec.Source != nil && rec.Source.Url != "" {
			r.pool.Execute(func() {
				r.addWorkerCount(uint32(1))
				r.stats.PutString("resolver.webdriver.pageview", rec.Source.Url)
				r.savePageView(rs.Id, rec.Source.Url)
				rec.Files = append(rec.Files,
					newFile("page_view_png", "pageview_page.png"),
					newFile("page_view_pdf", "pageview_page.pdf"),
					newFile("page_view_html", "pageview_page.html"),
				)
				r.stats.DeleteMetric("resolver.webdriver.pageview")
				r.addWorkerCount(^uint32(0))
			})
		}
	}
	return r.storage.SaveRecordSet(rs)
}

func (r *resolverImpl) saveBase64(id string, fname string) func(string) {
	return func(content string) {
		_, err := optional.MapErr(
			optional.OfError(base64.StdEncoding.DecodeString(content)),
			func(decoded []byte) (int, error) {
				return 0, r.storage.PutFile(id, fname, bytes.NewReader(decoded))
			}).Get()
		if err == nil {
			r.logger.Debugf("Written %q/%q", id, fname)
		} else {
			r.logger.Warningf("Could not save base64 data to %q/%q: %v", id, fname, err)
		}
	}
}

func (r *resolverImpl) savePageView(id string, url string) {
	r.browser.RunSession(func(d webdriver.WebDriver) {
		d.Navigate(url)
		d.TakeScreenshot().IfPresent(r.saveBase64(id, "pageview_page.png"))
		d.Print().IfPresent(r.saveBase64(id, "pageview_page.pdf"))
		d.GetPageSource().IfPresent(func(src string) {
			r.storage.PutFile(id, "pageview_page.html", bytes.NewReader([]byte(src)))
		})
	})
}
