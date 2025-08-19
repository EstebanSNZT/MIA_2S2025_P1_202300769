package structures

import "fmt"

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
	return fmt.Sprintf("Name: %s\nType: %s\nFit: %s\nStart: %d\nSize: %d\nStatus: %d\nCorrelative: %d\nID: %s\n",
		string(p.Name[:]), string(p.Type[:]), string(p.Fit[:]), p.Start, p.Size, p.Status[0], p.Correlative, string(p.ID[:]))
}
