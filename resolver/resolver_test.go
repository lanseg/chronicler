package resolver

import (
	"reflect"
	"testing"

	"chronicler/adapter"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
)

type fakeDownloader struct {
	common.Downloader

	urls []string
}

func (fd *fakeDownloader) Download(url string, s storage.Storage) (int64, error) {
	fd.urls = append(fd.urls, url)
	return 0, nil
}

type fakeAdapter struct {
	adapter.Adapter

	objects []*opb.Object
}

func (fa *fakeAdapter) Match(link *opb.Link) bool {
	return true
}

func (fa *fakeAdapter) Get(link *opb.Link) ([]*opb.Object, error) {
	return fa.objects, nil
}

func newFakeAdapter(o ...*opb.Object) adapter.Adapter {
	return &fakeAdapter{
		objects: o,
	}
}

func TestResolver(t *testing.T) {
	t.Run("resolver start stop", func(t *testing.T) {
		root := t.TempDir()
		loader := &fakeDownloader{
			urls: []string{},
		}
		adapters := []adapter.Adapter{
			newFakeAdapter(&opb.Object{
				Id: "123",
				Attachment: []*opb.Attachment{
					{Url: "http://some/other/url", Mime: "text/html"},
				},
			}),
		}
		r := NewResolver(root, loader, adapters)
		r.Start()
		if err := r.Resolve(&opb.Link{Href: "http://some/url"}); err != nil {
			t.Errorf("Failed while resolving: %q", err)
		}
		r.Wait()
		r.Stop()

		if !reflect.DeepEqual(loader.urls, []string{"http://some/other/url"}) {
			t.Error("Expected url not resolved")
		}
	})
}
