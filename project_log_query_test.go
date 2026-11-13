package envai

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestProjectLogListQueryEncoding(t *testing.T) {
	c := testClient(t, func(http.ResponseWriter, *http.Request) {})
	op := operation{ID: "getProjectLogs", Path: "/projects/p/logs", Parameters: []parameter{
		{Name: "types[]", In: "query", Repeated: true},
		{Name: "workspace_ids[]", In: "query", Repeated: true},
		{Name: "api_key_ids[]", In: "query", Repeated: true},
		{Name: "page", In: "query"},
	}}
	params := map[string]any{"types[]": []string{"first", "second"}, "workspace_ids[]": []string{"one", "two"}, "api_key_ids[]": []string{"key"}, "page": 3}
	raw, err := c.requestURL(op, params)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := url.Values{"types[]": {"first", "second"}, "workspace_ids[]": {"one", "two"}, "api_key_ids[]": {"key"}, "page": {"3"}}
	if !reflect.DeepEqual(u.Query(), want) {
		t.Fatalf("query = %v, want %v", u.Query(), want)
	}
	params["types[]"] = "wrong"
	if _, err := c.requestURL(op, params); err == nil {
		t.Fatal("accepted scalar list parameter")
	}
	params["types[]"] = []string{}
	if _, err := c.requestURL(op, params); err == nil {
		t.Fatal("accepted empty list parameter")
	}
	params["types[]"] = nil
	if _, err := c.requestURL(op, params); err == nil {
		t.Fatal("accepted nil list parameter")
	}
}
