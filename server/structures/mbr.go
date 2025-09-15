package structures

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path"
	"server/utilities"
	"sort"
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
	sb.WriteString("	node [shape=record];\n")
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

	sb.WriteString("</table>>];\n}")

	return sb.String(), nil
}

type PartitionInfo struct {
	Partition Partition
	Index     int
}

func (m *MBR) GenerateDiskLayoutDOT(file *os.File) (string, error) {
	var sb strings.Builder
	totalSize := m.Size
	diskName := path.Base(file.Name())

	sb.WriteString("digraph G {node [shape=none]; graph [splines=false]; subgraph cluster_disk {")
	sb.WriteString(fmt.Sprintf("label=\"Disco: %s (Tamaño Total: %d bytes)\";", diskName, totalSize))
	// 1. HTML Limpio: Usamos ` (raw string literal) para evitar \"
	sb.WriteString(`style=filled; fillcolor=white; color=black; penwidth=2; table [label=<
	<table border="0" cellborder="1" cellspacing="0" cellpadding="10" width="800"><tr>`)

	mbrStructSize := int32(binary.Size(*m))
	mbrPercentage := float64(mbrStructSize) / float64(totalSize) * 100
	sb.WriteString(fmt.Sprintf(`<td bgcolor="gray" align="center"><b>MBR</b><br/>%d bytes<br/>(%.2f%%)</td>`, mbrStructSize, mbrPercentage))
	lastOffset := int64(mbrStructSize)

	validPartitions := []PartitionInfo{}
	for i, part := range m.Partitions {
		if part.Size > 0 {
			validPartitions = append(validPartitions, PartitionInfo{Partition: part, Index: i})
		}
	}
	sort.Slice(validPartitions, func(i, j int) bool {
		return validPartitions[i].Partition.Start < validPartitions[j].Partition.Start
	})

	for _, pInfo := range validPartitions {
		part := pInfo.Partition
		freeSpaceBefore := int64(part.Start) - lastOffset
		if freeSpaceBefore > 0 {
			percentage := float64(freeSpaceBefore) / float64(totalSize) * 100
			cellWidth := max(30, int(float64(freeSpaceBefore)/float64(totalSize)*800))
			sb.WriteString(fmt.Sprintf(`<td bgcolor="#ffffffff" width="%d" align="center"><b>Libre</b><br/>%d bytes<br/>(%.2f%%)</td>`, cellWidth, freeSpaceBefore, percentage))
		}

		percentage := float64(part.Size) / float64(totalSize) * 100
		cellWidth := max(50, int(float64(part.Size)/float64(totalSize)*800))
		partName := strings.TrimRight(string(part.Name[:]), "\x00")

		switch part.Type[0] {
		case 'P':
			sb.WriteString(fmt.Sprintf(`<td bgcolor="lightblue" width="%d" align="center"><b>Primaria</b><br/>%s<br/>%d bytes<br/>(%.2f%%)</td>`, cellWidth, partName, part.Size, percentage))
		case 'E':
			sb.WriteString(fmt.Sprintf(`<td bgcolor="lightcoral" width="%d" align="center" cellpadding="0">`, cellWidth))
			sb.WriteString(`<table border="0" cellborder="1" cellspacing="0" cellpadding="5" width="100%" height="100%">`)

			extendedPercentage := float64(part.Size) / float64(totalSize) * 100
			sb.WriteString(fmt.Sprintf(`<tr><td colspan="100" align="center" bgcolor="orange"><b>Extendida</b><br/>%s<br/>%d bytes<br/>(%.2f%%)</td></tr><tr>`, partName, part.Size, extendedPercentage))

			currentEbrOffset := int64(part.Start)
			lastElementEndInE := currentEbrOffset
			ebrStructSize := int64(binary.Size(EBR{}))

			for {
				var ebr EBR
				if err := utilities.ReadObject(file, &ebr, currentEbrOffset); err != nil {
					break
				}

				ebrPercentage := float64(ebrStructSize) / float64(part.Size) * 100
				sb.WriteString(fmt.Sprintf(`<td bgcolor="gray" align="center"><b>EBR</b><br/>%d bytes<br/>(%.2f%%)</td>`, ebrStructSize, ebrPercentage))
				lastElementEndInE = currentEbrOffset + ebrStructSize

				if ebr.PartSize > 0 {
					logicalPercentage := float64(ebr.PartSize) / float64(part.Size) * 100
					logicalName := strings.TrimRight(string(ebr.PartName[:]), "\x00")
					sb.WriteString(fmt.Sprintf(`<td bgcolor="lightgreen" align="center"><b>Lógica</b><br/>%s<br/>%d bytes<br/>(%.2f%%)</td>`, logicalName, ebr.PartSize, logicalPercentage))
					lastElementEndInE = int64(ebr.PartStart + ebr.PartSize)
				}

				if ebr.PartNext <= 0 {
					break
				}
				currentEbrOffset = int64(ebr.PartNext)
			}

			endOfExtended := int64(part.Start + part.Size)
			freeSpaceAtEnd := endOfExtended - lastElementEndInE
			if freeSpaceAtEnd > 0 {
				freeExtPercentage := float64(freeSpaceAtEnd) / float64(part.Size) * 100
				sb.WriteString(fmt.Sprintf(`<td bgcolor="#D3D3D3" align="center"><b>Libre Ext.</b><br/>%d bytes<br/>(%.2f%%)</td>`, freeSpaceAtEnd, freeExtPercentage))
			}
			sb.WriteString("</tr></table></td>")
		}
		lastOffset = int64(part.Start + part.Size)
	}

	finalFreeSpace := int64(totalSize) - lastOffset
	if finalFreeSpace > 0 {
		percentage := float64(finalFreeSpace) / float64(totalSize) * 100
		cellWidth := max(30, int(float64(finalFreeSpace)/float64(totalSize)*800))
		sb.WriteString(fmt.Sprintf(`<td bgcolor="#F5F5F5" width="%d" align="center"><b>Libre</b><br/>%d bytes<br/>(%.2f%%)</td>`, cellWidth, finalFreeSpace, percentage))
	}
	sb.WriteString("</tr></table>>];}}")
	return sb.String(), nil
}
