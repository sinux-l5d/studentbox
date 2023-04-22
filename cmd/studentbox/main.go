package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/sinux-l5d/studentbox/internal/containers"
)

var (
	// global flags
	socket  string
	version = "dev"
)

func newManager(_ io.Writer, allowedImages map[string]string) (*containers.Manager, error) {
	if allowedImages == nil {
		allowedImages = map[string]string{}
	}

	opt := containers.DefaultManagerOptions()
	opt.SocketPath = socket
	opt.AllowedImages = allowedImages
	// get abs current dir
	pwd, _ := os.Getwd()
	opt.HostPath = pwd
	opt.Logger = log.New(io.Discard, "", log.Flags())
	return containers.NewManager(opt)
}

// func (app *cli.App) Printf(format string, a ...any) (int, error) {
// 	return fmt.Fprintf(app.Writer, format+"\n", a...)
// }

func main() {
	app := &cli.App{
		Name:    "studentbox",
		Usage:   "Manage Studentbox's containers from the CLI",
		Version: version,
		Authors: []*cli.Author{
			{
				Name:  "Simon Leonard",
				Email: "git-1001af4@sinux.sh",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "socket",
				Usage:       "Socket to use for communication with the podman daemon",
				Aliases:     []string{"s"},
				Value:       containers.DefaultManagerOptions().SocketPath,
				Destination: &socket,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List containers belonging to Studentbox",
				Action: func(c *cli.Context) error {
					manager, err := newManager(c.App.Writer, nil)
					if err != nil {
						return err
					}

					containers, err := manager.Containers()
					if err != nil {
						return err
					}

					if len(containers) == 0 {
						fmt.Fprintln(c.App.Writer, "No containers")
						return nil
					}

					fmt.Fprintln(c.App.Writer, "Containers (user/project):")
					for _, container := range containers {
						fmt.Fprintf(c.App.Writer, "- %s/%s\n", container.User, container.Project)
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Print status of a project's container",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "user",
						Aliases:  []string{"u"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "project",
						Aliases:  []string{"p"},
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					manager, err := newManager(c.App.Writer, nil)
					if err != nil {
						return err
					}
					user, project := c.String("user"), c.String("project")
					container, err := manager.GetContainer(user, project)
					if err != nil {
						if errors.Is(err, &containers.ErrContainerDontExists{}) {
							fmt.Fprintf(c.App.ErrWriter, "container for user %s, project %s doesn't exist\n", user, project)
						} else {
							return err
						}
					}

					status, err := container.Status()
					if err != nil {
						return err
					}

					fmt.Fprintf(c.App.Writer, "Status of %s: %s\n", container.Name, status)

					return nil
				},
			},
			{
				Name:   "spawn-dummy",
				Hidden: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "user",
						Aliases:  []string{"u"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "project",
						Aliases:  []string{"p"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    "image",
						Aliases: []string{"i"},
						Value:   "docker.io/library/busybox",
					},
				},
				Action: func(c *cli.Context) error {
					manager, err := newManager(c.App.Writer, map[string]string{"busybox": c.String("image")})
					if err != nil {
						return err
					}

					opt := containers.PodOptions{
						User:    c.String("user"),
						Project: c.String("project"),
						Containers: []containers.PodContainerOptions{
							{
								Image: "busybox",
								// TODO: get from config or labels
								Mounts: []string{"/studentbox-dummy"},
								Envvars: map[string]string{
									"STUDENTBOX_USER":    c.String("user"),
									"STUDENTBOX_PROJECT": c.String("project"),
								},
							},
						},
					}
					return manager.SpawnPod(&opt)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(app.Writer, "Error: %s\n", err)
		os.Exit(1)
	}
}
