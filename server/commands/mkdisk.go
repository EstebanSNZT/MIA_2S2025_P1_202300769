package commands

import (
	"fmt"
	"os"
	"server/arguments"
	"server/structures"
	"server/utilities"
)

type Mkdisk struct {
	Path string
	Size int
	Unit string
	Fit  string
}

func NewMkDisk(input string) (*Mkdisk, error) {
	cmdPath, err := arguments.ParsePath(input)
	if err != nil {
		return nil, err
	}

	cmdSize, err := arguments.ParseSize(input)
	if err != nil {
		return nil, err
	}

	cmdUnit, err := arguments.ParseUnit(input, true)
	if err != nil {
		return nil, err
	}

	cmdFit, err := arguments.ParseFit(input, true)
	if err != nil {
		return nil, err
	}

	return &Mkdisk{
		Path: cmdPath,
		Size: cmdSize,
		Unit: cmdUnit,
		Fit:  cmdFit,
	}, nil
}

func (m *Mkdisk) Execute() error {
	if err := utilities.CreateFile(m.Path); err != nil {
		return err
	}

	file, err := utilities.OpenFile(m.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	sizeBytes := utilities.ConvertToBytes(m.Size, m.Unit)

	if err = FillWithZeros(file, sizeBytes); err != nil {
		return fmt.Errorf("error llenando el disco de ceros: %w", err)
	}

	mbr := structures.NewMBR(sizeBytes, m.Fit)

	if err = utilities.WriteObject(file, *mbr, 0); err != nil {
		return fmt.Errorf("error escribiendo el MBR: %w", err)
	}

	return nil
}

func FillWithZeros(file *os.File, size int) error {
	zeroBuffer := make([]byte, 1024)

	for i := 0; i < size/1024; i++ {
		err := utilities.WriteObject(file, zeroBuffer, int64(i*1024))
		if err != nil {
			return err
		}
	}

	rest := size % 1024
	if rest > 0 {
		lastBuffer := make([]byte, rest)
		err := utilities.WriteObject(file, lastBuffer, int64(size-rest))
		if err != nil {
			return err
		}
	}

	return nil
}
