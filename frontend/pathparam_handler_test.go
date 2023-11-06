package frontend

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"
)

func paramResponse(p PathParams, w http.ResponseWriter, r *http.Request) {
	kv := []string{}
	for k, v := range p {
		kv = append(kv, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(kv)
	w.Write([]byte(strings.Join(kv, ", ")))
}

func TestPathMapper(t *testing.T) {
	handler := &PathParamHandler{}
	handler.Handle("/chronicler/{firstParam}/{secondParam}", paramResponse)
	handler.Handle("/wrong{url}handler{url1}", paramResponse)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", testingPort+1),
		Handler: handler,
	}

	go (func() {
		if err := server.ListenAndServe(); err != nil {
			t.Fatalf("Could not start a server: %s", err)
		}
	})()

	time.Sleep(3 * time.Second)

	for _, tc := range []struct {
		desc string
		path string
		want string
	}{
		{
			desc: "Simple two parameter path",
			path: "chronicler/arg0/arg1",
			want: "firstParam=arg0, secondParam=arg1",
		},
		{
			desc: "Simple two parameters with query params",
			path: "chronicler/arg1/arg2?result=21",
			want: "firstParam=arg1, secondParam=arg2",
		},
		{
			desc: "Duplicating slashes",
			path: "chronicler//arg1///arg2///",
			want: "",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			data, err := get(fmt.Sprintf("http://localhost:%d/%s", testingPort+1, tc.path))
			if err != nil {
				t.Errorf("Could not fetch data from %s", tc.path)
			}
			if string(data) != tc.want {
				t.Errorf("Expected get(%s) to return %s, but got %s", tc.path, tc.want, data)
			}
		})
	}
}
