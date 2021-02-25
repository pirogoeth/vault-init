package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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
		resp.Close()
	} else if err != nil {
		return fmt.Errorf("while provisioning Vault instance, could not inspect image: %w", err)
	}

	genSecret := "secret"

	containerCfg := image.Config
	containerCfg.Image = image.ID
	containerCfg.Env = append(
		containerCfg.Env,
		"VAULT_DEV_LISTEN_ADDRESS=0.0.0.0",
		fmt.Sprintf("VAULT_DEV_ROOT_TOKEN_ID=%s", genSecret),
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

func (p *Provisioner) GenerateVaultAPIConfig() (*vaultApi.Config, error) {
	if p.cfg == nil {
		return nil, fmt.Errorf("docker provisioner has not been properly initialized")
	}

	containerInfo, err := p.dockerClient.ContainerInspect(p.ctx, p.containerID)
	if err != nil {
		return nil, fmt.Errorf("while generating Vault API config, could not inspect container: %w", err)
	}

	httpAPIPort, err := p.checkPorts(containerInfo.NetworkSettings.Ports["8200/tcp"])
	if err != nil {
		return nil, fmt.Errorf("while generating Vault API config, could not test port bindings: %w", err)
	}
	hostIP := httpAPIPort.HostIP
	hostPort := httpAPIPort.HostPort

	log.Infof("Vault appears to be listening on %s:%s", hostIP, hostPort)

	return nil, nil
}

func (p *Provisioner) checkPorts(ports []nat.PortBinding) (nat.PortBinding, error) {
	tries := 3

	for tries > 0 {
		for _, port := range ports {
			hostIP := port.HostIP
			hostPort := port.HostPort

			testCfg := &vaultApi.Config{Address: fmt.Sprintf("http://%s:%s", hostIP, hostPort)}
			testCli, err := vaultApi.NewClient(testCfg)
			if err != nil {
				log.Debugf("while testing port binding %s:%s, could not create Vault client: %w", hostIP, hostPort, err)
				continue
			}

			health, err := testCli.Sys().Health()
			if err != nil {
				log.Debugf("while testing port binding %s:%s, could not get health status: %w", hostIP, hostPort, err)
				continue
			}

			log.Debugf("Got health response from %s:%s: %#v", hostIP, hostPort, health)

			return port, nil
		}

		tries--
		time.Sleep(1 * time.Second)
	}

	return nat.PortBinding{}, fmt.Errorf("no port mappings for 8200/tcp returned a health response")
}
