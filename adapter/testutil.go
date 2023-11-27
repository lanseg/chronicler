package adapter

import (
	"chronicler/webdriver"

	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lanseg/golang-commons/optional"
)

func writeJson(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}

func readJson(file string, obj interface{}) error {
	bytes, err := os.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &obj)
	if err != nil {
		return err
	}
	return nil
}

type fakeWebDriver struct {
	webdriver.NoopWebdriver

	file string
	url  string
}

func (fd *fakeWebDriver) Navigate(url string) {
	fd.url = url
}

func (fd *fakeWebDriver) GetPageSource() optional.Optional[string] {
	return optional.Map(
		optional.OfError(os.ReadFile(filepath.Join("testdata", fd.file))),
		func(b []byte) string {
			return string(b)
		})
}

func (fd *fakeWebDriver) GetCurrentURL() optional.Optional[string] {
	return optional.Of(fd.url)
}

func newFakeWebdriver(file string) webdriver.WebDriver {
	return &fakeWebDriver{
		file: file,
	}
}
