
package main

import (
	"encoding/json"
  "flag"
  "compress/gzip"
  "strings"
  "io/ioutil"
  "bytes"
  "fmt"
  "net/http"
  "encoding/hex"
  "io"
  "crypto/sha256"
  "os"
)

var (
  configTemplate =
  `{
    "architecture": "amd64",
    "config": {
      "Hostname":"aaa",
      "Domainname":"",
      "User":"",
      "AttachStdin":false,
      "AttachStdout":false,
      "AttachStderr":false,
      "Tty":false,
      "OpenStdin":false,
      "StdinOnce":false,
      "Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
      "Cmd":[],
      "ArgsEscaped":true,
      "Volumes":null,
      "WorkingDir":"/",
      "Entrypoint":["/hello"],
      "OnBuild":null,
      "Labels":{}
    },
    "container": "zzz",
    "container_config":{
    },
    "created":"2019-07-07T10:06:56.611368294Z",
    "docker_version":"1.13.1",
    "history":[
      {"created":"2019-07-07T10:06:56.523525096Z","created_by":"/bin/sh"},
      {"created":"2019-07-07T10:06:56.611368294Z","created_by":"/bin/sh", "empty_layer":true}
    ],
    "os": "linux",
    "rootfs": {
      "type": "layers",
      "diff_ids":[%s]
    }
  }`
)

func hash_file(fname string) string {
  file, err := os.Open(fname)
  if err != nil {
      return ""
  }
  defer file.Close()
  hash := sha256.New()
  _, err = io.Copy(hash, file)
  if err != nil {
      panic(err)
  }

  result := hex.EncodeToString(hash.Sum(nil))
  return result
}

func hash_data(in []byte) (string) {
  sh := sha256.Sum256(in)
  return hex.EncodeToString(sh[:])
}

func readFile(fname string) []byte {
  out, err := ioutil.ReadFile(fname)
  if err != nil {
    panic(err)
  }
  return out
}

func pushBlob(url string, blob []byte) {
  resp, err := http.Post(url+"/blobs/uploads/", "", nil)
  if err != nil {
    panic(err)
  }
  if resp.StatusCode != 202 {
    fmt.Printf("unpexpected status code %d\n", resp.StatusCode)
    panic("err")
  }
  location := resp.Header.Get("location")

  put, err := http.NewRequest(http.MethodPut,
                              location+"&digest=sha256:"+hash_data(blob),
                              bytes.NewReader(blob))
  if err != nil {
    panic(err)
  }
  client := &http.Client{}
  
  resp, err = client.Do(put)
  if err != nil {
    panic(err)
  }
  if resp.StatusCode != 201 {
    fmt.Println(resp)
    panic("unexpected response")
  }
}

func uploadManifest(url string, manifest string) {
  put, err := http.NewRequest(http.MethodPut, url, strings.NewReader(manifest))
  if err != nil {
    panic(err)
  }
  put.Header.Add("content-type", "application/vnd.docker.distribution.manifest.v2+json")
  client := &http.Client{}

  resp, err := client.Do(put)
  if err != nil {
    panic(err)
  }
  if resp.StatusCode != 201 {
    fmt.Println(resp)
    panic("unexpected response")
  }
}

func createConfigBlob(imagehash []string) []byte {
  first := true
  tmp := ""
  for _, hash := range imagehash {
    if first {
      first = false
    } else {
      tmp = tmp + ","
    }
    tmp = tmp + "\"sha256:"+hash+"\""
  }
  c := fmt.Sprintf(configTemplate, tmp)
  return []byte(c)
}

func createmanifest(imagehash []string, confighash string) string {
  manifest := Manifest {
    SchemaVersion: 2,
    MediaType: "application/vnd.docker.distribution.manifest.v2+json",
    Config : Content {
      MediaType: "application/vnd.docker.container.image.v1+json",
      Digest: "sha256:" + confighash,
    },
    Layers : []Content {
    },
  }
  for _, image := range imagehash {
    c := Content{MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip", Digest: "sha256:" + image}
    manifest.Layers = append(manifest.Layers, c)
  }
  m, _ := json.Marshal(manifest)
  return string(m)
}

func gzipBlob(in []byte) []byte {
  var buf bytes.Buffer
  w := gzip.NewWriter(&buf)
  w.Write(in)
  w.Close()
  return buf.Bytes()
}

func main() {
  repoUrlPtr   := flag.String("repo", "http://192.168.1.51:5000", "url of private repo")
  imageNamePtr := flag.String("image", "image1", "name of image")
  imageTagPtr  := flag.String("tag", "v1", "tag")
  tarNamePtr   := flag.String("tar", "image.tar", "tar with image content")
  tar2NamePtr   := flag.String("tar2", "", "second tar with image content")

  flag.Parse()

  repo_url := *repoUrlPtr
  image_name := *imageNamePtr
  image_tag := *imageTagPtr

  tar := readFile(*tarNamePtr)
  targz := gzipBlob(tar)

  imagehashes := []string {hash_data(tar)}
  imagegzhashes := []string {hash_data(targz)}

  pushBlob(repo_url + "/v2/"+image_name, targz)

  if len(*tar2NamePtr) > 0 {
    tar2 := readFile(*tar2NamePtr)
    targz2 := gzipBlob(tar2)
    pushBlob(repo_url + "/v2/"+image_name, targz2)
    imagehashes = append(imagehashes, hash_data(tar2))
    imagegzhashes = append(imagegzhashes, hash_data(targz2))
  }

  cfg := createConfigBlob(imagehashes)
  pushBlob(repo_url + "/v2/"+image_name, cfg)

  man := createmanifest(imagegzhashes, hash_data(cfg))
  uploadManifest(repo_url + "/v2/"+image_name+"/manifests/"+image_tag, man)

  fmt.Println("everything seems ok")
}