package api

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "testing"

  "github.com/jimmc/mimsrv/content"
)

func TestList(t *testing.T) {
  contentHandler := content.NewHandler(&content.Config{
    ContentRoot: "../content/testdata",
  })
  h := handler{
    config: &Config{
      Prefix: "/api/",
      ContentHandler: contentHandler,
    },
  }

  req, err := http.NewRequest("GET", "/api/list/d1", nil)
  if err != nil {
    t.Fatalf("error creating list request: %v", err)
  }

  rr := httptest.NewRecorder()
  handler := http.HandlerFunc(h.list)
  handler.ServeHTTP(rr, req)

  if got, want := rr.Code, http.StatusOK; got != want {
    t.Errorf("list call status: got %d, want %d", got, want)
  }

  body := rr.Body.String()
  if body == "" {
    t.Fatalf("list response body should not be empty")
  }
  var dat map[string]interface{}
  if err = json.Unmarshal(rr.Body.Bytes(), &dat); err != nil {
    t.Fatalf("error unmarshaling list body: %v", err)
  }

  if got, want := dat["IndexName"].(string), ""; got != want {
    t.Errorf("list.IndexName got %s, want %s", got, want)
  }
  if got, want := int(dat["UnfilteredFileCount"].(float64)), 3; got != want {
    t.Errorf("list.UnfilteredFileCount got %d, want %d", got, want)
  }
  items := dat["Items"].([]interface{})
  if got, want := len(items), 3; got != want {
    t.Errorf("list items count got %d, want %d", got, want)
  }
  item0 := items[0].(map[string]interface{})
  if got, want := item0["Name"].(string), "image1.jpg"; got != want {
    t.Errorf("list item0 Name got %s, want %s", got, want)
  }
  if got, want := item0["Text"].(string), "sample1\n"; got != want {
    t.Errorf("list item0 Text got %s, want %s", got, want)
  }
}

func TestText(t *testing.T) {
  contentHandler := content.NewHandler(&content.Config{
    ContentRoot: "../content/testdata",
  })
  h := handler{
    config: &Config{
      Prefix: "/api/",
      ContentHandler: contentHandler,
    },
  }

  req, err := http.NewRequest("GET", "/api/text/d1/image1.txt", nil)
  if err != nil {
    t.Fatalf("error creating text request: %v", err)
  }

  rr := httptest.NewRecorder()
  handler := http.HandlerFunc(h.text)
  handler.ServeHTTP(rr, req)

  if got, want := rr.Code, http.StatusOK; got != want {
    t.Errorf("text call status: got %d, want %d", got, want)
  }

  body := rr.Body.String()
  if body == "" {
    t.Fatalf("text response body should not be empty")
  }
  if got, want := body, "sample1\n"; got != want {
    t.Errorf("text got %s want %s", got, want)
  }
}
