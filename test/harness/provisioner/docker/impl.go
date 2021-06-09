package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	vaultApi "github.com/hashicorp/vault/api"

	"glow.dev.maio.me/seanj/vault-init/test/harness/provisioner"
)

var _ provisioner.Provisioner = (*Provisioner)(nil)

func New() provisioner.Provisioner {
	ctx, cancel := context.WithCancel(context.TODO())
	return &Provisioner{
		ctx:       ctx,
		ctxCancel: cancel,
	}
}

func (p *Provisioner) Configure(cfgJson *json.RawMessage) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return fmt.Errorf("while configuring docker provisioner, could not create docker client: %w", err)
	}
	p.dockerClient = dockerClient

	cfg := new(Config)
	if err := json.Unmarshal(*cfgJson, cfg); err != nil {
		return fmt.Errorf("while configuring docker provisioner, could not parse provisioner config: %w", err)
	}
	p.cfg = cfg

	return nil
}

func (p *Provisioner) Provision() error {
	if p.cfg == nil {
		return fmt.Errorf("docker provisioner has not been properly initialized")
	}

	image, _, err := p.dockerClient.ImageInspectWithRaw(p.ctx, p.cfg.Image)
	if client.IsErrNotFound(err) || p.cfg.AlwaysPull {
		log.Infof("Pulling Docker image: %s", p.cfg.Image)
		resp, err := p.dockerClient.ImagePull(p.ctx, p.cfg.Image, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("while provisioning Vault instance, could not pull container image: %w", err)
		}
		childCtx, childCancel := context.WithCancel(p.ctx)
		waitUntilCompletion(childCtx, childCancel, resp)
		resp.Close()
	} else if err != nil {
		return fmt.Errorf("while provisioning Vault instance, could not inspect image: %w", err)
	}

	containerCfg := image.Config
	containerCfg.Image = image.ID
	containerCfg.Env = append(
		containerCfg.Env,
		"VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200",
		fmt.Sprintf("VAULT_DEV_ROOT_TOKEN_ID=%s", p.cfg.Vault.RootSecret),
	)
	hostCfg := &container.HostConfig{
		AutoRemove:      false,
		NetworkMode:     container.NetworkMode("bridge"),
		PublishAllPorts: true,
	}
	netCfg := &network.NetworkingConfig{}
	name := ""

	created, err := p.dockerClient.ContainerCreate(p.ctx, containerCfg, hostCfg, netCfg, name)
	if err != nil {
		return fmt.Errorf("while provisioning Vault instance, could not create container: %w", err)
	}
	log.Infof("Vault instance created: docker:%s", created.ID)

	startOpts := types.ContainerStartOptions{}
	if err := p.dockerClient.ContainerStart(p.ctx, created.ID, startOpts); err != nil {
		return fmt.Errorf("while provisioning Vault instance, could not start container: %w", err)
	}
	log.Infof("Vault instance started: docker:%s", created.ID)

	p.containerID = created.ID

	return nil
}

func (p *Provisioner) Deprovision() error {
	if p.cfg == nil {
		return fmt.Errorf("docker provisioner has not been properly initialized")
	}

	removeOpts := types.ContainerRemoveOptions{Force: true}
	if err := p.dockerClient.ContainerRemove(p.ctx, p.containerID, removeOpts); err != nil {
		return fmt.Errorf("while deprovisioning Vault instance, could not remove container: %w", err)
	}

	return nil
}

func (p *Provisioner) SpawnVaultAPIClient() (*vaultApi.Client, error) {
	if p.cfg == nil {
		return nil, fmt.Errorf("docker provisioner has not been properly initialized")
	}

	containerInfo, err := p.dockerClient.ContainerInspect(p.ctx, p.containerID)
	if err != nil {
		return nil, fmt.Errorf("while generating Vault API client, could not inspect container: %w", err)
	}

	containerDefaultAddress := containerInfo.NetworkSettings.DefaultNetworkSettings.IPAddress
	err = p.checkAddress(containerDefaultAddress, 8200)
	if err != nil {
		return nil, fmt.Errorf("while spawning Vault API client, could not test port bindings: %w", err)
	}

	cfg := &vaultApi.Config{Address: fmt.Sprintf("http://%s:%d", containerDefaultAddress, 8200)}
	client, err := vaultApi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("while spawning Vault API client: %w", err)
	}

	client.SetToken(p.cfg.Vault.RootSecret)

	return client, nil
}

func (p *Provisioner) checkAddress(addr string, port uint16) error {
	tries := 3

	for tries > 0 {
		testCfg := &vaultApi.Config{Address: fmt.Sprintf("http://%s:%d", addr, port)}
		testCli, err := vaultApi.NewClient(testCfg)
		if err != nil {
			log.Debugf("while testing port binding %s:%d, could not create Vault client: %+v", addr, port, err)
			tries--
			time.Sleep(1 * time.Second)
			continue
		}

		health, err := testCli.Sys().Health()
		if err != nil {
			log.Debugf("while testing port binding %s:%d, could not get health status: %+v", addr, port, err)
			tries--
			time.Sleep(1 * time.Second)
			continue
		}

		log.Debugf("Got health response from %s:%d: %#v", addr, port, health)
		return nil
	}

	return fmt.Errorf("address %s:%d did not return a health response", addr, port)
}

func waitUntilCompletion(ctx context.Context, cancel context.CancelFunc, reader io.Reader) error {
	lines := make(chan string)
	go readToEnd(reader, lines)
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				cancel()
			} else {
				log.Debugf("Read: %s", line)
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			continue
		}
	}
}

func readToEnd(reader io.Reader, dest chan<- string) {
	buf := bufio.NewScanner(reader)
	for {
		if ok := buf.Scan(); !ok {
			close(dest)
			if err := buf.Err(); err != nil {
				panic(err)
			}

			return
		}

		dest <- buf.Text()
	}
}
