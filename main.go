package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/lo-ci-worker/sandbox"

	"docker.io/go-docker"
)

const containerName = "sandbox-container-name-goes-here"

func execute(commands io.Reader) {
	client, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	factory := sandbox.DockerContainerFactory{}
	container, err := factory.
		SetClient(client).
		WithName(containerName).
		WithImage("alpine").
		WithCommand(commands).
		Make()

	if err != nil {
		panic(err)
	}

	err = container.Start()
	if err != nil {
		panic(err)
	}

	defer container.Dispose()

	logs, err := container.WatchLogs()
	if err != nil {
		panic(err)
	}

	done, issues := container.Wait()

eventloop:
	for {
		select {
		case logLine := <-logs:
			fmt.Printf("- %s\n", logLine)
		case issue := <-issues:
			panic(issue)
		case result := <-done:
			fmt.Printf("\nExit Code: %d", result.ExitCode)
			break eventloop
		}
	}
}

func main() {
	commands := strings.NewReader(strings.Join([]string{
		"apk add --update nodejs nodejs-npm",
		"node -e \"console.log('Hello world!');\"",
	}, "\n"))

	execute(commands)
}
