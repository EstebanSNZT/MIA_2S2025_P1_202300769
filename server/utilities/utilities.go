package utilities

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

func CreateFile(name string) error {
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	if _, err := os.Stat(name); os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			return fmt.Errorf("error al crear el archivo: %v", err)
		}
		defer file.Close()
	}
	return nil
}

func OpenFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	return file, nil
}

func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}
	return nil
}

func ReadObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		return fmt.Errorf("error al leer del archivo: %v", err)
	}
	return nil
}

func ConvertToBytes(size int, unit string) int {
	switch unit {
	case "K":
		return size * 1024
	case "M":
		return size * 1024 * 1024
	default:
		return size
	}
}

func DeleteFile(name string) error {
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("error al eliminar el archivo: %v", err)
	}
	return nil
}
