package main
import "gopkg.in/redis.v2"
import "log"
import "encoding/json"
import "math"
import "net/url"
import "os"
import "net/http"
import "path/filepath"
import "io"
import "fmt"


type QueuedDownload struct {
    Url   string    `json:"url"`
    FileName string `json:"filename"`
}

var client * redis.Client
var targetDir string
var logger * log.Logger = log.New(os.Stderr, "", log.LstdFlags)

func getRedisAddr() string {
  str := os.Getenv("REDIS_URL")
  if str == "" {
    return "localhost:6379"
  }

  url, err := url.Parse(str)
  if err != nil {
    logger.Fatal(err)
    return ""
  } else {
    return url.Host
  }
}

func (qd * QueuedDownload) Download() {
  response, err := http.Get(qd.Url)
  if err != nil {
    logger.Println("Error while downloading", qd.Url, "-", err)
    return
  }
  defer response.Body.Close()

  if response.StatusCode != 200 {
    logger.Println("Error while fetching", qd.Url, "-", response.Status)
    return
  }

  logger.Println("Downloading ContentLength:", response.ContentLength, "bytes")

  outputPath := filepath.Join(targetDir, qd.FileName)
  tmpOutputPath := fmt.Sprintf("%s.downloading", outputPath)
  output, err := os.Create(tmpOutputPath)
  if err != nil {
    logger.Println("Error while creating", outputPath, "-", err)
    return
  }
  defer output.Close()

  n, err := io.Copy(output, response.Body)
  if err != nil {
    logger.Println("Error while downloading", qd.Url, "-", err)
    return
  }

  logger.Println(n, "bytes downloaded.")

  if err := os.Rename(tmpOutputPath, outputPath); err != nil {
    logger.Println("Error while moving file into place -", err)
  }
}

func main() {
  targetDir = os.Getenv("MEDIA_FETCHER_TARGET")
  if targetDir == "" {
    logger.Fatal("MEDIA_FETCHER_TARGET must be set")
  }
  os.MkdirAll(targetDir, 0777)

  if logPath := os.Getenv("MEDIA_FETCHER_LOG"); logPath != "" {
    logFile, err := os.Create(logPath)
    if err != nil {
      logger.Fatal("Error while creating", logFile, "-", err)
    }
    defer logFile.Close()
    logger = log.New(logFile, "", log.LstdFlags)
  }

  client = redis.NewTCPClient(&redis.Options{
      Addr:     getRedisAddr(),
      Password: "", // no password set
      DB:       0,  // use default DB
  })

  _, err := client.Ping().Result()
  if err != nil {
    logger.Fatal(err)
  }

  for {
    result := client.BLPop(math.MaxInt32, "download-queue")
    val, err := result.Result()
    if err != nil {
      logger.Println("Redis Fetch Error:", err)
    } else {
      download := &QueuedDownload{}
      json.Unmarshal([]byte(val[1]), &download)
      logger.Println(download)
      client.Set("download-queue:processing", download.Url)
      download.Download()
      client.Del("download-queue:processing")
    }
  }
}
