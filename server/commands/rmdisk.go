package commands

import (
	"server/arguments"
	"server/utilities"
)

type Rmdisk struct {
	Path string
}

func NewRmDisk(input string) (*Rmdisk, error) {
	cmdPath, err := arguments.ParsePath(input)
	if err != nil {
		return nil, err
	}

	return &Rmdisk{
		Path: cmdPath,
	}, nil
}

func (r *Rmdisk) Execute() error {
	err := utilities.DeleteFile(r.Path)
	if err != nil {
		return err
	}
	return nil
}
