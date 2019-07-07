
package main

import (
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
  config =
  `{"architecture":
    "amd64",
    "config": {
      "Hostname":"4cf915974228",
      "Domainname":"",
      "User":"",
      "AttachStdin":false,
      "AttachStdout":false,
      "AttachStderr":false,
      "Tty":false,
      "OpenStdin":false,
      "StdinOnce":false,
      "Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
      "Cmd":["/test.txt aa"],
      "ArgsEscaped":true,
      "Image":"sha256:622277a953205d23e0d8369cc4dc50f3b2d1e83250c8845b6946030a02e59e86",
      "Volumes":null,
      "WorkingDir":"",
      "Entrypoint":null,
      "OnBuild":null,
      "Labels":{}
    },
    "container":"de98817e7be1ca20c5b3c10fa62d331e5ef723788f7adf17adc50ff14192de99",
    "container_config":{
      "Hostname":"4cf915974228",
      "Domainname":"",
      "User":"",
      "AttachStdin":false,
      "AttachStdout":false,
      "AttachStderr":false,
      "Tty":false,
      "OpenStdin":false,
      "StdinOnce":false,
      "Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
      "Cmd":["/bin/sh","-c","#(nop) ", "CMD [\"/test.txt aa\"]"],
      "ArgsEscaped":true,
      "Image":"sha256:622277a953205d23e0d8369cc4dc50f3b2d1e83250c8845b6946030a02e59e86",
      "Volumes":null,
      "WorkingDir":"",
      "Entrypoint":null,
      "OnBuild":null,
      "Labels":{}
    },
    "created":"2019-07-07T10:06:56.611368294Z",
    "docker_version":"1.13.1",
    "history":[
      {"created":"2019-07-07T10:06:56.523525096Z","created_by":"/bin/sh -c #(nop) COPY file:f1726a17794eca97272e61c96b1518114dc8165b0a8ff15c80e242af38b68ebd in / "},
      {"created":"2019-07-07T10:06:56.611368294Z","created_by":"/bin/sh -c #(nop)  CMD [\"/test.txt aa\"]","empty_layer":true}
    ],
    "os":
    "linux",
    "rootfs":{
      "type": "layers",
      "diff_ids":["sha256:%s"]
    }
  }`

  manifest = `
  {
    "schemaVersion": 2,
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "config": {
       "mediaType": "application/vnd.docker.container.image.v1+json",
       "size": 1534,
       "digest": "sha256:%s"
    },
    "layers": [
       {
          "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
          "size": 110,
          "digest": "sha256:%s"
       }
    ]
 }
  `
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
  fmt.Println("location "+location)

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
  fmt.Println(resp)
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
  fmt.Println(resp)
}

func createConfigBlob(imagehash string) []byte {
  c := fmt.Sprintf(config, imagehash)
  return []byte(c)
}

func createmanifest(imagehash string, confighash string) string {
  c := fmt.Sprintf(manifest, confighash, imagehash)
  return c
}

func main() {
  tar := readFile("image.tar")
  targz := readFile("image.tar.gz")
  pushBlob("http://192.168.1.51:5000/v2/abc", targz)
  cfg := createConfigBlob(hash_data(tar))
  pushBlob("http://192.168.1.51:5000/v2/abc", cfg)

  man := createmanifest(hash_data(targz), hash_data(cfg))
  uploadManifest("http://192.168.1.51:5000/v2/abc/manifests/latest", man)
}