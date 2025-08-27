package commands

import "server/arguments"

type Mkdir struct {
	Path string
	P    bool
}

func NewMkdir(input string) (*Mkdir, error) {
	path, err := arguments.ParsePath(input, false)
	if err != nil {
		return nil, err
	}

	p, err := arguments.ParseP(input)
	if err != nil {
		return nil, err
	}

	return &Mkdir{
		Path: path,
		P:    p,
	}, nil
}

func (m *Mkdir) Execute() error {
	return nil
}
