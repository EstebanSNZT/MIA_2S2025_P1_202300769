package structures

import (
	"fmt"
	"strings"
)

type EBR struct {
	PartMount [1]byte  // Estado de la partición
	PartFit   [1]byte  // Tipo de ajuste
	PartStart int32    // Byte de inicio de la partición
	PartSize  int32    // Tamaño de la partición
	PartNext  int32    // Dirección del siguiente EBR (-1 si no hay otro)
	PartName  [16]byte // Nombre de la partición
}

func NewEBR(fit string, start int32, size int32, name string) *EBR {
	var nameArray [16]byte
	copy(nameArray[:], name)

	return &EBR{
		PartMount: [1]byte{'0'},
		PartFit:   [1]byte{fit[0]},
		PartStart: start,
		PartSize:  size,
		PartNext:  -1,
		PartName:  nameArray,
	}
}

func (e *EBR) GenerateTable() string {
	status := rune(e.PartMount[0])
	fit := rune(e.PartFit[0])
	name := strings.Trim(string(e.PartName[:]), "\x00 ")

	return fmt.Sprintf(`
	<tr><td bgcolor="lightgreen"><b>EBR</b></td><td bgcolor="lightgreen"><b>Partición Lógica</b></td></tr>
	<tr><td bgcolor="lightgray"><b>part_status</b></td><td>%c</td></tr>
	<tr><td bgcolor="lightgray"><b>part_fit</b></td><td>%c</td></tr>
	<tr><td bgcolor="lightgray"><b>part_start</b></td><td>%d</td></tr>
	<tr><td bgcolor="lightgray"><b>part_next</b></td><td>%d</td></tr>
	<tr><td bgcolor="lightgray"><b>part_size</b></td><td>%d</td></tr>
	<tr><td bgcolor="lightgray"><b>part_name</b></td><td>%s</td></tr>`,
		status, fit, e.PartStart, e.PartNext, e.PartSize, name)
}
