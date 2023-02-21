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
	"github.com/containers/podman/v4/pkg/specgen"
)

// Object handling containers creation and retrieving.
// See Container for container-specific operations
type Manager struct {
	ctx        *context.Context
	socketPath string
	log        *log.Logger
	images     map[string]string
}

// Option when creating a Manager
type ManagerOptions struct {
	SocketPath string
	Logger     *log.Logger
	// Images allowed. The key is the identifier, value is full name (e.g. docker.io/library/busybox)
	AllowedImages map[string]string
}

const (
	// Base name for all labels
	L_BASE = "studentbox"

	// Is managed by this library
	L_IS_OWNED = L_BASE + ".container"

	// Which user is this container owned by
	L_USER = L_BASE + ".user"

	// The project associated to the user
	L_PROJECT = L_BASE + ".project"
)

func DefaultManagerOptions() *ManagerOptions {
	sockDir := os.Getenv("XDG_RUNTIME_DIR")
	if sockDir == "" {
		sockDir = "/tmp"
	}
	return &ManagerOptions{
		SocketPath:    "unix://" + sockDir + "/podman/podman.sock",
		Logger:        log.New(os.Stdout, "[containers] ", log.LstdFlags),
		AllowedImages: map[string]string{},
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

	if len(opt.AllowedImages) == 0 {
		opt.Logger.Println("WARN: no images allowed has been set")
	}

	return &Manager{
		ctx:        &ctx,
		socketPath: opt.SocketPath,
		log:        opt.Logger,
		images:     opt.AllowedImages,
	}, nil
}

func (m *Manager) Containers() ([]*Container, error) {
	cs, err := containers.List(*m.ctx, &containers.ListOptions{
		// Only containers managed by this lib
		Filters: map[string][]string{
			"label": {L_IS_OWNED + "=true"},
		},
	})

	if err != nil {
		return nil, err
	}

	// Convert to []*Container
	containers := make([]*Container, 0, len(cs))
	for _, container := range cs {
		containers = append(containers, NewFromListContainer(*m.ctx, container))
	}

	return containers, nil
}

func (m *Manager) ContainerExists(user, project string) (bool, error) {
	exists, err := containers.Exists(*m.ctx, user+"-"+project, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check if container exists: %w", err)
	}
	return exists, nil
}

func (m *Manager) PullImageIfNotExists(image string) error {
	exists, err := images.Exists(*m.ctx, image, nil)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	quiet := true
	r, err := images.Pull(*m.ctx, image, &images.PullOptions{Quiet: &quiet})
	m.log.Println("INFO: PullImageIfNotExists", r, err)
	return err
}

func (m *Manager) SpawnContainer(user, project, imageName string) error {
	image, ok := m.images[imageName]
	if !ok {
		return errors.New("image " + imageName + " is not in allowed images")
	}

	exists, err := m.ContainerExists(user, project)
	if err != nil {
		// Wrap error
		return err
	}
	if exists {
		return errors.New("container already exists")
	}

	err = m.PullImageIfNotExists(image)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	spec := specgen.NewSpecGenerator(image, false)
	spec.Terminal = true
	spec.Name = user + "-" + project
	spec.Labels = map[string]string{
		L_IS_OWNED: "true",
		L_USER:     user,
		L_PROJECT:  project,
	}

	r, err := containers.CreateWithSpec(*m.ctx, spec, nil)
	if err != nil {
		return err
	}

	m.log.Println("INFO: Created container", r.ID)

	if len(r.Warnings) > 0 {
		m.log.Println("WARN:", r.Warnings)
	}

	err = containers.Start(*m.ctx, r.ID, nil)
	if err != nil {
		m.log.Println("ERROR: Failed to start container, rolling back: ", err)
		force := true
		containers.Remove(*m.ctx, r.ID, &containers.RemoveOptions{Force: &force})
		return err
	}
	return nil
}

// Get a container by name of user and project
// might return nil
func (m *Manager) GetContainer(user, project string) (*Container, error) {
	exists, err := m.ContainerExists(user, project)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrContainerDontExists{User: user, Project: project}
	}

	return &Container{
		ctx:     *m.ctx,
		Name:    user + "-" + project,
		User:    user,
		Project: project,
	}, nil
}
