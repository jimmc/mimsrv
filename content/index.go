package content

import (
  "fmt"
  "io/ioutil"
  "os"
  "path/filepath"
  "strings"
)

type imageEntry struct {
  filename string
  rotation string
}

type ImageIndex struct {
  indexName string
  entries map[string]*imageEntry
}

/* Reads the image index in the specified directory, or nil
 * if no index file.
 */
func (h *Handler) imageIndex(dir string) *ImageIndex {
  indexName := "index.mpr"
  indexPath := fmt.Sprintf("%s/%s", dir, indexName)
  b, err := ioutil.ReadFile(indexPath)
  if err != nil {
    return nil;
  }

  indexText := string(b)
  indexLines := strings.Split(indexText, "\n")

  entries := make(map[string]*imageEntry)
  for i := range indexLines {
    if indexLines[i] != "" {
      fields := strings.Split(indexLines[i], ";")
      filename := fields[0]
      entry := &imageEntry{
        filename: filename,
      }
      if len(fields) > 1 {
        entry.rotation = fields[1]
      }
      entries[filename] = entry
    }
  }
  return &ImageIndex{
    indexName: indexName,
    entries: entries,
  }
}

func (i *ImageIndex) filter(files []os.FileInfo) []os.FileInfo {
  filteredFiles := make([]os.FileInfo, 0, len(files))
  for _, f := range(files) {
    if i.entries[f.Name()] != nil {
      filteredFiles = append(filteredFiles, f)
    }
  }
  return filteredFiles
}

/* Returns an integer multiple of 90 representing the rotation of the
 * specified image in degrees according to the image index file in the
 * same folder as the specified image. If there is no image file, or
 * the image is not found in the index file, or there is an error reading
 * the index file, zero is returned.
 */
func (h *Handler) rotationFromIndex(imageFilePath string) int {
  base := filepath.Base(imageFilePath)
  dir := filepath.Dir(imageFilePath)
  indexName := "index.mpr"
  indexPath := fmt.Sprintf("%s/%s", dir, indexName)
  b, err := ioutil.ReadFile(indexPath)
  if err != nil {
    // Could be file-not-found, we could check for that and log if something else.
    return 0
  }
  indexText := string(b)
  indexLines := strings.Split(indexText, "\n")

  for i := range indexLines {
    if strings.HasPrefix(indexLines[i], base) {
      // We found our line.
      payload := strings.TrimPrefix(indexLines[i], base)
      payload = strings.TrimPrefix(payload, ";")
      if strings.HasPrefix(payload, "+rr") {
        return 180
      } else if strings.HasPrefix(payload, "+r") {
        return 90
      } else if strings.HasPrefix(payload, "-r") {
        return -90
      } else {
        return 0
      }
    }
  }

  return 0
}
