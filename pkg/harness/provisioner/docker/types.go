package docker

import (
	"context"

	"github.com/docker/docker/client"
)

type Config struct {
	// Image is the container image to pull. This should be a repository URL,
	// ie., containers.dev.maio.me/library/alpine:latest
	Image      string `json:"image"`
	AlwaysPull bool   `json:"always_pull"`
	Vault      struct {
		RootSecret string `json:"root_secret"`
	} `json:"vault"`
}

type Provisioner struct {
	cfg          *Config
	containerID  string
	ctx          context.Context
	ctxCancel    context.CancelFunc
	dockerClient *client.Client
}
