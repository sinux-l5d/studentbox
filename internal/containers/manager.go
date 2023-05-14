package containers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/containers/common/libnetwork/types"
	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/bindings/pods"
	"github.com/containers/podman/v4/pkg/domain/entities"
	"github.com/containers/podman/v4/pkg/specgen"
	"github.com/sinux-l5d/studentbox/internal/runtimes"
	"github.com/sinux-l5d/studentbox/internal/tools"
)

// Object handling containers creation and retrieving.
// See Container for container-specific operations
type Manager struct {
	ctx        *context.Context
	socketPath string
	hostPath   string
	dataPath   string
	log        *log.Logger
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
	pwd, _ := os.Getwd()
	if sockDir == "" {
		sockDir = "/tmp"
	}
	return &ManagerOptions{
		SocketPath: "unix://" + sockDir + "/podman/podman.sock",
		DataPath:   "./data",
		HostPath:   pwd,
		Logger:     log.New(os.Stdout, "[containers] ", log.LstdFlags),
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

	// test opt.DataPath is a writable directory, create it if don't exist
	// Relative because can be in container
	tools.EnsureDirCreated(opt.DataPath)

	if opt.HostPath == "" || !path.IsAbs(opt.HostPath) {
		return nil, errors.New("host path must be an absolute path, current value: \"" + opt.HostPath + "\"")
	}

	return &Manager{
		ctx:        &ctx,
		socketPath: opt.SocketPath,
		log:        opt.Logger,
		hostPath:   opt.HostPath,
		dataPath:   opt.DataPath,
	}, nil
}

func (m *Manager) GetAllContainers() ([]*Container, error) {
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

func (m *Manager) PodExists(user, project string) (bool, error) {
	exists, err := pods.Exists(*m.ctx, PREFIX+user+"-"+project, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check if container exists: %w", err)
	}
	return exists, nil
}

// Check image is allowed and pull it if not exists locally
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

// Get a container by name of user and project
// might return nil
func (m *Manager) GetContainers(user, project string) ([]*Container, error) {
	exists, err := m.PodExists(user, project)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrContainerDontExists{User: user, Project: project}
	}

	// Get all containers
	cs, err := containers.List(*m.ctx, &containers.ListOptions{
		Filters: map[string][]string{
			"label": {L_IS_OWNED + "=true", L_USER + "=" + user, L_PROJECT + "=" + project},
		},
	})
	if err != nil {
		return nil, err
	}

	// Convert to []*Container
	containers := make([]*Container, len(cs))
	for i, container := range cs {
		containers[i] = NewFromListContainer(*m.ctx, container)
	}

	return containers, nil
}

func (m Manager) toHostPath(relativePath string) string {
	return filepath.Join(m.hostPath, m.dataPath, relativePath)
}

type PodOptions struct {
	User    string
	Project string
	// Environment variables to pass to ALL containers
	InputEnvVars map[string]string
	Runtime      runtimes.Runtime
}

func (m *Manager) SpawnPod(opt *PodOptions) error {
	podSpecGen := specgen.NewPodSpecGenerator()
	podSpecGen.Name = PREFIX + opt.User + "-" + opt.Project
	podSpecGen.Labels = map[string]string{
		L_IS_OWNED: "true",
		L_USER:     opt.User,
		L_PROJECT:  opt.Project,
	}
	podSpecGen.PortMappings = append(podSpecGen.PortMappings, types.PortMapping{ContainerPort: 80})

	podSpec := entities.PodSpec{
		PodSpecGen: *podSpecGen,
	}

	podCreateResponse, err := pods.CreatePodFromSpec(*m.ctx, &podSpec)
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	m.log.Printf("INFO: Created pod %s", podCreateResponse.Id)

	for _, image := range opt.Runtime.Images {
		err = m.SpawnContainerInPod(podCreateResponse.Id, &image, opt.InputEnvVars, fmt.Sprintf("%s-%s-%s", PREFIX+opt.User, opt.Project, image.ShortName), opt.User, opt.Project)
		if err != nil {
			force := true
			m.log.Printf("ERROR: Failed to spawn container in pod, removing pod: %s", err)
			pods.Remove(*m.ctx, podCreateResponse.Id, &pods.RemoveOptions{Force: &force})
			return fmt.Errorf("failed to spawn container in pod: %w", err)
		}
	}

	return nil
}

func (m *Manager) SpawnContainerInPod(podID string, img *runtimes.Image, inputEnvVar map[string]string, containerName string, user string, project string) error {
	err := m.PullImageIfNotExists(img.FullyQualifiedName)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	relativeProjectDir := filepath.Join(user, project)

	// create specs with project directory
	spec, err := img.ToContainerSpec(m.toHostPath(relativeProjectDir), inputEnvVar)
	if err != nil {
		return fmt.Errorf("failed to create container spec: %w", err)
	}
	spec.Pod = podID
	spec.Name = containerName
	spec.Labels = map[string]string{
		L_IS_OWNED: "true",
		L_USER:     user,
		L_PROJECT:  project,
	}

	// be sure that all mounts are created
	for name := range img.Mounts {
		err := tools.EnsureDirCreated(filepath.Join(m.dataPath, relativeProjectDir, name))
		if err != nil {
			return fmt.Errorf("failed to create dir for mount: %w", err)
		}
	}

	// create container
	r, err := containers.CreateWithSpec(*m.ctx, spec, nil)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	m.log.Printf("INFO: Created container %s in pod %s", r.ID, podID)

	if len(r.Warnings) > 0 {
		m.log.Println("WARN:", r.Warnings)
	}

	// start container
	err = containers.Start(*m.ctx, r.ID, nil)
	if err != nil {
		m.log.Printf("ERROR: Failed to start container %s in pod %s, rolling back: %v", r.ID, podID, err)
		force := true
		containers.Remove(*m.ctx, r.ID, &containers.RemoveOptions{Force: &force})
		return err
	}

	return nil
}

// Return a map with key: container name, value: map of env vars
func (m *Manager) GetEnvVars(user, project string) (map[string]map[string]string, error) {
	containers, err := m.GetContainers(user, project)
	if err != nil {
		return nil, err
	}
	envVars := make(map[string]map[string]string)
	for _, container := range containers {
		envs, err := container.GetEnv()
		if err != nil {
			return nil, err
		}
		envVars[container.Name] = envs
	}

	return envVars, nil
}