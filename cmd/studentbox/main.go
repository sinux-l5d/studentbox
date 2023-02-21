package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sinux-l5d/studentbox/internal/containers"
)

var (
	// global flags
	socket string
)

func main() {
	app := &cli.App{
		Name:     "studentbox",
		Usage:    "Manage Studentbox's containers from the CLI",
		Version:  "v0.1.0",
		Compiled: time.Now(),
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
				Name:  "list",
				Usage: "List containers belonging to Studentbox",
				Action: func(c *cli.Context) error {
					manager, err := containers.NewManager(&containers.ManagerOptions{
						SocketPath: socket,
						Logger:     log.New(c.App.Writer, "", log.Flags()),
					})
					if err != nil {
						return err
					}

					containers, err := manager.Containers()
					if err != nil {
						return err
					}
					if len(containers) == 0 {
						fmt.Fprintln(c.App.Writer, "No containers")
					} else {
						for _, container := range containers {
							fmt.Fprintf(c.App.Writer, "%s/%s\n", container.User, container.Project)
						}
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(app.Writer, "Error: %s", err)
		os.Exit(1)
	}
}
