package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"server/arguments"
	"server/stores"
	"server/structures"
	"server/utilities"
	"strings"
)

type Rep struct {
	Id         string
	Path       string
	Name       string
	PathFileLs string
}

func NewRep(input string) (*Rep, error) {
	id, err := arguments.ParseId(input)
	if err != nil {
		return nil, fmt.Errorf("error al analizar id: %w", err)
	}

	path, err := arguments.ParsePath(input, false)
	if err != nil {
		return nil, err
	}

	name, err := arguments.ParseName(input)
	if err != nil {
		return nil, err
	}

	pathFileLs, err := arguments.ParsePathFileLs(input)
	if err != nil {
		return nil, err
	}

	return &Rep{
		Id:         id,
		Path:       path,
		Name:       name,
		PathFileLs: pathFileLs,
	}, nil
}

func (r *Rep) Execute() (string, error) {
	switch r.Name {
	case "mbr":
		dotCode, err := r.generateMBRReport()
		if err != nil {
			return "", fmt.Errorf("error al generar reporte MBR: %w", err)
		}

		fmt.Println(dotCode) // Para depuración

		if err := r.generateImage(dotCode); err != nil {
			return "", fmt.Errorf("error al generar imagen: %w", err)
		}
		return "¡Reporte MBR generado exitosamente!", nil
	default:
		return "", fmt.Errorf("tipo de reporte no reconocido: %s", r.Name)
	}
}

func (r *Rep) generateMBRReport() (string, error) {
	mounted := stores.MountedPartitions[r.Id]
	if mounted == nil {
		return "", fmt.Errorf("no existe partición montada con ID: %s", r.Id)
	}

	file, err := utilities.OpenFile(mounted.Path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var mbr structures.MBR

	if err = utilities.ReadObject(file, &mbr, 0); err != nil {
		return "", fmt.Errorf("error al leer el MBR: %w", err)
	}

	dotCode, err := mbr.GenerateTable(file)
	if err != nil {
		return "", fmt.Errorf("error al generar el código DOT: %w", err)
	}

	return dotCode, nil
}

func (r *Rep) generateImage(dotCode string) error {
	format, dotPath, err := r.verifyExtension()
	if err != nil {
		return err
	}

	dir := filepath.Dir(r.Path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	if err := os.WriteFile(dotPath, []byte(dotCode), 0644); err != nil {
		return fmt.Errorf("error al escribir archivo DOT: %v", err)
	}

	cmd := exec.Command("dot", format, dotPath, "-o", r.Path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	return nil
}

func (r *Rep) verifyExtension() (string, string, error) {
	ext := strings.ToLower(filepath.Ext(r.Path))
	validExtensions := map[string]string{
		".png":  "-Tpng",
		".svg":  "-Tsvg",
		".pdf":  "-Tpdf",
		".jpg":  "-Tjpg",
		".jpeg": "-Tjpeg",
	}

	if format, ok := validExtensions[ext]; ok {
		dotPath := strings.TrimSuffix(r.Path, ext) + ".dot"
		return format, dotPath, nil
	}

	return "", "", fmt.Errorf("el archivo debe tener una extensión válida para Graphviz")
}
