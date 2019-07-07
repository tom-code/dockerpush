
package main

import (
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
  config =
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
      "Cmd":["/hello"],
      "ArgsEscaped":true,
      "xImage":"sha256:%s",
      "Volumes":null,
      "WorkingDir":"",
      "Entrypoint":null,
      "OnBuild":null,
      "Labels":{}
    },
    "container":"zzz",
    "container_config":{
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
      "Cmd":["/bin/sh","-c","#(nop) ", "CMD [\"/hello\"]"],
      "ArgsEscaped":true,
      "xImage":"sha256:%s",
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
  c := fmt.Sprintf(config, imagehash, imagehash, imagehash)
  return []byte(c)
}

func createmanifest(imagehash string, confighash string) string {
  c := fmt.Sprintf(manifest, confighash, imagehash)
  return c
}

func gzipBlob(in []byte) []byte {
  var buf bytes.Buffer
  w := gzip.NewWriter(&buf)
  w.Write(in)
  w.Close()
  return buf.Bytes()
}

func main() {

  repo_url := "http://192.168.1.51:5000"
  image_name := "i1"
  image_tag := "v1"

  tar := readFile("image.tar")
  targz := gzipBlob(tar)
  pushBlob(repo_url + "/v2/"+image_name, targz)
  cfg := createConfigBlob(hash_data(tar))
  pushBlob(repo_url + "/v2/"+image_name, cfg)

  man := createmanifest(hash_data(targz), hash_data(cfg))
  uploadManifest(repo_url + "/v2/"+image_name+"/manifests/"+image_tag, man)
}