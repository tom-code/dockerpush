package main


//it would be better to use stuff from https://github.com/docker/distribution/blob/master/manifest/
//but we can't since it vendors digest.Digest and we can't assign to it :((((( dirty dirty

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

type ConfigHistoryItem struct {
  Created string
  CreatedBy string
  EmptyLayer bool
}

type ConfigRootFs struct {
  Type string         `json:"type"`
  DiffIds []string    `json:"diff_ids"`
}
type ConfigConfig struct {
  Hostname string
  Domainname string
  User string
  AttachStdIn bool
  AttachStdOut bool
  AttachStdErr bool
  Tty bool
  OpenStdin bool
  StdinInce bool
  Env []string
  Cmd string
  ArgsEscaped bool
  WorkingDir string
  Entrypoint []string
}

type Config struct {
  Architecture string
  Container string
  Os string
  Created string
  DockerVersion string `json:"docker_version"`
  Config ConfigConfig
  History []ConfigHistoryItem
  Rootfs ConfigRootFs `json:"rootfs"`
}
