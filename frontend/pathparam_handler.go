package frontend

import (
	"net/http"
	"net/url"
	"strings"
)

type PathParams map[string]string
type PathParamHandlerFunc func(params PathParams, w http.ResponseWriter, r *http.Request)

type PathParamMapper struct {
	pathNames []string

	handler PathParamHandlerFunc
}

func (ppm *PathParamMapper) match(u *url.URL) (map[string]string, bool) {
	pathValues := strings.Split(u.Path, "/")
	if len(pathValues) != len(ppm.pathNames) {
		return map[string]string{}, false
	}

	result := map[string]string{}
	for i, name := range ppm.pathNames {
		l := len(name)
		if l > 2 && name[0] == '{' && name[l-1] == '}' {
			result[name[1:l-1]] = pathValues[i]
			continue
		}
		if name != pathValues[i] {
			return map[string]string{}, false
		}
	}
	return result, true
}

func (ppm *PathParamMapper) Handle(w http.ResponseWriter, r *http.Request) bool {
	if params, ok := ppm.match(r.URL); ok {
		ppm.handler(params, w, r)
		return true
	}
	return false
}

type PathParamHandler struct {
	http.Handler

	mappers     []*PathParamMapper
	elseHandler http.Handler
}

func (pph *PathParamHandler) Handle(path string, handler PathParamHandlerFunc) {
	pph.mappers = append(pph.mappers, &PathParamMapper{
		pathNames: strings.Split(path, "/"),
		handler:   handler,
	})
}

func (pph *PathParamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, mapper := range pph.mappers {
		if mapper.Handle(w, r) {
			return
		}
	}
	if pph.elseHandler != nil {
		pph.elseHandler.ServeHTTP(w, r)
	}
}
