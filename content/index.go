package content

import (
  "fmt"
  "io/ioutil"
  "net/http"
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

const (
  indexExtension = ".mpr"
)

func (h *Handler) UpdateImageIndex(apiPath string, item, action, value string) (error, int) {
  contentRoot := strings.TrimSuffix(h.config.ContentRoot, "/")
  indexPath := fmt.Sprintf("%s/%s", contentRoot, apiPath)
  return updateImageIndexItem(indexPath, item, action, value)
}

func updateImageIndexItem(indexPath string, item, action, value string) (error, int) {
  if filepath.Ext(indexPath) != indexExtension {
    return fmt.Errorf("Index operations can only apply to .%s files, not to %s", indexExtension, indexPath), http.StatusBadRequest
  }
  if filepath.Base(indexPath) != "index.mpr" {
    return fmt.Errorf("Index operations can only apply to index.mpr files"), http.StatusBadRequest
  }
  if action == "" {
    return fmt.Errorf("No action specified"), http.StatusBadRequest
  }
  if action != "deltarotation" {
    return fmt.Errorf("Action %s is not valid", action), http.StatusBadRequest
  }
  if item == "" {
    return fmt.Errorf("No item specified"), http.StatusBadRequest
  }
  if value == "" {
    return fmt.Errorf("No value specified"), http.StatusBadRequest
  }

  lines, err := readFileLines(indexPath)
  if err != nil {
    return err, http.StatusInternalServerError
  }

  itemIndex, entry := findEntry(lines, item)
  if itemIndex < 0 {
    return fmt.Errorf("Item %s not found in index", item), http.StatusBadRequest
  }
  if action == "deltarotation" {
    entry.rotation, err = combineRotations(entry.rotation, value)
    if err != nil {
      return err, http.StatusBadRequest
    }
    lines[itemIndex] = entry.toString()
    err = backupAndWriteFileLines(indexPath, lines)
    if err != nil {
      return err, http.StatusInternalServerError
    }
    return nil, http.StatusOK
  }
  // Add other actions here when defined.

  return nil, http.StatusNotImplemented
}

func combineRotations(fileRotation, deltaRotation string) (string, error) {
  if fileRotation != "" && fileRotation != "+r" && fileRotation != "++r" && fileRotation != "-r" {
    return "", fmt.Errorf("Rotation %s in file is not valid", fileRotation)
  }
  if deltaRotation != "" && deltaRotation != "+r" && deltaRotation != "++r" && deltaRotation != "-r" {
    return "", fmt.Errorf("Rotation %s in file is not valid", deltaRotation)
  }
  switch fileRotation + deltaRotation {
    case "": return "", nil
    case "+r": return "+r", nil
    case "+rr": return "+rr", nil
    case "-r": return "-r", nil
    case "+r+r": return "+rr", nil
    case "+r+rr": return "-r", nil
    case "+r-r": return "", nil
    case "+rr+r": return "-r", nil
    case "+rr+rr": return "", nil
    case "+rr-r": return "+r", nil
    case "-r+r": return "", nil
    case "-r+rr": return "+r", nil
    case "-r-r": return "+rr", nil
    default: return "", nil     // can't happen
  }
}

func findEntry(lines []string, item string) (int, *imageEntry) {
  // Look for the matching line
  for i, line := range lines {
    if line != "" {
      entry := entryFromLine(line)
      if entry.filename == item {
        return i, entry
      }
    }
  }
  return -1, nil
}

/* Reads the image index in the specified directory, or nil
 * if no index file.
 */
func (h *Handler) imageIndex(dir string) *ImageIndex {
  indexName := "index.mpr"
  indexPath := fmt.Sprintf("%s/%s", dir, indexName)
  indexLines, err := readFileLines(indexPath)
  if err != nil {
    return nil
  }

  entries := make(map[string]*imageEntry)
  for i := range indexLines {
    if indexLines[i] != "" {
      entry := entryFromLine(indexLines[i])
      entries[entry.filename] = entry
    }
  }
  return &ImageIndex{
    indexName: indexName,
    entries: entries,
  }
}

func readFileLines(filename string) ([]string, error) {
  b, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  text := string(b)
  lines := strings.Split(text, "\n")
  return lines, nil
}

func entryFromLine(line string) *imageEntry {
  fields := strings.Split(line, ";")
  entry := &imageEntry{
    filename: fields[0],
  }
  if len(fields) > 1 {
    entry.rotation = fields[1]
  }
  return entry
}

func backupAndWriteFileLines(filename string, lines []string) error {
  newFilename := filename + ".new"
  err := writeFileLines(newFilename, lines)
  if err != nil {
    return fmt.Errorf("error writing new index file: %v", err)
  }
  backupFilename := filename + "~"
  err = os.Rename(filename, backupFilename)
  if err != nil {
    return fmt.Errorf("error renaming old index file to backup: %v", err)
  }
  err = os.Rename(newFilename, filename)
  if err != nil {
    return fmt.Errorf("error renaming new index file: %v", err)
  }
  return nil
}

func writeFileLines(filename string, lines []string) error {
  f, err := os.Create(filename)
  if err != nil {
    return err
  }

  for _, line := range lines {
    if line != "" {
      fmt.Fprintf(f, "%s\n", line)
    }
  }
  return f.Close()
}


func (e *imageEntry) toString() string {
  s := e.filename
  if e.rotation != "" {
    s = s + ";" + e.rotation
  }
  return s
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
