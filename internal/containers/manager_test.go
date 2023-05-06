package containers_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/sinux-l5d/studentbox/internal/containers"
	"github.com/sinux-l5d/studentbox/internal/runtimes"
)

func newManager() (*containers.Manager, error) {
	opt := containers.DefaultManagerOptions()
	// get abs current dir
	pwd, _ := os.Getwd()
	opt.HostPath = pwd
	return containers.NewManager(opt)
}

func BenchmarkPodSpawn(b *testing.B) {
	runtime := runtimes.OfficialRuntimes["lamp"]
	manager, _ := newManager()
	opt := containers.PodOptions{
		User: "student",
		Runtime: runtime,
	}
	for i := 0; i < b.N; i++ {
		opt.Project = fmt.Sprintf("benchmark-%d", i)
		_ = manager.SpawnPod(&opt)
	}
}