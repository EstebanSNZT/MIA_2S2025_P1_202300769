package structures

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type MBR struct {
	Size          int32        // Tamaño del MBR en bytes
	CreationDate  int64        // Fecha y hora de creación del MBR
	DiskSignature int32        // Firma del disco
	DiskFit       [1]byte      // Tipo de ajuste
	Partitions    [4]Partition // Particiones del MBR
}

func NewMBR(size int, fit string) *MBR {
	return &MBR{
		Size:          int32(size),
		CreationDate:  int64(time.Now().Unix()),
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
	creationDate := time.Unix(m.CreationDate, 0).Format("2006-01-02 15:04:05")

	stringBuilder := fmt.Sprintf("--------- MBR ---------\n- Size: %d bytes\n- Creation Date: %s\n- Disk Signature: %d\n- Disk Fit: %s\n",
		m.Size, creationDate, m.DiskSignature, string(m.DiskFit[:]))

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

func (m *MBR) GenerateTable(file *os.File) (string, error) {
	var sb strings.Builder

	sb.WriteString("digraph G {\n")
	sb.WriteString("	node [shape=record]\n")
	sb.WriteString(fmt.Sprintf(`	tabla [label=<
	<table border="0" cellborder="1" cellspacing="0">
	<tr><td colspan="2" bgcolor="gray"><b> REPORTE MBR </b></td></tr>
	<tr><td bgcolor="lightgray"><b>mbr_size</b></td><td>%d</td></tr>
	<tr><td bgcolor="lightgray"><b>mbr_creation_date</b></td><td>%s</td></tr>
	<tr><td bgcolor="lightgray"><b>mbr_disk_signature</b></td><td>%d</td></tr>`,
		m.Size, time.Unix(m.CreationDate, 0).Format("2006-01-02 15:04:05"), m.DiskSignature))

	for i := range m.Partitions {
		if m.Partitions[i].Size == 0 {
			continue
		}

		partitionTable, err := m.Partitions[i].GenerateTable(file, i)
		if err != nil {
			return "", err
		}
		sb.WriteString(partitionTable)
	}

	sb.WriteString("</table>>]\n}")

	return sb.String(), nil
}
