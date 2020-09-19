package content

import (
  "bytes"
  "fmt"
  "image"
  "image/color"
  _ "image/jpeg"
  "io"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "os/exec"
  "path"
  "path/filepath"
  "strings"
  "time"

  "github.com/disintegration/imaging"
  "github.com/rwcarlsen/goexif/exif"
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
  Path string           // Full API path if the item is not in the parent dir
  IndexPath string      // Full API path to the index file, if not the default index
  IndexEntry string     // The path to the file relative to the index, if not in the default index
  IsDir bool
  Size int64
  Type string
  ExifDateTime time.Time
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

// dirInfo stores info loaded from one directory.
type dirInfo struct {
  loc *time.Location    // info from TZ file
  flags dirFlags  // flags from summary.txt file
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

  flags := loadDirFlags(dirPath)

  files, err, status := h.readDirFiltered(dirPath, flags.sortByFileTimes)
  if err != nil {
    return nil, err, status
  }

  imageIndex := h.imageIndex(dirPath)
  unfilteredFileCount := len(files)
  if imageIndex != nil {
    files = imageIndex.filter(files)
  }

  loc := readTzFile(dirPath)

  result := h.mapFileInfosToListResult(files, dirPath, loc, flags.ignoreFileTimes)
  result.UnfilteredFileCount = unfilteredFileCount
  if imageIndex != nil {
    result.IndexName = imageIndex.indexName
  }
  return result, nil, 0
}

// ListFromIndex creates a list of files as given in the specified index file.
func (h *Handler) ListFromIndex(indexApiPath string) (*ListResult, error, int) {
  contentRoot := strings.TrimSuffix(h.config.ContentRoot, "/")
  indexApiPath = strings.TrimSuffix(indexApiPath, "/")
  indexApiDir := path.Dir(indexApiPath)
  indexPath := fmt.Sprintf("%s/%s", contentRoot, indexApiPath)
  indexName := path.Base(indexPath)
  dirPath := path.Dir(indexPath)
  imageIndex := h.loadIndexFile(dirPath, indexName)
  if imageIndex == nil {
    return nil, fmt.Errorf("Failed to load index file %s", indexApiPath), http.StatusBadRequest
  }

  dirSet := imageIndex.dirSet() // The directories as they appear in the index file entries.
  dirPaths := validDirsFromSet(contentRoot, dirPath, dirSet)  // Resolved and validated.

  dirInfos := make(map[string]dirInfo, len(dirPaths))
  for dirPath, _ := range dirPaths {
    di := dirInfo{}
    di.loc = readTzFile(dirPath)
    di.flags = loadDirFlags(dirPath)
    dirInfos[dirPath] = di
  }

  n := len(imageIndex.filenames)
  list := make([]ListItem, n, n)
  for i, fn := range imageIndex.filenames {
    d := path.Dir(fn)
    dir := path.Join(dirPath, d)
    base := path.Base(fn)
    dirInfo, ok := dirInfos[dir]
    if ok {
      realfn := path.Join(dir, base)
      f, err := os.Stat(realfn)
      if err != nil {
        return nil, fmt.Errorf("Error reading file info for %s", fn), http.StatusInternalServerError
      }
      h.mapFileInfoToListItem(f, &list[i], dir, dirInfo.loc, dirInfo.flags.ignoreFileTimes)
      list[i].Path = path.Join("/", indexApiDir, fn)
      list[i].IndexPath = indexApiPath
      list[i].IndexEntry = fn
    }
  }
  return &ListResult{
    Items: list,
  }, nil, 0
}

// validDirsFromSet takes a set of relative directories and returns
// the equivalent set of directories resolved against dirPath, removing
// any that are not withing contentRoot.
func validDirsFromSet(contentRoot, dirPath string, dirSet map[string]struct{}) map[string]struct{} {
  contentRoot = path.Clean(contentRoot)
  dirPaths := make(map[string]struct{}, 0)
  for d, _ := range dirSet {
    dir := path.Join(dirPath, d)
    if strings.HasPrefix(dir, contentRoot) {
      dirPaths[dir] = struct{}{}
    }
  }
  return dirPaths
}

func readTzFile(dirPath string) *time.Location {
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
  return loc
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
  } else if ext == ".mpr" {
    item.Type = "index"
  }
  if !ignoreFileTimes {
    t := f.ModTime()
    if loc != nil {
      t = t.In(loc)
    }
    item.ModTimeStr = t.Format(timeFormat)
  }
  h.loadTextFile(item, parentPath)
  h.loadExifDateTime(item, parentPath)
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

// loadExifDateTime opens the image file, reads the DateTime field
// from the exif data, and stores it in the item.
func (h *Handler) loadExifDateTime(item *ListItem, parentPath string) {
    imagepath := fmt.Sprintf("%s/%s", parentPath, item.Name)
    datetime, err := datetimeFromFile(imagepath)
    if err != nil {
        log.Printf("Error getting EXIF DateTime for parent %s: %v", parentPath, err)
    } else {
        item.ExifDateTime = datetime
    }
}

func (h *Handler) Image(path string, width, height, rot int) (image.Image, error, int) {
  im, exifOrientation, _, err := h.imageFromPath(path, width, height)
  if err != nil {
    return nil, fmt.Errorf("failed to decode image file: %v", err), http.StatusBadRequest
  }

  imageFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  exifRotation := exifOrientationToRotation(exifOrientation)
  rot = rot + h.rotationFromIndexAndExif(imageFilePath, exifRotation)
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
    im = imaging.Resize(im, width, height, imaging.Box)
  }

  if rot != 0 {
    im = imaging.Rotate(im, float64(rot), color.Black)
  }

  return im, nil, 0
}

func (h *Handler) imageFromPath(path string, width, height int) (image.Image, int, string, error) {
  ext := strings.ToLower(filepath.Ext(path))
  if (h.videoExts[ext]) {
    return h.imageFromVideo(path, width, height)
  } else {
    return h.imageFromFile(path)
  }
}

// Read and return an image from a file.
// The int value is the EXIF orientation from the file, or -1 if we don't have one.
// The string return value is the name of the image format.
func (h *Handler) imageFromFile(path string) (image.Image, int, string, error) {
  imageFilePath := fmt.Sprintf("%s/%s", h.config.ContentRoot, path)
  // log.Printf("Loading image file %s", imageFilePath)
  f, err := os.Open(imageFilePath)
  if err != nil {
    return nil, -1, "", fmt.Errorf("failed to open file: %v", err)
  }
  defer f.Close()

  // Read the orientation from the exif header in the file.
  orientation := orientationFromOpenFile(imageFilePath, f)
  // log.Printf("Exif Orientation for %s is %v", imageFilePath, orientation)

  _, err = f.Seek(0, io.SeekStart)
  if err != nil {
    log.Printf("Error seeking back to 0 on file %s: %v", imageFilePath, err)
    return nil, -1, "", err
  }

  img, imgFmt, err := image.Decode(f)
  return img, orientation, imgFmt, err
}

func datetimeFromFile(imageFilePath string) (time.Time, error) {
  f, err := os.Open(imageFilePath)
  if err != nil {
    return time.Time{}, fmt.Errorf("failed to open file: %v", err)
  }
  defer f.Close()
  return datetimeFromOpenFile(imageFilePath, f), nil
}

// datetimeFromOpenFile reads the Datetime from the exif header in the file
// and returns it as a time.Time. If for any reason it is unable to read it, it
// returns the zero time. This function moves the file position in f.
func datetimeFromOpenFile(path string, f *os.File) time.Time {
  dt := time.Time{}     // Set to default value.
  x, err := exif.Decode(f)
  if err != nil {
    log.Printf("Can't read Exif for datetime from %s: %v", path, err)
    return dt
  }
  dt, err = x.DateTime()
  if err != nil {
    log.Printf("Can't read DateTime from %s: %v", path, err)
    dt = time.Time{}
    return dt
  }
  log.Printf("DateTime for %s is %v", path, dt)
  return dt
}

// orientationFromOpenFile reads the Orientation from the exif header in the file
// and returns it as an int. If for any reason it is unable to read it, it
// returns -1. This function moves the file position in f.
func orientationFromOpenFile(path string, f *os.File) int {
  orientation := -1      // Preset to not-present value.
  x, err := exif.Decode(f)
  if err != nil {
    log.Printf("Can't read Exif for orientation from %s: %v", path, err)
    return orientation
  }
  oTag, err := x.Get(exif.Orientation)
  if err != nil {
    log.Printf("Can't read Orientation from %s: %v", path, err)
    return orientation
  }
  // log.Printf("Orientation for %s is %v", path, oTag)
  orientation, err = oTag.Int(0)
  if err != nil {
      log.Printf("Can't extract Orientation for %s from tag %v: %v", path, oTag, err)
      orientation = -1
  }
  return orientation
}

func (h *Handler) imageFromVideo(path string, width, height int) (image.Image, int, string, error) {
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
    return nil, -1, "", fmt.Errorf("Error extracting image from video file %v: %v", path, err)
  }
  r := bytes.NewReader(buf.Bytes())
  img, imgFmt, err := image.Decode(r)
  return img, -1, imgFmt, err
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

// Given an EXIF Orientation, returns the number of degress of counterclockwise
// rotation required to  display the image with the appropriate edge at the top.
// If the value is not one of the 8 defined orientations, returns 0.
func exifOrientationToRotation(exifOrientation int) int {
  switch exifOrientation {
  case 1, 2: return 0
  case 3, 4: return 180
  case 5, 6: return -90
  case 7, 8: return 90
  }
  return 0
}
