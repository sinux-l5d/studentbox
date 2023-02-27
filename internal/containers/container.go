package containers

import (
	"context"

	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/domain/entities"
)

type Container struct {
	// Contect from bindings.NewConnection to execute operations
	ctx     context.Context
	Name    string
	User    string
	Project string
}

func NewFromListContainer(ctx context.Context, container entities.ListContainer) *Container {
	return &Container{
		Name:    container.Names[0],
		User:    container.Labels[L_USER],
		Project: container.Labels[L_PROJECT],
		ctx:     ctx,
	}
}

func (c *Container) Status() (string, error) {
	inspect, err := containers.Inspect(c.ctx, c.Name, &containers.InspectOptions{})
	if err != nil {
		return "", err
	}
	return inspect.State.Status, nil
}
