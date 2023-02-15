package containers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/domain/entities"
	"github.com/containers/podman/v4/pkg/specgen"
)

type Manager struct {
	ctx context.Context
	socketPath string
	log *log.Logger
}

type ManagerOptions struct {
	SocketPath string
	Logger *log.Logger
}

func DefaultManagerOptions() *ManagerOptions {
	sockDir := os.Getenv("XDG_RUNTIME_DIR")
	if sockDir == "" {
		sockDir = "/tmp"
	}
	return &ManagerOptions{
		SocketPath: "unix://" + sockDir + "/podman/podman.sock",
		Logger: log.New(os.Stdout, "[containers/manager] ", log.LstdFlags),
	}
}

func NewManager(opt *ManagerOptions) (*Manager, error) {
	// Default options
	if opt == nil {
		return nil, errors.New("ManagerOption is required")
	}

	ctx := context.Background()
	ctx, err := bindings.NewConnection(ctx, opt.SocketPath)
	if err != nil {
		return nil, err
	}

	return &Manager{
		ctx: ctx,
		socketPath: opt.SocketPath,
		log: opt.Logger,
	}, nil
}

func (m *Manager) Containers() ([]entities.ListContainer, error) {
	cs, err := containers.List(m.ctx, &containers.ListOptions{
		// filter label studentbox.container=true
		Filters: map[string][]string{
			"label": {"studentbox.container=true"},
		},
	})

	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (m *Manager) ContainerExists(user, project string) (bool, error) {
	return containers.Exists(m.ctx, user + "-" + project, nil)
}

func (m *Manager) PullImageIfNotExists(image string) error {
	exists, err := images.Exists(m.ctx, image, nil)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	r, err  := images.Pull(m.ctx, image, &images.PullOptions{})
	m.log.Println("PullImageIfNotExists", r, err)
	return err
}

func (m *Manager) SpawnContainer(user, project string) error {

	exists, err := m.ContainerExists(user, project)
	if err != nil {
		// Wrap error
		return fmt.Errorf("failed to check if container exists: %w", err)
	}
	if exists {
		return errors.New("container already exists")
	}

	err = m.PullImageIfNotExists("docker.io/library/busybox")
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	spec := specgen.NewSpecGenerator("docker.io/library/busybox", false)
	spec.Terminal = true
	spec.Name = user + "-" + project
	spec.Labels = map[string]string{
		"studentbox.container": "true",
		"studentbox.user": user,
		"studentbox.project": project,
	}

	r, err := containers.CreateWithSpec(m.ctx, spec, nil)
	if err != nil {
		return err
	}

	m.log.Println("INFO: Created container", r.ID)

	if len(r.Warnings) > 0 {
		m.log.Println("WARN:", r.Warnings)
	}

	err = containers.Start(m.ctx, r.ID, nil)
	if err != nil {
		return err
	}
	return nil
}