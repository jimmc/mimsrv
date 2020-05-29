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
    t.Fatalf("failed to autocreate index: %v", err)
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
    return fmt.Errorf("file %s vs golden file %s: got <<%s>>, want <<%s>>", newFilename, refFilename, got, want)
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
  if got, want := len(index.entries), 9; got != want {
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
  testCases := []struct{
    filename string
    exifRotation int
    wantResult int
  }{
    { "testdata/with-index/no-such-file.jpg", 0, 0 },   // No index file
    { "testdata/with-index/no-such-file.jpg", 90, 90 },
    { "testdata/with-index/image1.jpg", 0, 0 },         // Not in index file
    { "testdata/with-index/image1.jpg", 90, 90 },
    { "testdata/with-index/image2.jpg", 0, 0 },         // In index file with no rotation
    { "testdata/with-index/image2.jpg", 90, 0 },
    { "testdata/with-index/image4.jpg", 0, 90 },        // In index file with +r
    { "testdata/with-index/image4.jpg", 90, 90 },
    { "testdata/with-index/image5.jpg", 0, 180 },       // In index file with +rr
    { "testdata/with-index/image5.jpg", 90, 180 },
    { "testdata/with-index/image6.jpg", 0, -90 },       // In index file with -r
    { "testdata/with-index/image6.jpg", 90, -90 },
    { "testdata/with-index/xo.jpg", 0, 0 },
    { "testdata/with-index/xo.jpg", 90, 90 },
    { "testdata/with-index/xo+rr.jpg", 0, 180 },
    { "testdata/with-index/xo+rr.jpg", 180, 360 },
    { "testdata/with-index/xo+r.jpg", 0, 90 },
    { "testdata/with-index/xo+r.jpg", 90, 180 },
    { "testdata/with-index/xo-r.jpg", 0, -90 },
    { "testdata/with-index/xo-r.jpg", 90, 0 },
  }

  h := NewHandler(&Config{
    ContentRoot: "testdata",
  });

  for _, test := range testCases {
    name := fmt.Sprintf("rotation(%v, %v)", test.filename, test.exifRotation)
    t.Run(name, func(t *testing.T) {
      if got, want := h.rotationFromIndexAndExif(test.filename, test.exifRotation), test.wantResult; got != want {
        t.Errorf("rotationFromIndexAndExif: got %d, want %d", got, want)
      }
    })
  }
}

func TestCombineRotations(t *testing.T) {
  testCases := []struct{
    name string
    fileRotation string
    deltaRotation string
    wantResult string
    wantErr bool
  } {
    { "bad file rot", "invalid", "", "", true },
    { "bad delta rot", "", "invalid", "", true },
    { "no rotations", "", "", "", false },
    { "file rot, no delta", "+r", "", "+r", false },
    { "file rot with delta", "+r", "+r", "+rr", false },
    { "xo, no delta", "xo", "", "xo", false },
    { "xo with delta", "xo", "-r", "xo-r", false },
    { "xo and file rot, no delta", "xo+rr", "", "xo+rr", false },
    { "xo and file rot with delta", "xo+rr", "-r", "xo+r", false },
  }

  for _, test := range testCases {
    t.Run(test.name, func(t *testing.T) {
      got, err := combineRotations(test.fileRotation, test.deltaRotation)
      gotErr := (err != nil)
      if gotErr != test.wantErr {
        if test.wantErr {
          t.Fatalf("did not get error as expected")
        } else {
          t.Fatalf("error combining rotations: %v", err)
        }
      }
      want := test.wantResult
      if got != want {
        t.Errorf("combine: got %s, want %s", got, want)
      }
    })
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
