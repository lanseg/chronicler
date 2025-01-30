package parser

import (
	"bytes"
	"fmt"
	"testing"
)

func TestHtmlReader(t *testing.T) {

	t.Run("simple test", func(t *testing.T) {
		reader := NewHtmlReader(bytes.NewReader(
			[]byte("<tag class=\"classa classb\" other-param=2 param>123</tag>")))
		reader.NextToken()
		if !reader.HasClass("classa") || !reader.HasClass("classb") {
			value, ok := reader.Attr("class")
			t.Errorf("Token expected to have classes \"classA\" and \"classB\", but got %q %v",
				value, ok)
		}
		if value, ok := reader.Attr("other-param"); value != "2" || !ok {
			t.Errorf("other-param should be 2")
		}
		if value, ok := reader.Attr("param"); !ok || value != "" {
			t.Errorf("param expected to be present, but empty and got %q %v", value, ok)
		}

		reader.NextToken()
		if value, ok := reader.Attr("class"); value != "" || ok {
			t.Errorf("Class field expected to be null")
		}
	})

	t.Run("calls before next return empty values and false", func(t *testing.T) {
		reader := NewHtmlReader(bytes.NewReader(
			[]byte("<tag class=\"classa classb\" other-param=2 param>123</tag>")))
		if reader.HasClass("classa") {
			t.Errorf("Class expected to be empty")
		}
	})

	t.Run("call after end return empty values and false", func(t *testing.T) {
		reader := NewHtmlReader(bytes.NewReader(
			[]byte("<tag class=\"classa classb\" other-param=2 param>123</tag>")))
		for reader.NextToken() {

		}
		reader.NextToken()
		fmt.Println(reader.Raw())
	})

	t.Run("malformed html", func(t *testing.T) {
		reader := NewHtmlReader(bytes.NewReader([]byte("<tag class=\"clas < <tag>> <tag></tag>")))
		for reader.NextToken() {
			t.Errorf("Malformed html should return nothing")
		}
	})
}
