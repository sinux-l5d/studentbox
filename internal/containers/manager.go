package containers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/bindings/pods"
	"github.com/containers/podman/v4/pkg/domain/entities"
	"github.com/containers/podman/v4/pkg/specgen"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// Object handling containers creation and retrieving.
// See Container for container-specific operations
type Manager struct {
	ctx        *context.Context
	socketPath string
	hostPath   string
	dataPath   string
	log        *log.Logger
	images     map[string]string
}

// Option when creating a Manager
type ManagerOptions struct {
	SocketPath string
	// Alsolute path of parent directory of DataPath on host (e.g. /home/studentbox/)
	HostPath string
	// Relative path of data directory (e.g. data/)
	// Note that HostPath + DataPath is the absolute path of data directory on host
	DataPath string
	Logger   *log.Logger
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

	// Image-specific config
	L_CONFIG        = L_BASE + ".config"
	L_CONFIG_MOUNTS = L_CONFIG + ".mounts"

	// Prefix for objects created by this library
	PREFIX = "sb-"
)

func DefaultManagerOptions() *ManagerOptions {
	sockDir := os.Getenv("XDG_RUNTIME_DIR")
	if sockDir == "" {
		sockDir = "/tmp"
	}
	return &ManagerOptions{
		SocketPath:    "unix://" + sockDir + "/podman/podman.sock",
		DataPath:      "./data",
		Logger:        log.New(os.Stdout, "[containers] ", log.LstdFlags),
		AllowedImages: map[string]string{},
	}
}

func NewManager(opt *ManagerOptions) (*Manager, error) {
	// Default options
	if opt == nil {
		return nil, &ParameterRequired{ParamName: "opt"}
	}

	ctx := context.Background()
	ctx, err := bindings.NewConnection(ctx, opt.SocketPath)
	if err != nil {
		return nil, err
	}

	if len(opt.AllowedImages) == 0 {
		opt.Logger.Println("WARN: no images allowed has been set")
	}

	// test opt.DataPath is a writable directory, create it if don't exist
	if _, err := os.Stat(opt.DataPath); os.IsNotExist(err) {
		err = os.MkdirAll(opt.DataPath, 0755)
		if err != nil {
			return nil, err
		}
	}

	if opt.HostPath == "" || !path.IsAbs(opt.HostPath) {
		return nil, errors.New("host path must be an absolute path, current value: \"" + opt.HostPath + "\"")
	}

	return &Manager{
		ctx:        &ctx,
		socketPath: opt.SocketPath,
		log:        opt.Logger,
		hostPath:   opt.HostPath,
		dataPath:   opt.DataPath,
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

// Check image is allowed and pull it if not exists locally
func (m *Manager) PullImageIfNotExists(image string) error {
	image, ok := m.images[image]
	if !ok {
		return errors.New("image " + image + " is not in allowed images")
	}

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
	exists, err := m.ContainerExists(user, project)
	if err != nil {
		// Wrap error
		return err
	}
	if exists {
		return errors.New("container already exists")
	}

	err = m.PullImageIfNotExists(imageName)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	spec := specgen.NewSpecGenerator(imageName, false)
	spec.Terminal = true
	spec.Name = PREFIX + user + "-" + project
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

func (m Manager) toHostPath(relativePath string) string {
	return filepath.Join(m.hostPath, m.dataPath, relativePath)
}

type PodContainerOptions struct {
	Image   string            // image name
	Mounts  []string          // absolute path in container
	Envvars map[string]string // envvar name and value
}

type PodOptions struct {
	User       string
	Project    string
	Containers []PodContainerOptions
}

func (opt PodContainerOptions) toContainerSpec(toProjectHostPath func(string) string) *specgen.SpecGenerator {
	spec := specgen.NewSpecGenerator(opt.Image, false)
	spec.Terminal = true
	spec.Name = opt.Image
	spec.Labels = map[string]string{
		L_IS_OWNED: "true",
	}
	spec.Env = opt.Envvars

	for _, volume := range opt.Mounts {
		// May break if two mounts have the same directory name
		dirName := filepath.Base(volume)
		spec.Mounts = append(spec.Mounts, specs.Mount{
			Destination: volume,
			Source:      toProjectHostPath(dirName),
			Type:        "bind",
		})
	}

	return spec
}

func (m *Manager) SpawnPod(opt *PodOptions) error {
	podSpecGen := specgen.NewPodSpecGenerator()
	podSpecGen.Name = PREFIX + opt.User + "-" + opt.Project
	podSpecGen.Labels = map[string]string{
		L_IS_OWNED: "true",
		L_USER:     opt.User,
		L_PROJECT:  opt.Project,
	}

	podSpec := entities.PodSpec{
		PodSpecGen: *podSpecGen,
	}

	podCreateResponse, err := pods.CreatePodFromSpec(*m.ctx, &podSpec)
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	m.log.Printf("INFO: Created pod %s", podCreateResponse.Id)

	for _, container := range opt.Containers {
		err = m.SpawnContainerInPod(podCreateResponse.Id, &container, fmt.Sprintf("%s-%s-%s", PREFIX+opt.User, opt.Project, container.Image))
		if err != nil {
			force := true
			m.log.Printf("ERROR: Failed to spawn container in pod, removing pod: %s", err)
			pods.Remove(*m.ctx, podCreateResponse.Id, &pods.RemoveOptions{Force: &force})
			return fmt.Errorf("failed to spawn container in pod: %w", err)
		}
	}

	return nil
}

func (m *Manager) SpawnContainerInPod(podID string, containerOptions *PodContainerOptions, containerName string) error {
	err := m.PullImageIfNotExists(containerOptions.Image)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	spec := containerOptions.toContainerSpec(func(path string) string {
		return m.toHostPath(filepath.Join(containerName, path))
	})
	spec.Pod = podID

	r, err := containers.CreateWithSpec(*m.ctx, spec, nil)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	m.log.Printf("INFO: Created container %s in pod %s", r.ID, podID)

	if len(r.Warnings) > 0 {
		m.log.Println("WARN:", r.Warnings)
	}

	err = containers.Start(*m.ctx, r.ID, nil)
	if err != nil {
		m.log.Printf("ERROR: Failed to start container %s in pod %s, rolling back: %v", r.ID, podID, err)
		force := true
		containers.Remove(*m.ctx, r.ID, &containers.RemoveOptions{Force: &force})
		return err
	}

	return nil
}
