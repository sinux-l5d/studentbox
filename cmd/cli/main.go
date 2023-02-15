package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sinux-l5d/studentbox/internal/containers"
)

var (
	// global flags
	socket = flag.String("socket", containers.DefaultManagerOptions().SocketPath, "Socket to use for communication with the podman daemon")
)

func main() {
	log := log.New(os.Stdout, "[cli] ", log.LstdFlags)
	flag.Parse()

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	spawnCmd := flag.NewFlagSet("spawn", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("subcommand is required")
		os.Exit(1)
	}

	var fn func(*containers.Manager) error

	switch os.Args[1] {
	case "list":
		listCmd.Parse(os.Args[2:])
		fn = func(m *containers.Manager) error {
			cs, err := m.Containers()
			if err != nil {
				return err
			}
			if len(cs) == 0 {
				fmt.Println("No containers found")
				return nil
			} else {
				fmt.Printf("Found %d containers:\n", len(cs))
			}
			for _, c := range cs {
				fmt.Printf("%#v\n", c)
			}
			return nil
		}
	case "spawn":
		spawnCmd.Parse(os.Args[2:])
		fn = func(m *containers.Manager) error {
			return m.SpawnContainer("sinux", "test")
		}
	default:
		fmt.Printf("Unknown subcommand %q", os.Args[1])
		os.Exit(1)
	}

	manager, err := containers.NewManager(&containers.ManagerOptions{
		SocketPath: *socket,
		Logger: log,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = fn(manager)
	if err != nil {
		log.Fatal(err)
	}

}