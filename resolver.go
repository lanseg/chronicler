package main

import (
	"bytes"
	"encoding/base64"
	"net/http"

	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/optional"

	"chronicler/downloader"
	rpb "chronicler/records/proto"
	"chronicler/storage"
	"chronicler/webdriver"
)

type Resolver interface {
	Resolve(id string) error
}

func NewResolver(browser webdriver.Browser, storage storage.Storage) Resolver {
	return &resolverImpl{
		logger:     cm.NewLogger("Resolver"),
		browser:    browser,
		storage:    storage,
		downloader: downloader.NewDownloader(&http.Client{}, storage),
	}
}

type resolverImpl struct {
	Resolver

	logger     *cm.Logger
	storage    storage.Storage
	browser    webdriver.Browser
	downloader downloader.Downloader
}

func newFile(id string, name string) *rpb.File {
	return &rpb.File{
		FileId:   id,
		LocalUrl: name,
	}
}

func (r *resolverImpl) Resolve(id string) error {
	rs, err := r.storage.GetRecordSet(id).Get()
	if err != nil {
		return err
	}
	for _, rec := range rs.Records {
		for _, file := range rec.GetFiles() {
			r.logger.Infof("Started downloading %s", file)
			if err := r.downloader.ScheduleDownload(rs.Id, file.FileUrl); err != nil {
				r.logger.Warningf("Could not download file %s", file)
			}
		}

		if rec.Source != nil && rec.Source.Url != "" {
			r.savePageView(rs.Id, rec.Source.Url)
			rec.Files = append(rec.Files,
				newFile("page_view_png", "pageview_page.png"),
				newFile("page_view_pdf", "pageview_page.pdf"),
				newFile("page_view_html", "pageview_page.html"),
			)
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
