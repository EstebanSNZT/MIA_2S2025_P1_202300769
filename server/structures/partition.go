package structures

import (
	"fmt"
	"os"
	"server/utilities"
	"strings"
)

type Partition struct {
	Status      [1]byte  // Estado de la partición
	Type        [1]byte  // Tipo de partición
	Fit         [1]byte  // Ajuste de la partición
	Start       int32    // Byte de inicio de la partición
	Size        int32    // Tamaño de la partición
	Name        [16]byte // Nombre de la partición
	Correlative int32    // Correlativo de la partición
	ID          [4]byte  // ID de la partición
}

func (p *Partition) SetData(typePart string, fit string, start int, size int, name string) {
	p.Status = [1]byte{0}
	p.Type = [1]byte{typePart[0]}
	p.Fit = [1]byte{fit[0]}
	p.Start = int32(start)
	p.Size = int32(size)
	copy(p.Name[:], name)
	p.Correlative = -1
}

func (p *Partition) String() string {
	return fmt.Sprintf("- Name: %s\n- Type: %s\n- Fit: %s\n- Start: %d\n- Size: %d\n- Status: %d\n- Correlative: %d\n- ID: %s\n",
		string(p.Name[:]), string(p.Type[:]), string(p.Fit[:]), p.Start, p.Size, p.Status[0], p.Correlative, string(p.ID[:]))
}

func (p *Partition) GenerateTable(file *os.File, i int) (string, error) {
	status := rune(p.Status[0])
	pType := rune(p.Type[0])
	fit := rune(p.Fit[0])
	name := strings.Trim(string(p.Name[:]), "\x00 ")

	var color string
	switch pType {
	case 'P':
		color = "lightblue"
	case 'E':
		color = "orange"
	default:
		color = "white"
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`
	<tr><td colspan="2" bgcolor="%s"><b> PARTICIÓN %d </b></td></tr>
	<tr><td bgcolor="lightgray"><b>part_status</b></td><td>%c</td></tr>
	<tr><td bgcolor="lightgray"><b>part_type</b></td><td>%c</td></tr>
	<tr><td bgcolor="lightgray"><b>part_fit</b></td><td>%c</td></tr>
	<tr><td bgcolor="lightgray"><b>part_start</b></td><td>%d</td></tr>
	<tr><td bgcolor="lightgray"><b>part_size</b></td><td>%d</td></tr>
	<tr><td bgcolor="lightgray"><b>part_name</b></td><td>%s</td></tr>
	`, color, i+1, status, pType, fit, p.Start, p.Size, name))

	if pType == 'E' {
		var ebr EBR
		offset := p.Start
		for {
			if err := utilities.ReadObject(file, &ebr, int64(offset)); err != nil {
				return "", fmt.Errorf("error leyendo EBR: %v", err)
			}

			sb.WriteString(ebr.GenerateTable())

			if ebr.PartNext <= 0 {
				break
			}

			offset = ebr.PartNext
		}
	}

	return sb.String(), nil
}
