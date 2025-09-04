package commands

import (
	"encoding/binary"
	"fmt"
	"os"
	"server/arguments"
	"server/structures"
	"server/utilities"
)

type Fdisk struct {
	Path string
	Name string
	Size int
	Unit string
	Type string
	Fit  string
}

func NewFdisk(input string) (*Fdisk, error) {
	path, err := arguments.ParsePath(input, true)
	if err != nil {
		return nil, err
	}

	name, err := arguments.ParseName(input)
	if err != nil {
		return nil, err
	}

	size, err := arguments.ParseSize(input, true)
	if err != nil {
		return nil, err
	}

	unit, err := arguments.ParseUnit(input, false)
	if err != nil {
		return nil, err
	}

	cmdType, err := arguments.ParseType(input)
	if err != nil {
		return nil, err
	}

	fit, err := arguments.ParseFit(input, false)
	if err != nil {
		return nil, err
	}

	return &Fdisk{
		Path: path,
		Name: name,
		Size: size,
		Unit: unit,
		Type: cmdType,
		Fit:  fit,
	}, nil
}

func (f *Fdisk) Execute() error {
	file, err := utilities.OpenFile(f.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	var mbr structures.MBR

	if err = utilities.ReadObject(file, &mbr, 0); err != nil {
		return fmt.Errorf("error al leer el MBR: %w", err)
	}

	switch f.Type {
	case "P":
		if err := f.CreatePrimaryPartition(file, &mbr); err != nil {
			return fmt.Errorf("error al crear la partición primaria: %w", err)
		}
	case "E":
		if err := f.CreateExtendedPartition(file, &mbr); err != nil {
			return fmt.Errorf("error al crear la partición extendida: %w", err)
		}
	case "L":
		if err := f.CreateLogicalPartition(file, &mbr); err != nil {
			return fmt.Errorf("error al crear la partición lógica: %w", err)
		}
	}

	return nil
}

func (f *Fdisk) CreatePrimaryPartition(file *os.File, mbr *structures.MBR) error {
	if mbr.HasPartition(f.Name) {
		return fmt.Errorf("la partición con nombre '%s' ya existe en este disco", f.Name)
	}
	return f.addPartitionToMBR(file, mbr)
}

func (f *Fdisk) CreateExtendedPartition(file *os.File, mbr *structures.MBR) error {
	if mbr.HasExtendedPartition() {
		return fmt.Errorf("ya existe una partición extendida en este disco")
	}
	if mbr.HasPartition(f.Name) {
		return fmt.Errorf("la partición con nombre '%s' ya existe en este disco", f.Name)
	}
	return f.addPartitionToMBR(file, mbr)
}

func (f *Fdisk) addPartitionToMBR(file *os.File, mbr *structures.MBR) error {
	sizeBytes := utilities.ConvertToBytes(f.Size, f.Unit)

	if err := mbr.AddPartition(f.Type, f.Fit, sizeBytes, f.Name); err != nil {
		return err
	}

	if err := utilities.WriteObject(file, *mbr, 0); err != nil {
		return fmt.Errorf("error al escribir el MBR actualizado: %w", err)
	}

	return nil
}

func (f *Fdisk) CreateLogicalPartition(file *os.File, mbr *structures.MBR) error {
	extendedPartition := mbr.GetExtendedPartition()
	if extendedPartition == nil {
		return fmt.Errorf("aún no existe una partición extendida en este disco")
	}

	lastEBR, lastEBRposition, err := findLastEBR(file, extendedPartition.Start)
	if err != nil {
		return fmt.Errorf("error al buscar el último EBR: %w", err)
	}

	var newEBRPosition int32
	var newPartStart int32

	ebrSize := int32(binary.Size(structures.EBR{}))
	if lastEBRposition == -1 {
		newEBRPosition = extendedPartition.Start
	} else {
		newEBRPosition = lastEBR.PartStart + lastEBR.PartSize
	}
	newPartStart = newEBRPosition + ebrSize

	sizeBytes := int32(utilities.ConvertToBytes(f.Size, f.Unit))
	ebr := structures.NewEBR(f.Fit, newPartStart, sizeBytes, f.Name)

	if newPartStart+sizeBytes > extendedPartition.Start+extendedPartition.Size {
		return fmt.Errorf("no hay suficiente espacio en la partición extendida para crear la partición lógica '%s'", f.Name)
	}

	if err := utilities.WriteObject(file, *ebr, int64(newEBRPosition)); err != nil {
		return fmt.Errorf("error al escribir el nuevo EBR: %w", err)
	}

	if lastEBRposition != -1 {
		lastEBR.PartNext = newEBRPosition
		if err := utilities.WriteObject(file, &lastEBR, int64(lastEBRposition)); err != nil {
			return fmt.Errorf("error al escribir el EBR anterior: %w", err)
		}
	}

	return nil
}

func findLastEBR(file *os.File, start int32) (structures.EBR, int32, error) {
	var lastEBR structures.EBR
	var currentPos int32 = start
	var lastPos int32 = -1

	for {
		if err := utilities.ReadObject(file, &lastEBR, int64(currentPos)); err != nil {
			return structures.EBR{}, -1, fmt.Errorf("error leyendo EBR en posición %d: %w", currentPos, err)
		}

		if lastPos == -1 && lastEBR.PartSize <= 0 {
			// No hay EBR previos, la partición lógica se creará en el inicio de la extendida
			break
		}

		lastPos = currentPos

		if lastEBR.PartNext < 0 {
			break
		}
		currentPos = lastEBR.PartNext
	}

	return lastEBR, lastPos, nil
}
