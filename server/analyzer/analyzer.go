package analyzer

import (
	"fmt"
	"server/commands"
	"strings"
)

func Analyzer(input string) (string, error) {
	trimmed := strings.TrimSpace(input)

	spaceIndex := strings.Index(trimmed, " ")

	var command, arguments string

	if spaceIndex == -1 {
		command = strings.ToLower(trimmed)
		arguments = ""
	} else {
		command = strings.ToLower(trimmed[:spaceIndex])
		arguments = strings.TrimSpace(trimmed[spaceIndex+1:])
	}

	switch command {
	case "mkdisk":
		mkdisk, err := commands.NewMkDisk(arguments)
		if err != nil {
			return "Disco no creado.", fmt.Errorf(" mkdisk: %w", err)
		}

		if err = mkdisk.Execute(); err != nil {
			return "Disco no creado.", fmt.Errorf(" mkdisk: %w", err)
		}
		return fmt.Sprintf("Disco creado exitosamente:\n - Ruta: %s\n - Tamaño: %d %s\n - Ajuste: %s", mkdisk.Path, mkdisk.Size, mkdisk.Unit, mkdisk.Fit), nil

	case "rmdisk":
		rmdisk, err := commands.NewRmDisk(arguments)
		if err != nil {
			return "Disco no eliminado.", fmt.Errorf(" rmdisk: %w", err)
		}

		if err = rmdisk.Execute(); err != nil {
			return "Disco no eliminado.", fmt.Errorf(" rmdisk: %w", err)
		}
		return "¡Disco eliminado exitosamente!", nil

	case "fdisk":
		fdisk, err := commands.NewFdisk(arguments)
		if err != nil {
			return "Partición no creada.", fmt.Errorf(" fdisk: %w", err)
		}

		if err = fdisk.Execute(); err != nil {
			return "Partición no creada.", fmt.Errorf(" fdisk: %w", err)
		}
		return fmt.Sprintf("Partición creada exitosamente:\n - Ruta: %s\n - Nombre: %s\n - Tamaño: %d %s\n - Tipo: %s\n - Ajuste: %s", fdisk.Path, fdisk.Name, fdisk.Size, fdisk.Unit, fdisk.Type, fdisk.Fit), nil

	default:
		return "", fmt.Errorf(": comando no reconocido: %s", command)
	}
}
