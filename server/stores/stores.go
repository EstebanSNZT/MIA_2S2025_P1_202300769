package stores

import (
	"fmt"
	"os"
	"server/structures"
	"server/utilities"
)

type MountedPartition struct {
	Path      string
	Partition *structures.Partition
}

type MountedDisk struct {
	Letter         string
	PartitionCount int
}

var MountedPartitions = make(map[string]*MountedPartition)
var MountedDisks = make(map[string]*MountedDisk)
var alphabet = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}
var nextLetterIndex = 0

func AllocateMountID(path string) (string, int, error) {
	disk, exists := MountedDisks[path]

	if !exists {
		if nextLetterIndex >= len(alphabet) {
			return "", 0, fmt.Errorf("no hay más letras disponibles para montar discos")
		}

		disk = &MountedDisk{
			Letter:         alphabet[nextLetterIndex],
			PartitionCount: 0,
		}
		MountedDisks[path] = disk
		nextLetterIndex++
	}

	disk.PartitionCount++
	return disk.Letter, disk.PartitionCount, nil
}

func GetSuperBlock(id string) (*structures.SuperBlock, *os.File, int64, error) {
	mountedPartition := MountedPartitions[id]
	if mountedPartition == nil {
		return nil, nil, 0, fmt.Errorf("no existe partición montada con ID: %s", id)
	}

	file, err := utilities.OpenFile(mountedPartition.Path)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error al abrir el archivo de la partición: %v", err)
	}

	offset := int64(mountedPartition.Partition.Start)

	var superBlock structures.SuperBlock
	if err := utilities.ReadObject(file, &superBlock, offset); err != nil {
		return nil, nil, 0, fmt.Errorf("error al leer el superblock: %v", err)
	}

	return &superBlock, file, offset, nil
}
