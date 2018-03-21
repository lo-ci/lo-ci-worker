package sandbox

import "io"

type ContainerFactory interface {
	WithName(name string) ContainerFactory
	WithImage(image string) ContainerFactory
	WithCommand(command io.Reader) ContainerFactory
	Make() (Container, error)
}
