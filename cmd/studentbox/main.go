package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/sinux-l5d/studentbox/internal/containers"
	"github.com/sinux-l5d/studentbox/internal/runtimes"
)

var (
	// global flags
	socket  string
	version = "dev"
)

func newManager(w io.Writer) (*containers.Manager, error) {
	opt := containers.DefaultManagerOptions()
	opt.SocketPath = socket
	// get abs current dir
	pwd, _ := os.Getwd()
	opt.HostPath = pwd
	if w == nil {
		opt.Logger = log.New(io.Discard, "", log.Flags())
	} else {
		opt.Logger = log.New(w, "[containers] ", log.Flags())
	}
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
					manager, err := newManager(nil)
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
					manager, err := newManager(nil)
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
				},
				Action: func(c *cli.Context) error {
					// manager, err := newManager(c.App.Writer, map[string]string{"busybox": c.String("image")})
					manager, err := newManager(c.App.Writer)
					if err != nil {
						return err
					}

					opt := containers.PodOptions{
						User:    c.String("user"),
						Project: c.String("project"),
						Runtime: runtimes.OfficialRuntimes["lamp"],
						// Runtime: runtimes.Runtime{
						// 	Name: "dummy",
						// 	Images: map[string]runtimes.Image{
						// 		"busybox": {
						// 			ShortName:          "busybox",
						// 			FullyQualifiedName: "docker.io/library/busybox",
						// 			Mounts: map[string]string{
						// 				"dummy-host": "/dummy-container",
						// 			},
						// 		},
						// 	},
						// },
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
