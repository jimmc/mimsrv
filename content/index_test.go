package content

import (
  "testing"
)

func TestLoad(t *testing.T) {
  h := NewHandler(&Config{
    ContentRoot: "testdata",
  });

  index := h.imageIndex("no-such-directory")
  if index != nil {
    t.Errorf("reading index for non-existent directory should fail")
  }

  index = h.imageIndex("testdata/with-index")
  if index == nil {
    t.Fatalf("index should have been loaded")
  }
  if got, want := index.indexName, "index.mpr"; got != want {
    t.Errorf("index file name: got %s, want %s", got, want)
  }
  if got, want := len(index.entries), 5; got != want {
    t.Fatalf("index entries count: got %d, want %d", got, want)
  }
  if index.entries["image1.jpg"] != nil {
    t.Errorf("index should not include image1.jpg")
  }
  if got, want := index.entries["image2.jpg"].filename, "image2.jpg"; got != want {
    t.Errorf("name for image2 in index: got %s, want %s", got, want)
  }
  if got, want := index.entries["image2.jpg"].rotation, ""; got != want {
    t.Errorf("rotation for image2 in index: got %s, want %s", got, want)
  }
  if got, want := index.entries["image5.jpg"].rotation, "+rr"; got != want {
    t.Errorf("rotation for image5 in index: got %s, want %s", got, want)
  }
}

func TestRotationFromIndex(t *testing.T) {
  h := NewHandler(&Config{
    ContentRoot: "testdata",
  });

  if got, want := h.rotationFromIndex("testdata/with-index/no-such-file.jpg"), 0; got != want {
    t.Errorf("rotation from index for no-such-file: got %d, want %d")
  }
  if got, want := h.rotationFromIndex("testdata/with-index/image1.jpg"), 0; got != want {
    t.Errorf("rotation from index for image1: got %d, want %d")
  }
  if got, want := h.rotationFromIndex("testdata/with-index/image2.jpg"), 0; got != want {
    t.Errorf("rotation from index for image2: got %d, want %d")
  }
  if got, want := h.rotationFromIndex("testdata/with-index/image4.jpg"), 90; got != want {
    t.Errorf("rotation from index for image5: got %d, want %d")
  }
  if got, want := h.rotationFromIndex("testdata/with-index/image5.jpg"), 180; got != want {
    t.Errorf("rotation from index for image5: got %d, want %d")
  }
}
