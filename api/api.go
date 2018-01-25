package api

import (
  "encoding/json"
  "fmt"
  "image"
  "image/color"
  "image/jpeg"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strconv"
  "strings"

  "github.com/disintegration/imaging"
)

type Config struct {
  Prefix string         // The path prefix being routed to this handler
  ContentRoot string    // The root directory of our photo hierarchy
}

type ListItem struct {
  Name string
}

type ListResult struct {
  Items []ListItem
}

type handler struct {
  config *Config
}

func NewHandler(c *Config) http.Handler {
  h := handler{config: c}
  mux := http.NewServeMux()
  mux.HandleFunc(h.apiPrefix("list"), h.list)
  mux.HandleFunc(h.apiPrefix("image"), h.image)
  mux.HandleFunc(h.apiPrefix("text"), h.text)
  return mux
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
  path := strings.TrimPrefix(r.URL.Path, h.apiPrefix("list"))
  if strings.HasPrefix(path, "..") || strings.Contains(path, "/..") {
    // Hmmm, we never get here, someone upstream from us is collapsing them
    http.Error(w, "Relative paths are not allowed", http.StatusForbidden)
    return
  }

  filepath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  f, err := os.Open(filepath)
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to open file: %v", err), http.StatusNotFound)
    return
  }
  defer f.Close()
  files, err := f.Readdir(0)       // Read all file names
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to read dir: %v", err), http.StatusBadRequest)
    return
  }

  result := mapFileInfosToListResult(files)

  b, err := json.MarshalIndent(result, "", "  ")
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to create json dir: %v", err), http.StatusInternalServerError)
    return
  }
  w.WriteHeader(http.StatusOK)
  w.Write(b)
}

func mapFileInfosToListResult(files []os.FileInfo) ListResult {
  n := len(files)
  list := make([]ListItem, n, n)
  for i, f := range files {
    mapFileInfoToListItem(f, &list[i])
  }
  return ListResult{
    Items: list,
  }
}

func mapFileInfoToListItem(f os.FileInfo, item *ListItem) {
  item.Name = f.Name()
}

func (h *handler) image(w http.ResponseWriter, r *http.Request) {
  path := strings.TrimPrefix(r.URL.Path, h.apiPrefix("image"))
  filepath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  f, err := os.Open(filepath)
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to open file: %v", err), http.StatusNotFound)
    return
  }
  defer f.Close()

  im, _, err := image.Decode(f)
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to decode image file: %v", err), http.StatusBadRequest)
    return
  }

  var width int
  widthStr := r.FormValue("w")
  if widthStr == "" {
    width = 0
  } else {
    width, err = strconv.Atoi(widthStr)
    if err != nil {
      http.Error(w, "Bad value for w parameter", http.StatusBadRequest)
    }
  }

  var height int
  heightStr := r.FormValue("h")
  if heightStr == "" {
    height = 0
  } else {
    height, err = strconv.Atoi(heightStr)
    if err != nil {
      http.Error(w, "Bad value for h parameter", http.StatusBadRequest)
    }
  }

  var rot int
  rotStr := r.FormValue("r")
  if rotStr == "" {
    rot = 0
  } else {
    rot, err = strconv.Atoi(rotStr)
    if err != nil {
      http.Error(w, "Bad value for r parameter", http.StatusBadRequest)
    }
  }

  imRect := im.Bounds()
  imWidth := imRect.Max.X - imRect.Min.X
  imHeight := imRect.Max.Y - imRect.Min.Y
  log.Printf("image size: w=%d, h=%d; resize parameters w=%d, h=%d",
      imWidth, imHeight, width, height)
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

  w.Header().Set("Content-Type", "image/jpeg")
  w.WriteHeader(http.StatusOK)
  options := &jpeg.Options{
    Quality: 90,
  }
  jpeg.Encode(w, im, options)
}

func (h *handler) text(w http.ResponseWriter, r *http.Request) {
  path := strings.TrimPrefix(r.URL.Path, h.apiPrefix("text"))
  filepath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  b, err := ioutil.ReadFile(filepath)
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusNotFound)
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write(b)
}

func (h *handler) apiPrefix(s string) string {
  return fmt.Sprintf("%s%s/", h.config.Prefix, s)
}
