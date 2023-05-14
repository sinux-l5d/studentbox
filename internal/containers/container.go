package containers

import (
	"context"
	"strings"

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

func (c *Container) GetEnv() (map[string]string, error) {
	inspect, err := containers.Inspect(c.ctx, c.Name, &containers.InspectOptions{})
	if err != nil {
		return nil, err
	}

	// Convert list to map
	env := make(map[string]string)
	for _, e := range inspect.Config.Env {
		parts := strings.SplitN(e,"=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	return env, nil
}

// Return the HostIP and HostPort of the container
// Works only on infra containers, return empty strings if not found
func (c *Container) GetPort() (string, string, error) {
	inspect, err := containers.Inspect(c.ctx, c.Name, &containers.InspectOptions{})
	if err != nil {
		return "", "", err
	}

	config, ok := inspect.HostConfig.PortBindings["80/tcp"]
	if !ok {
		return "", "", nil
	}
	return config[0].HostIP, config[0].HostPort, nil
}