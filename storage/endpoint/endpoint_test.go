package endpoint

import (
	"fmt"
	"testing"

	"chronicler/downloader"
	rpb "chronicler/records/proto"
	"chronicler/webdriver"
)

type fakeBrowser struct {
	webdriver.Browser
}

func (fws *fakeBrowser) RunSession(func(webdriver.WebDriver)) error {
	return nil
}

type fakeDownloader struct {
	downloader.Downloader
}

func (fd *fakeDownloader) ScheduleDownload(string, string) error {
	return nil
}

type FakeDriver struct {
	webdriver.WebDriver
}

func newRecordSet(id int, name string) *rpb.RecordSet {
	rs := &rpb.RecordSet{
		Id: fmt.Sprintf("%s %d", name, id),
	}
	return rs
}

func TestStorage(t *testing.T) {
	t.Run("Test endpoint", func(t *testing.T) {
	})
}
