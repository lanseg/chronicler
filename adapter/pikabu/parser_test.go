package pikabu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		name    string
		file    string
		wantErr string
	}{
		{
			name: "post no text with comments",
			file: "pikabu_10819340.html",
		},
		{
			name: "post text only with comments",
			file: "pikabu_11261377.html",
		},
		{
			name: "post with pikabu video links",
			file: "pikabu_video_11265076.html",
		},
		{
			name: "post with video links loaded by curl",
			file: "pikabu_video_curl_11266557.html",
		},
		{
			name: "comment placeholder causes panic",
			file: "pikabu_panic_11298350.html",
		},
		{
			name: "comment placeholder causes panic",
			file: "pikabu_panic_11298350.html",
		},
		{
			name:    "page not found",
			file:    "pikabu_notfound_1.html",
			wantErr: "Page was removed",
		},
		{
			name:    "page deleted",
			file:    "pikabu_deleted_11298856.html",
			wantErr: "Page was removed",
		},
		{
			name: "pikabu jobs",
			file: "pikabu_jobs_11308419.html",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tc.file))
			if err != nil {
				t.Errorf("Error while reading testdata file %s: %s", tc.file, err)
				return
			}
			result, err := parsePost(string(data), timeSrc)
			if err != nil && tc.wantErr == "" {
				t.Errorf("Unexpected error while parsing page %s: %s", tc.file, err)
				return
			}
			if err == nil && tc.wantErr != "" {
				t.Errorf("Expecting error %q, but got none", tc.wantErr)
				return
			}
			if err != nil && tc.wantErr != "" {
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("Expected error %q, but got %q", tc.wantErr, err.Error())
					return
				} else {
					return
				}
			}

			want, err := readJson[rpb.Response](fmt.Sprintf("%s.json", tc.file))
			if err != nil {
				t.Errorf("Error while parsing json file %s.json: %s", tc.file, err)
				return
			}

			result.Result = records.SortRecordSets(result.Result, &rpb.Sorting{Field: rpb.Sorting_CREATE_TIME, Order: rpb.Sorting_ASC})
			want.Result = records.SortRecordSets(want.Result, &rpb.Sorting{Field: rpb.Sorting_CREATE_TIME, Order: rpb.Sorting_ASC})

			resultStr := fmt.Sprintf("%s", result)
			wantStr := fmt.Sprintf("%s", want)

			if resultStr != wantStr {
				res, _ := json.Marshal(result.Result)
				wnt, _ := json.Marshal(want.Result)
				fmt.Println(string(res))
				fmt.Println(string(wnt))

				t.Errorf("Expected and actual result for %q differ", tc.name)
			}
		})
	}
}
