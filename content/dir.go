package content

import (
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "path"
  "path/filepath"
  "sort"
  "strings"
)

// dirFlags is the set of flags about the directory that can be stored
// in summary.txt in initial bang lines.
type dirFlags struct {
  ignoreFileTimes bool
}

func (h *Handler) readDirFiltered(dirPath string) ([]os.FileInfo, error, int) {
  f, err := os.Open(dirPath)
  if err != nil {
    return nil, fmt.Errorf("failed to open file: %v", err), http.StatusNotFound
  }
  defer f.Close()
  files, err := f.Readdir(0)       // Read all file names
  if err != nil {
    return nil, fmt.Errorf("failed to read dir: %v", err), http.StatusBadRequest
  }
  files = h.filterOnExtension(dirPath, files)
  sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
  return files, nil, 0
}

func (h *Handler) filterOnExtension(dirPath string, files []os.FileInfo) []os.FileInfo {
  filteredFiles := make([]os.FileInfo, 0, len(files))
  i := 0
  for _, f := range files {
    if h.keepFileInList(dirPath, f) {
      filteredFiles = filteredFiles[:i+1]
      filteredFiles[i] = f
      i = i + 1
    }
  }
  return filteredFiles
}

func (h *Handler) keepFileInList(dirPath string, f os.FileInfo) bool {
  isDir := f.IsDir() || isSymlinkToDir(dirPath, f)
  if isDir {
    // Don't display hidden dirs, in particular our cache dir
    if strings.HasPrefix(f.Name(), ".") {
      return false;
    }
    return true;
  }
  ext := strings.ToLower(filepath.Ext(f.Name()))
  if h.imageExts[ext] || h.videoExts[ext] {
    return true;
  }
  return false;
}

// SymlinkPointsToDirectory returns true if f refers to a symlink
// and that symlink points to a directory. If any errors, returns false.
func isSymlinkToDir(dirPath string, f os.FileInfo) bool {
  filemode := f.Mode()
  if filemode & os.ModeSymlink == 0 {
    return false        // Not a symlink
  }
  fPath := path.Join(dirPath, f.Name())
  dest, err := os.Readlink(fPath)
  log.Printf("os.Readlink on %s returns %s, err=%v", fPath, dest, err)
  if err != nil {
    return false
  }
  if !path.IsAbs(dest) {
    dest = path.Join(dirPath, dest)
  }
  log.Printf("dest to Lstat is %s", dest)
  ff, err := os.Lstat(dest)
  if err != nil {
    log.Printf("Lstat error %v", err)
    return false
  }
  log.Printf("IsDir on Lstat result is %v", ff.IsDir())
  return ff.IsDir()
}

func loadDirFlags(dirPath string) dirFlags {
  summarypath := fmt.Sprintf("%s/summary.txt", dirPath)
  b, err := ioutil.ReadFile(summarypath)
  if err != nil {
    // We ignore the error if it is that the file does not exist
    if !os.IsNotExist(err) {
      log.Printf("Error reading %s: %v", summarypath, err)
    }
    return parseDirFlags("")
  }
  text := string(b)
  return parseDirFlags(text)
}

func parseDirFlags(text string) dirFlags {
  lines := strings.Split(text, "\n")
  flags := dirFlags{}
  for _, line := range lines {
    if !strings.HasPrefix(line, "!") {
      return flags
    }
    line = strings.TrimPrefix(line, "!")
    if line == "ignoreFileTimes" {
      flags.ignoreFileTimes = true
    }
  }
  return flags
}
