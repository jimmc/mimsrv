package content

import (
  "bytes"
  "fmt"
  "image"
  "image/color"
  _ "image/jpeg"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "os/exec"
  "path/filepath"
  "strings"
  "time"

  "github.com/disintegration/imaging"
)

const (
  textExtension = ".txt"
  timeFormat = "3:04:05pm Mon Jan 2, 2006 MST"
  cacheDir = ".mimcache/"
)

type Config struct {
  ContentRoot string    // The root directory of our content hierarchy
}

type Handler struct {
  config *Config
  imageExts map[string]bool
  videoExts map[string]bool
}

type ListItem struct {
  Name string
  IsDir bool
  Size int64
  Type string
  ModTime int64          // seconds since the epoch
  ModTimeStr string      // ModTime converted to a string by the server
  Text string
  TextError string       // The error if we get one trying to read the text file
}

type ListResult struct {
  IndexName string
  UnfilteredFileCount int
  Items []ListItem
}

type UpdateTextCommand struct {
  Content string
}

func NewHandler(c *Config) Handler {
  h := Handler{config: c}
  h.init()
  return h
}

func (h *Handler) init() {
  h.imageExts = map[string]bool {
    ".gif": true,
    ".jpeg": true,
    ".jpg": true,
    ".png": true,
  }
  h.videoExts = map[string]bool {
    ".mp4": true,
    ".mpg": true,
    ".mts": true,
  }
}

func (h *Handler) List(dirApiPath string) (*ListResult, error, int) {
  contentRoot := strings.TrimSuffix(h.config.ContentRoot, "/")
  dirApiPath = strings.TrimSuffix(dirApiPath, "/")
  dirPath := fmt.Sprintf("%s/%s", contentRoot, dirApiPath)

  files, err, status := h.readDirFiltered(dirPath)
  if err != nil {
    return nil, err, status
  }

  imageIndex := h.imageIndex(dirPath)
  unfilteredFileCount := len(files)
  if imageIndex != nil {
    files = imageIndex.filter(files)
  }

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

  flags := loadDirFlags(dirPath)

  result := h.mapFileInfosToListResult(files, dirPath, loc, flags.ignoreFileTimes)
  result.UnfilteredFileCount = unfilteredFileCount
  if imageIndex != nil {
    result.IndexName = imageIndex.indexName
  }
  return result, nil, 0
}

func (h *Handler) mapFileInfosToListResult(files []os.FileInfo, parentPath string,
    loc *time.Location, ignoreFileTimes bool) *ListResult {
  n := len(files)
  list := make([]ListItem, n, n)
  for i, f := range files {
    h.mapFileInfoToListItem(f, &list[i], parentPath, loc, ignoreFileTimes)
  }
  return &ListResult{
    Items: list,
  }
}

func (h *Handler) mapFileInfoToListItem(f os.FileInfo, item *ListItem, parentPath string, loc *time.Location, ignoreFileTimes bool) {
  item.Name = f.Name()
  item.IsDir = f.IsDir() || isSymlinkToDir(parentPath, f)
  item.Size = f.Size()
  item.ModTime = f.ModTime().Unix()
  ext := strings.ToLower(filepath.Ext(item.Name))
  if h.imageExts[ext] {
    item.Type = "image"
  } else if h.videoExts[ext] {
    item.Type = "video"
  }
  if !ignoreFileTimes {
    t := f.ModTime()
    if loc != nil {
      t = t.In(loc)
    }
    item.ModTimeStr = t.Format(timeFormat)
  }
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
  im, _, err := h.imageFromPath(path, width, height)
  if err != nil {
    return nil, fmt.Errorf("failed to decode image file: %v", err), http.StatusBadRequest
  }

  imageFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
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

func (h *Handler) imageFromPath(path string, width, height int) (image.Image, string, error) {
  ext := strings.ToLower(filepath.Ext(path))
  if (h.videoExts[ext]) {
    return h.imageFromVideo(path, width, height)
  } else {
    return h.imageFromFile(path)
  }
}

func (h *Handler) imageFromFile(path string) (image.Image, string, error) {
  imageFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  f, err := os.Open(imageFilePath)
  if err != nil {
    return nil, "", fmt.Errorf("failed to open file: %v", err)
  }
  defer f.Close()
  return image.Decode(f)
}

func (h *Handler) imageFromVideo(path string, width, height int) (image.Image, string, error) {
  inputFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  cmd := exec.Command("ffmpeg",
      "-i", inputFilePath,
      "-vframes", "1",
      "-f", "singlejpeg",
      "-")
  var buf bytes.Buffer
  cmd.Stdout = &buf
  if err := cmd.Run(); err != nil {
    log.Printf("Error extracting image from video file %v: %v", path, err)
    return nil, "", fmt.Errorf("Error extracting image from video file %v: %v", path, err)
  }
  r := bytes.NewReader(buf.Bytes())
  return image.Decode(r)
}

func (h *Handler) transcodeVideoToCache(path string) error {
  inputFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  transcodedFilePath := h.mp4PathInCache(path)
  transcodedFileDir := filepath.Dir(transcodedFilePath)
  if _, err := os.Stat(transcodedFileDir); os.IsNotExist(err) {
    log.Printf("Creating cache directory %s", transcodedFileDir)
    if err := os.Mkdir(transcodedFileDir, 0700); err != nil {
      return err
    }
  }
  cmd := exec.Command("ffmpeg",
      "-i", inputFilePath,
      "-c:v", "libx264",
      "-preset", "slow",
      "-crf", "18",
      "-c:a", "aac",
      "-strict", "experimental",
      "-b:a", "128k",
      transcodedFilePath)
  log.Printf("Transcoding video file %s to %s", inputFilePath, transcodedFilePath)
  if err := cmd.Run(); err != nil {
    log.Printf("Error transcoding video file %v: %v", path, err)
    return fmt.Errorf("Error transcoding video file %v: %v", path, err)
  }
  log.Printf("Done transcoding video file %s", transcodedFilePath)
  return nil
}

// VideoFilePath returns the path on disk to the specified video file.
// If the extension is not one of our video extensions, returns the empty string.
func (h *Handler) VideoFilePath(path string) (string, error) {
  ext := strings.ToLower(filepath.Ext(path))
  if !h.videoExts[ext] {
    return "", nil;          // Not a video file
  }
  videoFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  if ext == ".mpg" || ext == ".mts" {
    transcodedFilePath := h.mp4PathInCache(path)
    if _, err := os.Stat(transcodedFilePath); os.IsNotExist(err) {
      if err := h.transcodeVideoToCache(path); err != nil {
        return "", err
      }
    }
    return transcodedFilePath, nil
  }
  return videoFilePath, nil
}

func (h *Handler) mp4PathInCache(path string) string {
  inputFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  dir, filename := filepath.Split(inputFilePath)
  ext := filepath.Ext(filename)
  base := strings.TrimSuffix(filename, ext)
  return dir + cacheDir + base + ".mp4"
}

func (h *Handler) Text(path string) ([]byte, error, int) {
  textFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  if filepath.Ext(textFilePath) != textExtension {
    return nil, fmt.Errorf("Text operations can only apply to %s files, not to %s", textExtension, textFilePath), http.StatusBadRequest
  }
  b, err := ioutil.ReadFile(textFilePath)
  if err != nil {
    return nil, fmt.Errorf("failed to read file: %v", err), http.StatusNotFound
  }
  return b, nil, 0
}

func (h *Handler) PutText(path string, command UpdateTextCommand) (error, int) {
  textFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  if filepath.Ext(textFilePath) != textExtension {
    return fmt.Errorf("Text operations can only apply to %s files, not to %s", textExtension, textFilePath), http.StatusBadRequest
  }
  content := command.Content
  if content == "" {
    // Delete the file rather than writing out an empty file.
    err := os.Remove(textFilePath)
    if err != nil && !os.IsNotExist(err) {
      return err, http.StatusInternalServerError
    }
  } else {
    b := []byte(command.Content)
    err := ioutil.WriteFile(textFilePath, b, 0644)
    if err != nil {
      return err, http.StatusInternalServerError
    }
  }
  return nil, http.StatusOK
}
