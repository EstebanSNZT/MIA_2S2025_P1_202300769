package commands

import (
	"server/arguments"
	"server/stores"
	"server/utilities"
)

type Rmdisk struct {
	Path string
}

func NewRmDisk(input string) (*Rmdisk, error) {
	path, err := arguments.ParsePath(input, true)
	if err != nil {
		return nil, err
	}

	return &Rmdisk{
		Path: path,
	}, nil
}

func (r *Rmdisk) Execute() error {
	for id, partition := range stores.MountedPartitions {
		if partition.Path == r.Path {
			delete(stores.MountedPartitions, id)
		}
	}

	delete(stores.MountedDisks, r.Path)

	err := utilities.DeleteFile(r.Path)
	if err != nil {
		return err
	}
	return nil
}
