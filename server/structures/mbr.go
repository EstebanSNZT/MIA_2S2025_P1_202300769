package structures

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type MBR struct {
	Size          int32        // Tamaño del MBR en bytes
	CreationDate  [19]byte     // Fecha y hora de creación del MBR
	DiskSignature int32        // Firma del disco
	DiskFit       [1]byte      // Tipo de ajuste
	Partitions    [4]Partition // Particiones del MBR
}

func NewMBR(size int, fit string) *MBR {
	formattedTime := time.Now().Format("15:04:05 02/01/2006")

	var dateArray [19]byte
	copy(dateArray[:], formattedTime)

	return &MBR{
		Size:          int32(size),
		CreationDate:  dateArray,
		DiskSignature: rand.Int31(),
		DiskFit:       [1]byte{fit[0]},
		Partitions:    [4]Partition{},
	}
}

func (m *MBR) AddPartition(typePart string, fit string, size int, name string) error {
	offset := binary.Size(*m)

	for i := range m.Partitions {
		if m.Partitions[i].Size == 0 {
			if offset+size > int(m.Size) {
				return fmt.Errorf("no hay suficiente espacio en el disco para crear la partición '%s'", name)
			}
			m.Partitions[i].SetData(typePart, fit, offset, size, name)
			return nil
		} else {
			offset += int(m.Partitions[i].Size)
		}
	}

	return fmt.Errorf("no se encontró espacio disponible para crear la partición '%s'", name)
}

func (m *MBR) String() string {
	stringBuilder := fmt.Sprintf("--------- MBR ---------\n- Size: %d bytes\n- Creation Date: %s\n- Disk Signature: %d\n- Disk Fit: %s\n",
		m.Size, string(m.CreationDate[:]), m.DiskSignature, string(m.DiskFit[:]))

	for i, partition := range m.Partitions {
		if partition.Size > 0 {
			stringBuilder += fmt.Sprintf("--- Partition %d ---\n%s", i+1, partition.String())
		} else {
			stringBuilder += fmt.Sprintf("--- Partition %d ---\nNot allocated\n", i+1)
		}
	}

	return stringBuilder
}

func (m *MBR) GetExtendedPartition() *Partition {
	for i := range m.Partitions {
		if string(m.Partitions[i].Type[:]) == "E" {
			return &m.Partitions[i]
		}
	}
	return nil
}

func (m *MBR) GetPartitionByName(name string) *Partition {
	trimmed := strings.TrimSpace(name)
	for i := range m.Partitions {
		partName := strings.Trim(string(m.Partitions[i].Name[:]), "\x00 ")
		if partName == trimmed && m.Partitions[i].Size > 0 {
			return &m.Partitions[i]
		}
	}
	return nil
}

func (m *MBR) HasExtendedPartition() bool {
	for i := range m.Partitions {
		if string(m.Partitions[i].Type[:]) == "E" {
			return true
		}
	}
	return false
}

func (m *MBR) HasPartition(name string) bool {
	for i := range m.Partitions {
		if string(m.Partitions[i].Name[:]) == name {
			return true
		}
	}
	return false
}
