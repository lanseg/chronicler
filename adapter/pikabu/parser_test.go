package pikabu

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	cm "github.com/lanseg/golang-commons/common"

	"chronicler/records"
	rpb "chronicler/records/proto"
)

func readJson[T any](file string) (*T, error) {
	bytes, err := os.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		return nil, err
	}
	return cm.FromJson[T](bytes)
}

func TestPikabuParser(t *testing.T) {

	timeSrc := func() time.Time {
		fakeTime, _ := time.Parse(time.RFC3339, "2024-03-27T10:04:05Z")
		return fakeTime
	}

	for _, tc := range []struct {
		name string
		file string
	}{
		// {
		// 	name: "post no text with comments",
		// 	file: "pikabu_10819340.html",
		// },
		// {
		// 	name: "post text only with comments",
		// 	file: "pikabu_11261377.html",
		// },
		{
			name: "post with pikabu video links",
			file: "pikabu_video_11265076.html",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tc.file))
			if err != nil {
				t.Errorf("Error while reading testdata file %s: %s", tc.file, err)
				return
			}
			result, err := parsePost(string(data), timeSrc)
			if err != nil {
				t.Errorf("Error while parsing file %s: %s", tc.file, err)
				return
			}

			want, err := readJson[rpb.Response](fmt.Sprintf("%s.json", tc.file))
			if err != nil {
				t.Errorf("Error while parsing json file %s.json: %s", tc.file, err)
				return
			}

			records.SortRecordSets(result.Result, &rpb.Sorting{})
			records.SortRecordSets(want.Result, &rpb.Sorting{})

			resultStr := fmt.Sprintf("%s", result)
			wantStr := fmt.Sprintf("%s", want)

			if resultStr != wantStr {
				t.Errorf("Expected and actual result for %q differ", tc.name)
			}
		})
	}
}
