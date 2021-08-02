package podman

import (
	"context"

	"github.com/containers/podman/v3/libpod"
)

type Config struct {
	// RuntimeConfig is an optional path to a Podman runtime configuration.
	RuntimeConfig *string `json:"runtime_config,omitempty"`
	// Image is the container image to pull. This should be a repository URL,
	// ie., containers.dev.maio.me/library/alpine:latest
	Image      string `json:"image"`
	AlwaysPull bool   `json:"always_pull"`
	Vault      struct {
		RootSecret string `json:"root_secret"`
	} `json:"vault"`
}

type Provisioner struct {
	cfg         *Config
	containerID string
	ctx         context.Context
	ctxCancel   context.CancelFunc
	runtime     *libpod.Runtime
}
