//go:build !linux
// +build !linux

package utils

func AvailableDiskSize() (uint64, error) {
	//todo
	return 0, nil
}
