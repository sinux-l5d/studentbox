package containers

import "fmt"

type ErrContainerDontExists struct {
	User    string
	Project string
}

func (e *ErrContainerDontExists) Error() string {
	return fmt.Sprintf("container %s-%s don't exists", e.User, e.Project)
}
