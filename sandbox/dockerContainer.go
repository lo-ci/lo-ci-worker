package sandbox

import (
	"bufio"
	"context"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
)

type DockerContainer struct {
	ID     string
	client *docker.Client
}

func NewDockerContainer(client *docker.Client, id string) DockerContainer {
	con := DockerContainer{}
	con.client = client
	con.ID = id
	return con
}

func (con DockerContainer) Start() error {
	return con.client.ContainerStart(context.Background(), con.ID, types.ContainerStartOptions{})
}

func (con DockerContainer) Wait() (<-chan DoneResponse, <-chan error) {
	doneChannel := make(chan DoneResponse)
	errorChannel := make(chan error)

	go func() {
		stats, errors :=
			con.client.ContainerWait(context.Background(), con.ID, container.WaitConditionNotRunning)

		select {
		case err := <-errors:
			errorChannel <- err
		case status := <-stats:
			doneChannel <- DoneResponse{ExitCode: status.StatusCode}
		}
	}()

	return doneChannel, errorChannel
}

func (con DockerContainer) WatchLogs() (<-chan string, error) {
	output := make(chan string)
	stream, err := con.client.ContainerLogs(context.Background(), con.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	})
	if err != nil {
		return nil, err
	}

	go func() {
		defer stream.Close()
		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			output <- scanner.Text()[8:]
		}
	}()

	return output, nil
}

func (con DockerContainer) Dispose() {
	con.client.ContainerRemove(context.Background(), con.ID,
		types.ContainerRemoveOptions{Force: true})
}
