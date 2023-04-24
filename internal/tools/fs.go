package tools

import "os"

func EnsureDirCreated(path string) error {
	println("Asked to create dir: " + path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
