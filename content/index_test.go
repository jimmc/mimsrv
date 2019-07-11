package content

import (
  "bytes"
  "fmt"
  "io"
  "io/ioutil"
  "net/http"
  "os"
  "testing"
)

func TestUpdateImageIndex(t *testing.T) {
  h := NewHandler(&Config{
    ContentRoot: "testdata",
  });

  srcFilename := "testdata/index1.mpr"
  testTmpDir := "testdata/tmp"
  testIndexFilename := testTmpDir + "/index.mpr"
  bakFilename := testIndexFilename + "~"
  golden0Filename := "testdata/index0-golden.mpr"
  goldenFilename := "testdata/index1-golden.mpr"
  golden2Filename := "testdata/index2-golden.mpr"

  os.RemoveAll(testTmpDir)
  err := os.Mkdir(testTmpDir, 0744)
  if err != nil {
    t.Fatalf(err.Error())
  }
  defer os.RemoveAll(testTmpDir)

  command := UpdateCommand{
    Item: "i",
    Action: "a",
    Value: "v",
  }

  err, _ = h.updateImageIndexItem("foo.txt", command)
  if err == nil {
    t.Errorf("updating foo.txt as index file should fail")
  }
  err, _ = h.updateImageIndexItem("foo.mpr", command)
  if err == nil {
    t.Errorf("updating index file other than index.mpr should fail")
  }
  err, _ = h.updateImageIndexItem(testIndexFilename, command)
  if err == nil {
    t.Errorf("updating index file other than index.mpr should fail")
  }
  err, _ = h.updateImageIndexItem(testIndexFilename, UpdateCommand{
    Action: "deltarotation",
    Value: "v",
  })
  if err == nil {
    t.Errorf("blank item should fail")
  }
  err, _ = h.updateImageIndexItem(testIndexFilename, UpdateCommand{
    Item: "i",
    Value: "v",
  })
  if err == nil {
    t.Errorf("blank action should fail")
  }
  err, _ = h.updateImageIndexItem(testIndexFilename, UpdateCommand{
    Item: "i",
    Action: "deltarotation",
  })
  if err == nil {
    t.Errorf("blank value should fail")
  }
  err, _ = h.updateImageIndexItem(testIndexFilename, UpdateCommand{
    Item: "i",
    Action: "deltarotation",
    Value: "+r",
  })
  if err == nil {
    t.Errorf("updating non-existant index without autocreate should fail")
  }
  err, _ = h.updateImageIndexItem(testTmpDir + "/foo.mpr", UpdateCommand{
    Item: "i",
    Action: "deltarotation",
    Value: "+r",
    Autocreate: true,
  })
  if err == nil {
    t.Errorf("updating custom index with autocreate should fail")
  }

  f, err := os.Create(testTmpDir + "/img001.jpg")
  if err != nil {
    t.Errorf(err.Error())
  }
  f.Close()
  f, err = os.Create(testTmpDir + "/img002.jpg")
  if err != nil {
    t.Errorf(err.Error())
  }
  f.Close()
  err, _ = h.updateImageIndexItem(testIndexFilename, UpdateCommand{
    Item: "img001.jpg",
    Action: "deltarotation",
    Value: "+r",
    Autocreate: true,
  })
  if err != nil {
    t.Fatalf("failed to update autocreated index: %v", err)
  }
  err = compareFiles(testIndexFilename, golden0Filename)
  if err != nil {
    t.Error(err.Error())
  }

  err = copyFile(srcFilename, testIndexFilename)
  if err != nil {
    t.Fatal(err.Error())
  }
  err, _ = h.updateImageIndexItem(testIndexFilename, UpdateCommand{
    Item: "nosuchimage.jpg",
    Action: "deltarotation",
    Value: "+r",
  })
  if err == nil {
    t.Errorf("rotate non-existing image should fail")
  }
  err, status := h.updateImageIndexItem(testIndexFilename, UpdateCommand{
    Item: "img001.jpg",
    Action: "deltarotation",
    Value: "+r",
  })
  if err != nil {
    t.Fatalf("rotate existing image failed: %v", err)
  }
  if got, want := status, http.StatusOK; got != want {
    t.Errorf("update index status: got %d, want %d", got, want)
  }

  // Make sure we renamed the old file as a backup
  err = compareFiles(bakFilename, srcFilename)
  if err != nil {
    t.Error(err.Error())
  }
  // Make sure the file we created is correct
  err = compareFiles(testIndexFilename, goldenFilename)
  if err != nil {
    t.Error(err.Error())
  }

  err, _ = h.updateImageIndexItem(testIndexFilename, UpdateCommand{
    Item: "nosuchimage.jpg",
    Action: "drop",
  })
  if err == nil {
    t.Errorf("drop non-existing image should fail")
  }
  err, _ = h.updateImageIndexItem(testIndexFilename, UpdateCommand{
    Item: "img001.jpg",
    Action: "drop",
  })
  if err != nil {
    t.Errorf("Error dropping file img001.jpg: %v", err)
  }
  err = compareFiles(testIndexFilename, golden2Filename)
  if err != nil {
    t.Error(err.Error())
  }
}

func compareFiles(newFilename, refFilename string) error {
  got, err := ioutil.ReadFile(newFilename)
  if err != nil {
    return fmt.Errorf("failed to read test file %s: %v", newFilename, err)
  }
  want, err := ioutil.ReadFile(refFilename)
  if err != nil {
    return fmt.Errorf("failed to read reference file %s: %v", refFilename, err)
  }
  if !bytes.Equal(got, want) {
    return fmt.Errorf("file %s contents: got <<%s>>, want <<%s>>", newFilename, got, want)
  }
  return nil
}

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
    t.Errorf("rotation from index for no-such-file: got %d, want %d", got, want)
  }
  if got, want := h.rotationFromIndex("testdata/with-index/image1.jpg"), 0; got != want {
    t.Errorf("rotation from index for image1: got %d, want %d", got, want)
  }
  if got, want := h.rotationFromIndex("testdata/with-index/image2.jpg"), 0; got != want {
    t.Errorf("rotation from index for image2: got %d, want %d", got, want)
  }
  if got, want := h.rotationFromIndex("testdata/with-index/image4.jpg"), 90; got != want {
    t.Errorf("rotation from index for image5: got %d, want %d", got, want)
  }
  if got, want := h.rotationFromIndex("testdata/with-index/image5.jpg"), 180; got != want {
    t.Errorf("rotation from index for image5: got %d, want %d", got, want)
  }
}

func copyFile(from, to string) error {
  src, err := os.Open(from)
  if err != nil {
    return err
  }
  defer src.Close()
  dst, err := os.Create(to)
  if err != nil {
    return err
  }
  defer dst.Close()
  _, err = io.Copy(dst, src)
  return err
}
