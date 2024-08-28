package webdriver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type HttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type RequestBuilder[T any] struct {
	url           string
	method        string
	headers       map[string]string
	body          io.Reader
	alreadyFailed error
}

func (rb *RequestBuilder[T]) WithMethod(method string) *RequestBuilder[T] {
	rb.method = method
	return rb
}

func (rb *RequestBuilder[T]) WithHeader(key string, value string) *RequestBuilder[T] {
	rb.headers[key] = value
	return rb
}

func (rb *RequestBuilder[T]) WithBody(body io.Reader) *RequestBuilder[T] {
	rb.body = body
	return rb
}

func (rb *RequestBuilder[T]) WithJsonBody(body interface{}) *RequestBuilder[T] {
	rb.WithHeader("Content-type", "application/json")
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		rb.alreadyFailed = err
		return rb
	}
	return rb.WithBytesBody(jsonBytes)
}

func (rb *RequestBuilder[T]) WithBytesBody(body []byte) *RequestBuilder[T] {
	rb.body = bytes.NewReader(body)
	return rb
}

func (rb *RequestBuilder[T]) Do(client HttpClient) (io.ReadCloser, error) {
	if rb.alreadyFailed != nil {
		return nil, rb.alreadyFailed
	}
	rq, err := http.NewRequest(rb.method, rb.url, rb.body)
	if err != nil {
		return nil, err
	}
	for k, v := range rb.headers {
		rq.Header.Set(k, v)
	}
	result, err := client.Do(rq)
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (rb *RequestBuilder[T]) DoAndRead(client HttpClient) ([]byte, error) {
	reader, err := rb.Do(client)
	if err != nil {
		return nil, err
	}
	// TODO: Consider io.copy maybe
	defer reader.Close()
	return io.ReadAll(reader)
}

func (rb *RequestBuilder[T]) DoAndUnmarshal(client HttpClient) (*T, error) {
	resultBytes, err := rb.DoAndRead(client)
	if err != nil {
		return nil, err
	}
	result := new(T)
	return result, json.Unmarshal(resultBytes, result)
}

func NewTypedRequestBuilder[T any](url string) *RequestBuilder[T] {
	return &RequestBuilder[T]{
		url:     url,
		method:  "GET",
		headers: map[string]string{},
		body:    nil,
	}
}

func NewRequestBuilder(url string) *RequestBuilder[any] {
	return NewTypedRequestBuilder[any](url)
}
