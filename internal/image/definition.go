package image

import (
	"fmt"
)

const (
	TypeRAW = "raw"

	ArchTypeX86 Arch = "x86_64"
	ArchTypeARM Arch = "aarch64"
)

type Arch string

func (a Arch) Short() string {
	switch a {
	case ArchTypeX86:
		return "amd64"
	case ArchTypeARM:
		return "arm64"
	default:
		message := fmt.Sprintf("unknown arch: %s", a)
		panic(message)
	}
}

type Definition struct {
	Image           Image
	Installation    Installation
	OperatingSystem OperatingSystem
	Release         Release
}

type Image struct {
	ImageType       string
	Arch            Arch
	OutputImageName string
}

type OperatingSystem struct {
	Users []User `yaml:"users"`
}

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Installation struct {
	Target string `yaml:"target"`
}

type Release struct {
	KubernetesURL string `yaml:"kubernetesUrl"`
}
