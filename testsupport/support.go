package testsupport

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/google/go-github/github"
)

func GetProjectPath(relativePath string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	projectPath, _ := filepath.Abs(filepath.Join(basepath, relativePath))

	return projectPath
}

func setupGithubApiHttpMock() (func(), error) {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	projectPath, _ := filepath.Abs(filepath.Join(basepath, ".."))

	cmd := exec.Command("killgrave", "-config", projectPath+"/testsupport/fixtures/github/killgrave.config.yml")
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	success := waitTcpPort("localhost:3000")
	if !success {
		return nil, fmt.Errorf("Could not connect on localhost:3000")
	}

	teardown := func() {
		_ = syscall.Kill(cmd.Process.Pid, 2)
	}

	return teardown, nil
}

func SetupGithubClientMock() (client *github.Client, teardown func()) {
	mockGhApiTeardown, err := setupGithubApiHttpMock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting GH API HTTP Mock: %v\n", err)
		os.Exit(1)
	}

	client = github.NewClient(nil)
	url, _ := url.Parse("http://localhost:3000/api-v3/")
	client.BaseURL = url
	client.UploadURL = url

	teardown = func() {
		mockGhApiTeardown()
	}

	return client, teardown
}

func waitTcpPort(host string) bool {
	retry := 10
	for retry > 0 {
		timeout := time.Duration(1) * time.Second
		conn, err := net.DialTimeout("tcp", host, timeout)
		if err == nil && conn != nil {
			conn.Close()
			return true
		}
		time.Sleep(timeout)
		retry--
	}

	return false
}
