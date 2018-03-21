package sandbox

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
)

var newClient = docker.NewEnvClient

type DockerContainerFactory struct {
	name    string
	image   string
	client  *docker.Client
	command io.Reader
}

func (dcf *DockerContainerFactory) WithName(name string) ContainerFactory {
	dcf.name = name
	return dcf
}

func (dcf *DockerContainerFactory) WithImage(image string) ContainerFactory {
	dcf.image = image
	return dcf
}

func (dcf *DockerContainerFactory) WithCommand(command io.Reader) ContainerFactory {
	dcf.command = command
	return dcf
}

func makeContainer(client *docker.Client, ctx context.Context, name string, image string) (string, error) {
	container, err := client.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   []string{"sh", "/var/tmp/cmd.sh"},
	}, nil, nil, name)
	if err != nil {
		return "", err
	}

	return container.ID, nil
}

func createCommandTar(command io.Reader) (io.Reader, error) {
	writeBuffer := new(bytes.Buffer)
	commandBuffer := new(bytes.Buffer)
	commandBuffer.ReadFrom(command)

	writer := tar.NewWriter(writeBuffer)

	header := &tar.Header{
		Name: "cmd.sh",
		Mode: 0755,
		Size: int64(commandBuffer.Len()),
	}

	err := writer.WriteHeader(header)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(commandBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(writeBuffer.Bytes()), nil
}

func (dcf *DockerContainerFactory) SetClient(client *docker.Client) ContainerFactory {
	dcf.client = client
	return dcf
}

func (dcf *DockerContainerFactory) Make() (Container, error) {
	ctx := context.Background()

	if dcf.client == nil {
		return nil, errors.New("Client must be set first. Call SetClient(client) before Make()")
	}

	_, err := dcf.client.ImagePull(ctx, dcf.image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}

	id, err := makeContainer(dcf.client, ctx, dcf.name, dcf.image)
	if err != nil {
		return nil, err
	}

	container := NewDockerContainer(dcf.client, id)

	file, err := createCommandTar(dcf.command)
	if err != nil {
		container.Dispose()
		return nil, err
	}

	err = dcf.client.CopyToContainer(ctx, container.ID, "/var/tmp", file, types.CopyToContainerOptions{})
	if err != nil {
		container.Dispose()
		return nil, err
	}

	return container, nil
}
