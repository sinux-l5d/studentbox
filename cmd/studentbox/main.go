package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sinux-l5d/studentbox/internal/containers"
	"github.com/sinux-l5d/studentbox/internal/runtimes"
)

var (
	// global flags
	socket   string
	hostPath string
	version  = "dev"
)

func newManager(w io.Writer) (*containers.Manager, error) {
	opt := containers.DefaultManagerOptions()
	opt.SocketPath = socket
	if hostPath != "" {
		opt.HostPath = hostPath
	}
	// get abs current dir
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
			&cli.StringFlag{
				Name:        "hostpath",
				Usage:       "Path to the host's root directory. Used mostly when ran in a container",
				Value:       containers.DefaultManagerOptions().HostPath,
				EnvVars:     []string{"HOSTPATH"},
				Destination: &hostPath,
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

					containers, err := manager.GetAllContainers()
					if err != nil {
						return err
					}

					if len(containers) == 0 {
						fmt.Fprintln(c.App.Writer, "No containers")
						return nil
					}

					fmt.Fprintln(c.App.Writer, "Containers:")
					for _, container := range containers {
						ip, port, err := container.GetPort()
						if err != nil || (ip == "" && port == "") {
							fmt.Fprintf(c.App.Writer, "- (%s/%s) %s \n", container.User, container.Project, container.Name)
						} else {
							fmt.Fprintf(c.App.Writer, "- (%s/%s) %s -> %s:%s\n", container.User, container.Project, container.Name, ip, port)
						}
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Print status of a project's runtime",
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
					cntnrs, err := manager.GetContainers(user, project)
					if err != nil {
						if errors.Is(err, &containers.ErrContainerDontExists{}) {
							fmt.Fprintf(c.App.ErrWriter, "container for user %s, project %s doesn't exist\n", user, project)
						} else {
							return err
						}
					}

					status := make(map[string]string)
					for _, container := range cntnrs {
						s, err := container.Status()
						if err != nil {
							return err
						}
						status[container.Name] = s
					}

					fmt.Fprintf(c.App.Writer, "Status of project %s/%s:\n", user, project)
					for name, s := range status {
						fmt.Fprintf(c.App.Writer, "- %s: %s\n", name, s)
					}

					return nil
				},
			},
			{
				Name:  "spawn",
				Usage: "Spawn a runtime (pod of container) for a project",
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
						Name:     "runtime",
						Aliases:  []string{"r"},
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:    "env",
						Aliases: []string{"e"},
						Usage:   "Set environment variables (e.g. -e FOO=bar)",
						Action: func(_ *cli.Context, v []string) error {
							for _, value := range v {
								if !strings.Contains(value, "=") || strings.HasPrefix(value, "=") || strings.HasSuffix(value, "=") {
									return fmt.Errorf("invalid environment variable %s", value)
								}
							}
							return nil
						},
					},
				},
				Action: func(c *cli.Context) error {
					manager, err := newManager(c.App.Writer)
					if err != nil {
						return err
					}

					runtime, exists := runtimes.OfficialRuntimes[c.String("runtime")]
					if !exists {
						return fmt.Errorf("runtime %s doesn't exist", c.String("runtime"))
					}

					envvar := make(map[string]string, len(c.StringSlice("env")))
					for _, value := range c.StringSlice("env") {
						split := strings.SplitN(value, "=", 2)
						if len(split) != 2 {
							return fmt.Errorf("invalid environment variable %s", value)
						}
						envvar[split[0]] = split[1]
					}

					opt := containers.PodOptions{
						User:         c.String("user"),
						Project:      c.String("project"),
						InputEnvVars: envvar,
						Runtime:      runtime,
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
			{
				Name: "envs",
				Usage: "Print environment variables of a project's runtime",
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
						Name: "container",
						Aliases: []string{"c"},
						Usage: "Container partial name (e.g. -c mysql)",
						Required: false,
					},
					&cli.StringFlag{
						Name: "query",
						Aliases: []string{"q"},
						Usage: "Query environment variables (e.g. -q FOO)",
						Required: false,
					},
				},
				Action: func(c *cli.Context) error {
					manager, err := newManager(c.App.Writer)
					if err != nil {
						return err
					}

					user, project := c.String("user"), c.String("project")
					cntnrs, err := manager.GetEnvVars(user, project)
					if err != nil {
						return err
					}

					if c.String("container") != "" {
						for k := range cntnrs {
							if !strings.Contains(k, c.String("container")) {
								delete(cntnrs, k)
							}
						}
					}

					if c.String("query") != "" {
						// v is a map[string]string
						// check if query is a partial, lowercase match
						for _, v := range cntnrs {
							for k := range v {
								if !strings.Contains(strings.ToLower(k), strings.ToLower(c.String("query"))) {
									delete(v, k)
								}
							}
						}
					}

					if len(cntnrs) == 0 {
						fmt.Fprintf(c.App.Writer, "No environment variables found\n")
						return nil
					}

					fmt.Fprintf(c.App.Writer, "Environment variables of project %s/%s:\n", user, project)
					for name, envvars := range cntnrs {
						if len(envvars) == 0 {
							continue
						}
						fmt.Fprintf(c.App.Writer, "- %s:\n", name)
						for k, v := range envvars {
							fmt.Fprintf(c.App.Writer, "  - %s: %s\n", k, v)
						}
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(app.Writer, "Error: %s\n", err)
		os.Exit(1)
	}
}
