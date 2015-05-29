package operatingsystem

import (
	"errors"
)

func GetOperatingSystem() (string, error) {
	// TODO: Implement OS detection
	return "", errors.New("Cannot detect OS version")
}

func IsContainerized() (bool, error) {
	// TODO: Implement jail detection
	return false, errors.New("Cannot detect if we are in container")
}
