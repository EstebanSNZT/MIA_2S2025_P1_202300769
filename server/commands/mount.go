package commands

import "server/arguments"

type Mount struct {
	Path string
	Name string
}

func NewMount(input string) (*Mount, error) {
	cmdPath, err := arguments.ParsePath(input)
	if err != nil {
		return nil, err
	}

	cmdName, err := arguments.ParseName(input)
	if err != nil {
		return nil, err
	}

	return &Mount{
		Path: cmdPath,
		Name: cmdName,
	}, nil
}

func (m *Mount) Execute() error {

	return nil
}
