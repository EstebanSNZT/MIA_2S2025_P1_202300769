package commands

import (
	"fmt"
	"server/arguments"
	"server/stores"
	"server/structures"
	"server/utilities"
)

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
	file, err := utilities.OpenFile(m.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	var mbr structures.MBR

	if err = utilities.ReadObject(file, &mbr, 0); err != nil {
		return fmt.Errorf("error al leer el MBR: %w", err)
	}

	partition := mbr.GetPartitionByName(m.Name)

	if partition == nil {
		return fmt.Errorf("no se encontró una partición con el nombre '%s'", m.Name)
	}

	diskLetter, partitionCorrelative, err := stores.AllocateMountID(m.Path)
	if err != nil {
		return fmt.Errorf("error al asignar ID de montaje: %w", err)
	}

	partitionId := fmt.Sprintf("69%s%d", diskLetter, partitionCorrelative)
	copy(partition.ID[:], partitionId)
	partition.Correlative = int32(partitionCorrelative)

	mountedPartition := &stores.MountedPartition{
		Path:      m.Path,
		Partition: partition,
	}

	stores.MountedPartitions[partitionId] = mountedPartition

	return nil
}
