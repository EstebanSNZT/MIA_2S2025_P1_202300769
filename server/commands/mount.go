package commands

import (
	"fmt"
	"server/arguments"
	"server/stores"
	"server/structures"
	"server/utilities"
	"strings"
)

type Mount struct {
	Path string
	Name string
}

func NewMount(input string) (*Mount, error) {
	allowed := []string{"path", "name"}
	if err := arguments.ValidateParams(input, allowed); err != nil {
		return nil, err
	}

	path, err := arguments.ParsePath(input, true)
	if err != nil {
		return nil, err
	}

	name, err := arguments.ParseName(input)
	if err != nil {
		return nil, err
	}

	return &Mount{
		Path: path,
		Name: name,
	}, nil
}

func (m *Mount) Execute() (string, error) {
	for _, mounted := range stores.MountedPartitions {
		if strings.Trim(string(mounted.Partition.Name[:]), "\x00 ") == m.Name {
			return "", fmt.Errorf("la partición '%s' ya está montada", m.Name)
		}
	}

	file, err := utilities.OpenFile(m.Path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var mbr structures.MBR
	if err = utilities.ReadObject(file, &mbr, 0); err != nil {
		return "", fmt.Errorf("error al leer el MBR: %w", err)
	}

	partition := mbr.GetPartitionByName(m.Name)

	if partition == nil {
		return "", fmt.Errorf("no se encontró una partición con el nombre '%s'", m.Name)
	}

	if partition.Type == [1]byte{'E'} {
		return "", fmt.Errorf("no se puede montar una partición extendida")
	}

	diskLetter, partitionCorrelative, err := stores.AllocateMountID(m.Path)
	if err != nil {
		return "", fmt.Errorf("error al asignar ID de montaje: %w", err)
	}

	partitionId := fmt.Sprintf("69%d%s", partitionCorrelative, diskLetter)
	copy(partition.ID[:], partitionId)
	partition.Status = [1]byte{'1'}
	partition.Correlative = int32(partitionCorrelative)

	mountedPartition := &stores.MountedPartition{
		Path:      m.Path,
		Partition: partition,
	}

	stores.MountedPartitions[partitionId] = mountedPartition

	partition.Status = [1]byte{'1'}

	if err := utilities.WriteObject(file, mbr, 0); err != nil {
		return "", fmt.Errorf("error al actualizar el MBR: %w", err)
	}

	var superBlock structures.SuperBlock
	if err = utilities.ReadObject(file, &superBlock, int64(partition.Start)); err != nil {
		return "", fmt.Errorf("error al leer el superbloque: %w", err)
	}

	if superBlock.Magic == 0xEF53 {
		superBlock.MntCount++
		if err = utilities.WriteObject(file, superBlock, int64(partition.Start)); err != nil {
			return "", fmt.Errorf("error al actualizar el superbloque: %w", err)
		}
	}

	return fmt.Sprintf("¡Partición montada exitosamente!\n - ID: %s\n - Ruta: %s\n - Nombre: %s", partitionId, m.Path, m.Name), nil
}
