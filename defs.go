package main


//it would be better to use stuff from https://github.com/docker/distribution/blob/master/manifest/
//but we can't since it vendors digest.Digest and we can't assign to it :((((( dirty boys

type Content struct {
  MediaType string `json:"mediaType"`
  Digest string `json:"digest"`
}

type Manifest struct {
  SchemaVersion int `json:"schemaVersion"`
  MediaType string  `json:"mediaType,omitempty"`
  Config  Content   `json:"config"`
  Layers []Content  `json:"layers"`
}