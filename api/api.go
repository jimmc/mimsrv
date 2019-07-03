package api

import (
  "encoding/json"
  "fmt"
  "image/jpeg"
  "net/http"
  "strconv"
  "strings"

  "github.com/jimmc/mimsrv/auth"
  "github.com/jimmc/mimsrv/content"
  "github.com/jimmc/mimsrv/permissions"
)

type Config struct {
  Prefix string         // The path prefix being routed to this handler
  ContentHandler content.Handler
}

type handler struct {
  config *Config
  validExts map[string]bool
}

func NewHandler(c *Config) http.Handler {
  h := handler{config: c}
  mux := http.NewServeMux()
  mux.HandleFunc(h.apiPrefix("list"), h.list)
  mux.HandleFunc(h.apiPrefix("image"), h.image)
  mux.HandleFunc(h.apiPrefix("video"), h.video)
  mux.HandleFunc(h.apiPrefix("index"), h.index)
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

  var result *content.ListResult
  var err error
  var status int
  if strings.HasSuffix(path, ".mpr") {
    result, err, status = h.config.ContentHandler.ListFromIndex(path)
  } else {
    result, err, status = h.config.ContentHandler.List(path)
  }
  if err != nil {
    http.Error(w, err.Error(), status)
    return
  }

  b, err := json.MarshalIndent(result, "", "  ")
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to create json dir: %v", err), http.StatusInternalServerError)
    return
  }
  w.WriteHeader(http.StatusOK)
  w.Write(b)
}

func (h *handler) image(w http.ResponseWriter, r *http.Request) {
  path := strings.TrimPrefix(r.URL.Path, h.apiPrefix("image"))

  width, err := formParamInt(r, "w")
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  height, err := formParamInt(r, "h")
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  rot, err := formParamInt(r, "r")
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  im, err, status := h.config.ContentHandler.Image(path, width, height, rot)
  if err != nil {
    http.Error(w, err.Error(), status)
    return
  }

  w.Header().Set("Content-Type", "image/jpeg")
  w.WriteHeader(http.StatusOK)
  options := &jpeg.Options{
    Quality: 90,
  }
  jpeg.Encode(w, im, options)
}

func (h *handler) video(w http.ResponseWriter, r *http.Request) {
  path := strings.TrimPrefix(r.URL.Path, h.apiPrefix("video"))
  videoFilePath, err := h.config.ContentHandler.VideoFilePath(path)
  if err != nil {
    http.Error(w, "Error on video file", http.StatusInternalServerError)
    return
  }
  if videoFilePath == "" {
    http.Error(w, "Not a video file", http.StatusBadRequest)
    return
  }
  http.ServeFile(w, r, videoFilePath)
}

func (h *handler) index(w http.ResponseWriter, r *http.Request) {
  if !auth.CurrentUserHasPermission(r, permissions.CanEdit) {
    http.Error(w, "Not authorized to edit", http.StatusUnauthorized)
    return
  }
  if r.Method != http.MethodPost {
    http.Error(w, "POST method is required", http.StatusMethodNotAllowed)
    return
  }
  apiPath := strings.TrimPrefix(r.URL.Path, h.apiPrefix("index"))

  item := r.FormValue("item")   // name of the index item on which to operate
  action := r.FormValue("action") // action to take on an index item
  value := r.FormValue("value")  // value that goes with the action
  autocreateStr := r.FormValue("autocreate")
  autocreate := false
  if strings.ToLower(autocreateStr) == "true" {
    autocreate = true;
  }

  command := content.UpdateCommand{
    Item: item,
    Action: action,
    Value: value,
    Autocreate: autocreate,
  }
  err, status := h.config.ContentHandler.UpdateImageIndex(apiPath, command)
  if err != nil {
    http.Error(w, err.Error(), status)
    return
  }
  w.WriteHeader(http.StatusOK)
  w.Write([]byte(`{"status": "ok"}`))
}

func (h *handler) text(w http.ResponseWriter, r *http.Request) {
  path := strings.TrimPrefix(r.URL.Path, h.apiPrefix("text"))
  switch r.Method {
    case http.MethodGet:
      b, err, status := h.config.ContentHandler.Text(path)
      if err != nil {
        http.Error(w, err.Error(), status)
        return
      }
      w.WriteHeader(http.StatusOK)
      w.Write(b)
      return
    case http.MethodPut:
      if !auth.CurrentUserHasPermission(r, permissions.CanEdit) {
        http.Error(w, "Not authorized to edit", http.StatusUnauthorized)
        return
      }
      cmd := content.UpdateTextCommand{
        Content: r.FormValue("content"),
      }
      err, status := h.config.ContentHandler.PutText(path, cmd)
      if err != nil {
        http.Error(w, err.Error(), status)
        return
      }
      w.WriteHeader(http.StatusOK)
      w.Write([]byte(`{"status": "ok"}`))
      return
    default:
      http.Error(w, "Method must be GET or PUT", http.StatusMethodNotAllowed)
      return
  }
}

func (h *handler) apiPrefix(s string) string {
  return fmt.Sprintf("%s%s/", h.config.Prefix, s)
}

func formParamInt(r *http.Request, name string) (int, error) {
  strVal := r.FormValue(name)
  if strVal == "" {
    return 0, nil
  }
  intVal, err := strconv.Atoi(strVal)
  if err != nil {
    return 0, fmt.Errorf("bad value for %s parameter", name)
  }
  return intVal, nil
}
