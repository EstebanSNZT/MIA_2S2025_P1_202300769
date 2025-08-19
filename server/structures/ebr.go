package structures

type EBR struct {
	PartStatus [1]byte  // Estado de la partición
	PartFit    [1]byte  // Tipo de ajuste
	PartStart  int32    // Byte de inicio de la partición
	PartSize   int32    // Tamaño de la partición
	PartNext   int32    // Dirección del siguiente EBR (-1 si no hay otro)
	PartName   [16]byte // Nombre de la partición
}

func NewEBR(fit string, start int32, size int32, name string) *EBR {
	var nameArray [16]byte
	copy(nameArray[:], name)

	return &EBR{
		PartStatus: [1]byte{0}, // Estado inicial como no asignado
		PartFit:    [1]byte{fit[0]},
		PartStart:  start,
		PartSize:   size,
		PartName:   nameArray,
	}
}
