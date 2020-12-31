package testintegration

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/lobsterdore/release-dash/testsupport"
)

type integrationSupport struct {
	AppContainer    testcontainers.Container
	GhMockContainer testcontainers.Container
	Network         testcontainers.Network
	ProjectName     string
}

func NewIntegrationSupport() (*integrationSupport, error) {
	var err error
	ctx := context.Background()
	networkName := "integration-test"
	intSupport := integrationSupport{
		ProjectName: "release-dash",
	}

	intSupport.Network, err = intSupport.createTestNetwork(networkName)
	if err != nil {
		return nil, err
	}

	intSupport.GhMockContainer, err = intSupport.startGithubApiHttpMock(networkName)
	if err != nil {
		return nil, err
	}

	GhMockContainerName, err := intSupport.GhMockContainer.Name(ctx)
	if err != nil {
		return nil, err
	}

	var appEnvVars = map[string]string{
		"GITHUB_CHANGELOG_FETCH_TIMER_SECONDS": "5",
		"GITHUB_REPO_FETCH_TIMER_SECONDS":      "30",
		"GITHUB_URL_DEFAULT":                   "http://" + GhMockContainerName[1:] + ":3000/api-v3/",
		"GITHUB_URL_UPLOAD":                    "http://" + GhMockContainerName[1:] + ":3000/",
		"LOGGING_LEVEL":                        "debug",
	}

	intSupport.AppContainer, err = intSupport.startApplication(networkName, appEnvVars)
	if err != nil {
		return nil, err
	}

	return &intSupport, nil
}

func (i *integrationSupport) createTestNetwork(networkName string) (testcontainers.Network, error) {
	var err error
	ctx := context.Background()
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           i.parseName(networkName),
			CheckDuplicate: false,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create test network: %s", err)
	}

	return network, nil
}

func (i *integrationSupport) parseName(name string) string {
	return i.ProjectName + "-" + name
}

func (i *integrationSupport) startApplication(networkName string, envVars map[string]string) (testcontainers.Container, error) {
	var err error

	containerName := i.parseName("app")
	ctx := context.Background()
	projectPath := testsupport.GetProjectPath("..")

	req := testcontainers.ContainerRequest{
		Env:          envVars,
		ExposedPorts: []string{"8080/tcp"},
		FromDockerfile: testcontainers.FromDockerfile{
			Context: projectPath,
		},
		Hostname:   containerName,
		Name:       containerName,
		Networks:   []string{i.parseName(networkName)},
		WaitingFor: wait.ForLog("Dashboard changelog repo data refreshed"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to start app container: %s", err)
	}

	return container, nil
}

func (i *integrationSupport) startGithubApiHttpMock(networkName string) (testcontainers.Container, error) {
	var err error

	containerName := i.parseName("gh-mock")
	ctx := context.Background()
	projectPath := testsupport.GetProjectPath("..")

	req := testcontainers.ContainerRequest{
		BindMounts: map[string]string{projectPath + "/testsupport/fixtures/github/": "/config"},
		Cmd:        []string{"-config", "/config/killgrave.config.yml"},
		Hostname:   containerName,
		Image:      "friendsofgo/killgrave:0.4.0",
		Name:       containerName,
		Networks:   []string{i.parseName(networkName)},
		WaitingFor: wait.ForLog("The fake server is on tap now"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to start GH Mock container: %s", err)
	}

	return container, nil
}

func (i *integrationSupport) teardown() error {
	if i.AppContainer != nil {
		err := i.AppContainer.Terminate(context.Background())
		if err != nil {
			return err
		}
		i.AppContainer = nil
	}
	if i.GhMockContainer != nil {
		err := i.GhMockContainer.Terminate(context.Background())
		if err != nil {
			return err
		}
		i.GhMockContainer = nil
	}
	if i.Network != nil {
		err := i.Network.Remove(context.Background())
		if err != nil {
			return err
		}
		i.Network = nil
	}
	return nil
}
