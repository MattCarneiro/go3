package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "regexp"
  "strings"

  "github.com/gin-gonic/gin"
  "google.golang.org/api/drive/v3"
  "google.golang.org/api/googleapi/transport"
  "google.golang.org/api/option"
)

var (
  mimeTypes = map[string]string{
    "pdf":   "application/pdf",
    "image": "image/",
    "video": "video/",
  }
  apiKey string
)

func main() {
  apiKey = os.Getenv("GOOGLE_DRIVE_API_KEY")
  if apiKey == "" {
    log.Fatal("GOOGLE_DRIVE_API_KEY environment variable is required")
  }

  r := gin.Default()
  r.POST("/check-downloadable", checkDownloadableHandler)
  
  port := os.Getenv("PORT")
  if port == "" {
    port = "3000"
  }
  
  r.Run(":" + port)
}

func checkDownloadableHandler(c *gin.Context) {
  var req struct {
    Link string `json:"link"`
    Type string `json:"type"`
  }
  
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
    return
  }

  if _, ok := mimeTypes[req.Type]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
    return
  }

  id := extractIdFromLink(req.Link)
  if id == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link format"})
    return
  }

  var downloadable bool
  var err error
  
  if strings.Contains(req.Link, "/folders/") {
    downloadable, err = checkFolder(id, req.Type)
  } else {
    downloadable, err = isDownloadable(id, req.Type)
  }

  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  result := "no"
  if downloadable {
    result = "yes"
  }
  
  c.JSON(http.StatusOK, gin.H{"result": result})
}

func isDownloadable(fileId, fileType string) (bool, error) {
  srv, err := drive.NewService(c, option.WithAPIKey(apiKey))
  if err != nil {
    return false, err
  }

  file, err := srv.Files.Get(fileId).Fields("mimeType").Do()
  if err != nil {
    return false, err
  }

  switch fileType {
  case "pdf":
    return file.MimeType == mimeTypes["pdf"], nil
  case "image":
    return strings.HasPrefix(file.MimeType, mimeTypes["image"]), nil
  case "video":
    return strings.HasPrefix(file.MimeType, mimeTypes["video"]), nil
  default:
    return false, nil
  }
}

func checkFolder(folderId, fileType string) (bool, error) {
  srv, err := drive.NewService(c, option.WithAPIKey(apiKey))
  if err != nil {
    return false, err
  }

  files, err := srv.Files.List().
    Q(fmt.Sprintf("'%s' in parents", folderId)).
    Fields("files(id, mimeType)").
    Do()
  if err != nil {
    return false, err
  }

  if len(files.Files) == 0 {
    return false, nil
  }

  for _, file := range files.Files {
    switch fileType {
    case "pdf":
      if file.MimeType == mimeTypes["pdf"] {
        return true, nil
      }
    case "image":
      if strings.HasPrefix(file.MimeType, mimeTypes["image"]) {
        return true, nil
      }
    case "video":
      if strings.HasPrefix(file.MimeType, mimeTypes["video"]) {
        return true, nil
      }
    }
  }

  return false, nil
}

func extractIdFromLink(link string) string {
  fileIdRegex := regexp.MustCompile(`/d/([a-zA-Z0-9-_]+)`)
  folderIdRegex := regexp.MustCompile(`/folders/([a-zA-Z0-9-_]+)`)

  if fileIdMatch := fileIdRegex.FindStringSubmatch(link); len(fileIdMatch) > 1 {
    return fileIdMatch[1]
  }
  if folderIdMatch := folderIdRegex.FindStringSubmatch(link); len(folderIdMatch) > 1 {
    return folderIdMatch[1]
  }
  return ""
}
