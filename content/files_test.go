package content

import (
  "net/http"
  "testing"
)

func TestListNoIndex(t *testing.T) {
  h := NewHandler(&Config{
    ContentRoot: "testdata",
  });

  list, err, status := h.List("no-such-directory")
  if err == nil {
    t.Errorf("listing non-exstant directory should fail")
  }
  if status != http.StatusNotFound {
    t.Errorf("listing non-existant directory should return NotFound status")
  }

  list, err, status = h.List("d1")
  if err != nil {
    t.Fatalf("failed to list test directory d1")
  }
  if list == nil {
    t.Fatalf("no list returned for directory d1")
  }
  if got, want := list.UnfilteredFileCount, 3; got != want {
    t.Errorf("list d1 unfiltered item count: got %d, want %d", got, want)
  }
  if got, want := len(list.Items), 3; got != want {
    t.Errorf("list d1 item count: got %d, want %d", got, want)
  }
  if got, want := list.Items[0].Name, "image1.jpg"; got != want {
    t.Errorf("first file in d1: got %s, want %s", got, want)
  }
  if got, want := list.Items[0].Text, "sample1\n"; got != want {
    t.Errorf("text for first file in d1: got %s, want %s", got, want)
  }
  if got, want := list.Items[1].Text, ""; got != want {
    t.Errorf("text for second file in d1: got %s, want blank", got)
  }
}

func TestListWithIndex(t *testing.T) {
  h := NewHandler(&Config{
    ContentRoot: "testdata",
  });

  list, err, _ := h.List("with-index")
  if err != nil {
    t.Fatalf("failed to list test directory with-index")
  }
  if got, want := list.UnfilteredFileCount, 7; got != want {
    t.Errorf("list with-index unfiltered item count: got %d, want %d", got, want)
  }
  if got, want := len(list.Items), 5; got != want {
    t.Errorf("list with-index item count: got %d, want %d", got, want)
  }
  if got, want := list.Items[0].Name, "image2.jpg"; got != want {
    t.Errorf("first file in d1: got %s, want %s", got, want)
  }
}

func TestText(t *testing.T) {
  h := NewHandler(&Config{
    ContentRoot: "testdata",
  });

  _, err, status := h.Text("no-such-file")
  if err == nil {
    t.Errorf("querying non-exstant text file should fail")
  }
  if status != http.StatusNotFound {
    t.Errorf("querying non-existant text file should return NotFound status")
  }

  b, err, status := h.Text("d1/image1.txt")
  if err != nil {
    t.Errorf("failed to read text file d1/image1.txt: %v", err)
  }
  if got, want := string(b), "sample1\n"; got != want {
    t.Errorf("reading text file: got %s, want %s", got, want)
  }
}
