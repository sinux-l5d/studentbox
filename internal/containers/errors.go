package containers

import "fmt"

type ErrContainerDontExists struct {
	User    string
	Project string
}

func (e *ErrContainerDontExists) Error() string {
	return fmt.Sprintf("container \"%s-%s\" don't exists", e.User, e.Project)
}

type ParameterRequired struct {
	ParamName string
}

func (e *ParameterRequired) Error() string {
	return fmt.Sprintf("parameter %s is required", e.ParamName)
}

type ErrVolumeInvalid struct {
	Volume string
}

func (e *ErrVolumeInvalid) Error() string {
	return fmt.Sprintf("volume \"%s\" is invalid", e.Volume)
}
