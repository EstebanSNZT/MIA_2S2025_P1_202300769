package utilities

import (
	"encoding/binary"
	"fmt"
	"io"
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
	if _, err := file.Seek(position, 0); err != nil {
		return fmt.Errorf("error al buscar la posición en el archivo: %v", err)
	}
	if err := binary.Write(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}
	return nil
}

func ReadObject(file *os.File, data interface{}, position int64) error {
	if _, err := file.Seek(position, 0); err != nil {
		return fmt.Errorf("error al buscar la posición en el archivo: %v", err)
	}
	if err := binary.Read(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("error al leer del archivo: %v", err)
	}
	return nil
}

func ReadBytes(file *os.File, size int, position int64) ([]byte, error) {
	buffer := make([]byte, size)
	_, err := file.ReadAt(buffer, position)
	if err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("error al leer bytes desde la posición %d: %w", position, err)
		}
	}
	return buffer, nil
}

func WriteBytes(file *os.File, data []byte, position int64) error {
	if _, err := file.Seek(position, 0); err != nil {
		return fmt.Errorf("error al mover puntero en archivo (offset %d): %w", position, err)
	}

	bytesWritten, err := file.Write(data)
	if err != nil {
		return fmt.Errorf("error al escribir en archivo: %w", err)
	}

	if bytesWritten != len(data) {
		return fmt.Errorf("escritura incompleta: escritos %d bytes, esperados %d", bytesWritten, len(data))
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
