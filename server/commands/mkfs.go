package commands

import (
	"fmt"
	"server/arguments"
	"server/stores"
	"server/structures"
	"server/utilities"
)

type Mkfs struct {
	Id   string
	Type string
}

func NewMkfs(input string) (*Mkfs, error) {
	id, err := arguments.ParseId(input)
	if err != nil {
		return nil, fmt.Errorf("error al analizar id: %w", err)
	}

	fsType, err := arguments.ParseFsType(input)
	if err != nil {
		return nil, fmt.Errorf("error al analizar type: %w", err)
	}

	return &Mkfs{
		Id:   id,
		Type: fsType,
	}, nil
}

func (m *Mkfs) Execute() error {
	mountedPartition := stores.MountedPartitions[m.Id]
	if mountedPartition == nil {
		return fmt.Errorf("no existe partición montada con ID: %s", m.Id)
	}

	superBlock := structures.NewSuperBlock(mountedPartition.Partition)

	file, err := utilities.OpenFile(mountedPartition.Path)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de la partición: %v", err)
	}
	defer file.Close()

	if err := superBlock.InitializeBitMaps(file); err != nil {
		return fmt.Errorf("error al inicializar bitmaps: %v", err)
	}

	fileSystem := structures.NewFileSystem(file, superBlock)
	if err := fileSystem.CreateUsersFile(); err != nil {
		return fmt.Errorf("error al crear archivo de usuarios: %v", err)
	}

	if err = utilities.WriteObject(file, superBlock, int64(mountedPartition.Partition.Start)); err != nil {
		return fmt.Errorf("error al escribir superblock: %v", err)
	}

	return nil
}
