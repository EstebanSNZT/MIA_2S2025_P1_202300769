package analyzer

import (
	"fmt"
	"server/commands"
	"strings"
)

func Analyzer(input string) (string, error) {
	spaceIndex := strings.Index(input, " ")

	var command, arguments string

	if spaceIndex == -1 {
		command = strings.ToLower(input)
		arguments = ""
	} else {
		command = strings.ToLower(input[:spaceIndex])
		arguments = strings.TrimSpace(input[spaceIndex+1:])
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

	case "mount":
		mount, err := commands.NewMount(arguments)
		if err != nil {
			return "Disco no montado.", fmt.Errorf(" mount: %w", err)
		}

		result, err := mount.Execute()
		if err != nil {
			return "Disco no montado.", fmt.Errorf(" mount: %w", err)
		}
		return result, nil

	case "mounted":
		result, err := commands.Mounted(arguments)
		if err != nil {
			return "No se pudieron listar las particiones montadas.", fmt.Errorf(" mounted: %w", err)
		}
		return result, nil

	case "mkfs":
		mkfs, err := commands.NewMkfs(arguments)
		if err != nil {
			return "Sistema de archivos no formateado.", fmt.Errorf(" mkfs: %w", err)
		}

		if err = mkfs.Execute(); err != nil {
			return "Sistema de archivos no formateado.", fmt.Errorf(" mkfs: %w", err)
		}
		return "¡Sistema de archivos formateado exitosamente!", nil

	case "rep":
		rep, err := commands.NewRep(arguments)
		if err != nil {
			return "Reporte no creado.", fmt.Errorf(" rep: %w", err)
		}

		result, err := rep.Execute()
		if err != nil {
			return "Reporte no creado.", fmt.Errorf(" rep: %w", err)
		}
		return result, nil

	default:
		return "", fmt.Errorf(": comando no reconocido: %s", command)
	}
}
