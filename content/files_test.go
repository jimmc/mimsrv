package content

import (
  "io/ioutil"
  "net/http"
  "os"
  "testing"
)

func TestListNoIndex(t *testing.T) {
  h := NewHandler(&Config{
    ContentRoot: "testdata",
  });

  list, err, status := h.List("no-such-directory")
  if err == nil {
    t.Errorf("listing non-existant directory should fail")
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
  if got, want := list.UnfilteredFileCount, 8; got != want {
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

  _, err, _ := h.Text("bad-extension.not-txt")
  if err == nil {
    t.Errorf("querying text file with extension other than .txt should fail")
  }

  _, err, status := h.Text("no-such-file.txt")
  if err == nil {
    t.Errorf("querying non-existant text file should fail")
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

func TestPutText(t *testing.T) {
  testDir := "testdata/tmp"
  h := NewHandler(&Config{
    ContentRoot: testDir,
  });

  err := os.Mkdir(testDir, 0744)
  if err != nil {
    t.Fatalf("Unable to create test directory: %v", err)
  }
  defer os.RemoveAll(testDir)

  err, _ = h.PutText("test1.not-txt", UpdateTextCommand{
    Content: "hello",
  })
  if err == nil {
    t.Errorf("writing text file with extension other than .txt should fail")
  }

  err, status := h.PutText("test1.txt", UpdateTextCommand{
    Content: "hello",
  })
  if err != nil {
    t.Fatalf("failed to write file: %v", err)
  }
  if got, want := status, http.StatusOK; got != want {
    t.Errorf("PutText status for test1: got %d, want %d", got, want)
  }
  b, err := ioutil.ReadFile(testDir + "/test1.txt")
  if err != nil {
    t.Fatalf("Failed to read back text file: %v", err)
  }
  if got, want := string(b), "hello"; got != want {
    t.Errorf("test1.txt contents: got %s, want %s", got, want)
  }

  err, status = h.PutText("test1.txt", UpdateTextCommand{
    Content: "goodbye",
  })
  if err != nil {
    t.Fatalf("failed to rewrite file: %v", err)
  }
  b, err = ioutil.ReadFile(testDir + "/test1.txt")
  if err != nil {
    t.Fatalf("Failed to read back text file: %v", err)
  }
  if got, want := string(b), "goodbye"; got != want {
    t.Errorf("test1.txt contents: got %s, want %s", got, want)
  }

  // Test writing empty file over previous contents
  err, status = h.PutText("test1.txt", UpdateTextCommand{
    Content: "",
  })
  if err != nil {
    t.Fatalf("failed to write empty file: %v", err)
  }
  b, err = ioutil.ReadFile(testDir + "/test1.txt")
  if !os.IsNotExist(err) {
    t.Fatalf("Expected file-not-exist after writing empty content, got %v", err)
  }

  // Test writing empty file when already empty
  err, status = h.PutText("test1.txt", UpdateTextCommand{
    Content: "",
  })
  if err != nil {
    t.Fatalf("failed to rewrite empty file: %v", err)
  }
  b, err = ioutil.ReadFile(testDir + "/test1.txt")
  if !os.IsNotExist(err) {
    t.Fatalf("Expected file-not-exist after rewriting empty content, got %v", err)
  }
}
