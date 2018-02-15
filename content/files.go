package content

import (
  "fmt"
  "image"
  "image/color"
  _ "image/jpeg"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "path/filepath"
  "sort"
  "strings"
  "time"

  "github.com/disintegration/imaging"
)

const (
  timeFormat = "3:04:05pm Mon Jan 2, 2006 MST"
)

type Config struct {
  ContentRoot string    // The root directory of our content hierarchy
}

type Handler struct {
  config *Config
  validExts map[string]bool
}

type ListItem struct {
  Name string
  IsDir bool
  Size int64
  ModTime int64          // seconds since the epoch
  ModTimeStr string      // ModTime converted to a string by the server
  Text string
  TextError string       // The error if we get one trying to read the text file
}

type ListResult struct {
  Items []ListItem
}

func NewHandler(c *Config) Handler {
  h := Handler{config: c}
  h.init()
  return h
}

func (h *Handler) init() {
  h.validExts = map[string]bool {
    ".gif": true,
    ".jpg": true,
  }
}

func (h *Handler) List(dirApiPath string) (*ListResult, error, int) {
  contentRoot := strings.TrimSuffix(h.config.ContentRoot, "/")
  dirApiPath = strings.TrimSuffix(dirApiPath, "/")
  dirPath := fmt.Sprintf("%s/%s", contentRoot, dirApiPath)
  f, err := os.Open(dirPath)
  if err != nil {
    return nil, fmt.Errorf("failed to open file: %v", err), http.StatusNotFound
  }
  defer f.Close()
  files, err := f.Readdir(0)       // Read all file names
  if err != nil {
    return nil, fmt.Errorf("failed to read dir: %v", err), http.StatusBadRequest
  }
  sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

  var loc *time.Location
  tzpath := fmt.Sprintf("%s/TZ", dirPath)
  linkdest, err := os.Readlink(tzpath)
  if err != nil {
    if !os.IsNotExist(err) {
      log.Printf("Error reading TZ symlink %s: %v", tzpath, err)
    }
  } else {
    tzname := strings.TrimPrefix(linkdest, "/usr/share/zoneinfo/")
    loc, err = time.LoadLocation(tzname)
    if err != nil {
      log.Printf("Error loading timezone file %s: %v", tzpath, err)
    }
  }

  result := h.mapFileInfosToListResult(files, dirPath, loc)
  return result, nil, 0
}

func (h *Handler) mapFileInfosToListResult(files []os.FileInfo, parentPath string, loc *time.Location) *ListResult {
  n := len(files)
  list := make([]ListItem, 0, n)
  i := 0
  for _, f := range files {
    if h.keepFileInList(f) {
      list = list[:i+1]
      h.mapFileInfoToListItem(f, &list[i], parentPath, loc)
      i = i + 1
    }
  }
  return &ListResult{
    Items: list,
  }
}

func (h *Handler) keepFileInList(f os.FileInfo) bool {
  if f.IsDir() {
    return true;
  }
  ext := strings.ToLower(filepath.Ext(f.Name()))
  if h.validExts[ext] {
    return true;
  }
  return false;
}

func (h *Handler) mapFileInfoToListItem(f os.FileInfo, item *ListItem, parentPath string, loc *time.Location) {
  item.Name = f.Name()
  item.IsDir = f.IsDir()
  item.Size = f.Size()
  item.ModTime = f.ModTime().Unix()
  t := f.ModTime()
  if loc != nil {
    t = t.In(loc)
  }
  item.ModTimeStr = t.Format(timeFormat)
  h.loadTextFile(item, parentPath)
}

func (h *Handler) loadTextFile(item *ListItem, parentPath string) {
  var textpath string
  if item.IsDir {
    textpath = fmt.Sprintf("%s/%s/summary.txt", parentPath, item.Name)
  } else {
    textname := fmt.Sprintf("%s.txt", strings.TrimSuffix(item.Name, filepath.Ext(item.Name)))
    textpath = fmt.Sprintf("%s/%s", parentPath, textname)
  }
  b, err := ioutil.ReadFile(textpath)
  if err != nil {
    // We ignore the error if it is that the file does not exist
    if !os.IsNotExist(err) {
      item.TextError = fmt.Sprintf("%v", err)
    }
  } else {
    item.Text = string(b)
  }
}

func (h *Handler) Image(path string, width, height, rot int) (image.Image, error, int) {
  imageFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  f, err := os.Open(imageFilePath)
  if err != nil {
    return nil, fmt.Errorf("failed to open file: %v", err), http.StatusNotFound
  }
  defer f.Close()

  im, _, err := image.Decode(f)
  if err != nil {
    return nil, fmt.Errorf("failed to decode image file: %v", err), http.StatusBadRequest
  }

  rot = rot + h.rotationFromIndex(imageFilePath)
  if ((rot + 360) / 90) % 2 == 1 {
    width, height = height, width
  }

  imRect := im.Bounds()
  imWidth := imRect.Max.X - imRect.Min.X
  imHeight := imRect.Max.Y - imRect.Min.Y
  // log.Printf("image size: w=%d, h=%d; resize parameters w=%d, h=%d", imWidth, imHeight, width, height)
  if width != 0 || height != 0 {
    if width != 0 && height != 0 {
      // We always want to preserve the aspect ratio, but fit the resulting
      // image into a bounding box of the specified size.
      if float64(width) / float64(imWidth) > float64(height) / float64(imHeight) {
        width = 0
      } else {
        height = 0
      }
    }
    im = imaging.Resize(im, width, height, imaging.NearestNeighbor)
  }

  if rot != 0 {
    im = imaging.Rotate(im, float64(rot), color.Black)
  }

  return im, nil, 0
}

func (h *Handler) Text(path string) ([]byte, error, int) {
  textFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  b, err := ioutil.ReadFile(textFilePath)
  if err != nil {
    return nil, fmt.Errorf("failed to read file: %v", err), http.StatusNotFound
  }
  return b, nil, 0
}
