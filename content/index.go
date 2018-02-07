package content

import (
  "fmt"
  "io/ioutil"
  "path/filepath"
  "strings"
)

type ImageIndex struct {
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
  indexPath := fmt.Sprintf("%s/index.mpr", dir)
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
