package harness

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	vaultApi "github.com/hashicorp/vault/api"

	"github.com/pirogoeth/vault-init/initializer"
	"github.com/pirogoeth/vault-init/pkg/harness/provisioner"
	dockerProv "github.com/pirogoeth/vault-init/pkg/harness/provisioner/docker"
	podmanProv "github.com/pirogoeth/vault-init/pkg/harness/provisioner/podman"
	unmanagedProv "github.com/pirogoeth/vault-init/pkg/harness/provisioner/unmanaged"
	"github.com/pirogoeth/vault-init/pkg/harness/util/stringlist"
)

func (s *Scenario) SetupFixtures(vaultCli *vaultApi.Client) error {
	for _, mount := range s.Fixtures.Mounts {
		if err := s.wipeMountIfExists(vaultCli, mount); err != nil {
			return fmt.Errorf("error while removing preexisting mount: %w", err)
		}

		if err := s.createMount(vaultCli, mount); err != nil {
			return fmt.Errorf("error while creating mount fixtures: %w", err)
		}
	}

	for _, secret := range s.Fixtures.Secrets {
		if err := s.createSecret(vaultCli, secret); err != nil {
			return fmt.Errorf("error while creating secret fixtures: %w", err)
		}
	}

	return nil
}

func (s *Scenario) createMount(vaultCli *vaultApi.Client, mount *mountFixture) error {
	if err := vaultCli.Sys().Mount(mount.Path, mount.Config); err != nil {
		return fmt.Errorf("during mount fixture %s setup: error while mounting engine: %w", mount.Path, err)
	}

	log.Debugf("Engine mounted at %s", mount.Path)

	return nil
}

func (s *Scenario) wipeMountIfExists(vaultCli *vaultApi.Client, mount *mountFixture) error {
	cfg, err := vaultCli.Sys().MountConfig(mount.Path)
	if cfg != nil {
		return vaultCli.Sys().Unmount(mount.Path)
	}

	if err, ok := err.(*api.ResponseError); ok {
		if err.StatusCode == 400 {
			return nil
		}
	}

	return err
}

func (s *Scenario) createSecret(vaultCli *vaultApi.Client, secret *secretFixture) error {
	_, err := vaultCli.Logical().Write(secret.Path, secret.Data)
	if err != nil {
		return fmt.Errorf("during secret fixture %s setup: error while creating secret: %w", secret.Path, err)
	}

	data, err := vaultCli.Logical().Read(secret.Path)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	log.Debugf("Secret created: %#v", data)

	return nil
}

func (s *Scenario) TeardownFixtures(vaultCli *vaultApi.Client) error {
	for _, secret := range s.Fixtures.Secrets {
		if err := s.deleteSecret(vaultCli, secret); err != nil {
			return fmt.Errorf("error while destroying secret fixtures: %w", err)
		}
	}

	for _, mount := range s.Fixtures.Mounts {
		if err := s.deleteMount(vaultCli, mount); err != nil {
			return fmt.Errorf("error while destroying mount fixtures: %w", err)
		}
	}

	return nil
}

func (s *Scenario) deleteMount(vaultCli *vaultApi.Client, mount *mountFixture) error {
	if err := vaultCli.Sys().Unmount(mount.Path); err != nil {
		return fmt.Errorf("during mount fixture %s teardown: error while unmounting engine: %w", mount.Path, err)
	}

	log.Debugf("Engine %s unmounted", mount.Path)

	return nil
}

func (s *Scenario) deleteSecret(vaultCli *vaultApi.Client, secret *secretFixture) error {
	if _, err := vaultCli.Logical().Delete(secret.Path); err != nil {
		return fmt.Errorf("during secret fixture %s teardown: error while deleting secret: %w", secret.Path, err)
	}

	log.Debugf("Secret %s deleted", secret.Path)

	return nil
}

func (s *Scenario) CreateProvisioner() (provisioner.Provisioner, error) {
	cfg := s.VaultProvider.ProvisionerCfg
	var provisioner provisioner.Provisioner
	switch strings.ToLower(cfg.Driver) {
	case "docker":
		provisioner = dockerProv.New()
	case "podman":
		provisioner = podmanProv.New()
	case "unmanaged":
		provisioner = unmanagedProv.New()
	default:
		return nil, fmt.Errorf("unknown provisioner driver: %s", cfg.Driver)
	}

	if err := provisioner.Configure(&cfg.Config); err != nil {
		return nil, fmt.Errorf("while creating provisioner %s, failed to configure provisioner: %w", cfg.Driver, err)
	}

	return provisioner, nil
}

func (s *Scenario) ConfigureVaultInitFromVaultClient(vaultCli *vaultApi.Client) error {
	if s.VaultInitCfg.VaultAddress == "" {
		s.VaultInitCfg.VaultAddress = vaultCli.Address()
		log.Debugf("Setting vault-init connection address: %s", s.VaultInitCfg.VaultAddress)
	} else {
		log.Infof("Scenario sets a Vault address - not overriding")
	}

	if s.VaultInitCfg.VaultToken == "" {
		s.VaultInitCfg.VaultToken = vaultCli.Token()
		log.Debugf("Setting vault-init token: %s", s.VaultInitCfg.VaultToken)
	} else {
		log.Infof("Scenario sets a Vault token - not overriding")
	}

	return nil
}

func (s *Scenario) RunTests() error {
	results := make(chan testSuiteResult)
	for _, test := range s.Tests {
		test.Environment[EnvUnderTest] = "yes"
		go s.runTest(results, test)
	}
	completed := 0
	for {
		select {
		case result := <-results:
			fmt.Printf("test result %#v\n", result)
			printReader(result.StdoutReader)
			printReader(result.StderrReader)
			completed += 1
		case <-time.After(1 * time.Second):
		}

		if len(s.Tests) == completed {
			break
		}
	}
	return fmt.Errorf("not implemented")
}

func (s *Scenario) runTest(resultChan chan<- testSuiteResult, suite *testSuite) {
	for key, value := range suite.Environment {
		os.Setenv(key, value)
	}

	command := []string{"go", "test"}
	command = append(command, suite.Args...)
	if !stringlist.Contains(command, ArgGoTestJson) {
		command = append(command, ArgGoTestJson)
	}
	command = append(command, suite.Suite)

	initCfg, err := s.VaultInitCfg.Clone()
	if err != nil {
		panic(fmt.Errorf("while cloning initializer.Config, got error: %w", err))
	}
	initCfg.Command = command

	rTestStderr, wTestStderr := io.Pipe()
	initCfg.ForwarderStderrWriters = append(initCfg.ForwarderStderrWriters, wTestStderr)

	rTestStdout, wTestStdout := io.Pipe()
	initCfg.ForwarderStdoutWriters = append(initCfg.ForwarderStdoutWriters, wTestStdout)

	result := newTestSuiteResult(rTestStderr, rTestStdout)

	ctx, _ := context.WithCancel(context.Background())
	err = initializer.Run(ctx, initCfg)
	if err != nil {
		result.Error = err
	}
	// childCancel()

	resultChan <- *result
}

func printReader(r io.Reader) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Printf("error reading from reader: %#v", err)
	}

	data := string(b)
	fmt.Printf("printReader(%#v): %s\n", r, data)
}
