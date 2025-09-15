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
	allowed := []string{"name", "path", "id", "path_file_ls"}
	if err := arguments.ValidateParams(input, allowed); err != nil {
		return nil, err
	}

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

		if err := r.generateImage(dotCode); err != nil {
			return "", fmt.Errorf("error al generar imagen: %w", err)
		}
		return "¡Reporte MBR generado exitosamente!", nil
	case "disk":
		dotCode, err := r.generateDiskReport()
		if err != nil {
			return "", fmt.Errorf("error al generar reporte de disco: %w", err)
		}

		if err := r.generateImage(dotCode); err != nil {
			return "", fmt.Errorf("error al generar imagen: %w", err)
		}
		return "¡Reporte de disco generado exitosamente!", nil
	case "sb":
		dotCode, err := r.generateSBReport()
		if err != nil {
			return "", fmt.Errorf("error al generar reporte de superbloque: %w", err)
		}

		if err := r.generateImage(dotCode); err != nil {
			return "", fmt.Errorf("error al generar imagen: %w", err)
		}
		return "¡Reporte de superbloque generado exitosamente!", nil
	case "inode":
		dotCode, err := r.generateInodesReport()
		if err != nil {
			return "", fmt.Errorf("error al generar reporte de inodos: %w", err)
		}

		if err := r.generateImage(dotCode); err != nil {
			return "", fmt.Errorf("error al generar imagen: %w", err)
		}
		return "¡Reporte de inodos generado exitosamente!", nil
	case "block":
		dotCode, err := r.generateBlocksReport()
		if err != nil {
			return "", fmt.Errorf("error al generar reporte de bloques: %w", err)
		}

		if err := r.generateImage(dotCode); err != nil {
			return "", fmt.Errorf("error al generar imagen: %w", err)
		}
		return "¡Reporte de bloques generado exitosamente!", nil

	case "bm_inode", "bm_block":
		var content string
		var err error

		if r.Name == "bm_inode" {
			content, err = r.generateInodeBitmapReport()
		} else {
			content, err = r.generateBlockBitmapReport()
		}

		if err != nil {
			return "", fmt.Errorf("error al generar reporte de bitmap: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(r.Path), os.ModePerm); err != nil {
			return "", fmt.Errorf("error al crear directorios para el reporte: %w", err)
		}
		if err := os.WriteFile(r.Path, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("error al escribir el archivo de reporte: %w", err)
		}

		return fmt.Sprintf("¡Reporte %s generado exitosamente!", r.Name), nil

	case "file":
		if err := r.generateFileReport(); err != nil {
			return "", fmt.Errorf("error al generar reporte de archivo: %w", err)
		}

		return fmt.Sprintf("¡Archivo extraído exitosamente a %s!", r.Path), nil

	case "ls":
		dotCode, err := r.generateLsReport()
		if err != nil {
			return "", fmt.Errorf("error al generar reporte ls: %w", err)
		}

		if err := r.generateImage(dotCode); err != nil {
			return "", fmt.Errorf("error al generar imagen: %w", err)
		}
		return "¡Reporte ls generado exitosamente!", nil

	case "tree":
		dotCode, err := r.generateTreeReport()
		if err != nil {
			return "", fmt.Errorf("error al generar reporte tree: %w", err)
		}

		if err := r.generateImage(dotCode); err != nil {
			return "", fmt.Errorf("error al generar imagen: %w", err)
		}
		return "¡Reporte tree generado exitosamente!", nil

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

func (r *Rep) generateDiskReport() (string, error) {
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

	dotCode, err := mbr.GenerateDiskLayoutDOT(file)
	if err != nil {
		return "", fmt.Errorf("error al generar el código DOT: %w", err)
	}

	return dotCode, nil
}

func (r *Rep) generateSBReport() (string, error) {
	superBlock, file, _, err := stores.GetSuperBlock(r.Id)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return superBlock.GenerateTable(), nil
}

func (r *Rep) generateInodesReport() (string, error) {
	superBlock, file, _, err := stores.GetSuperBlock(r.Id)
	if err != nil {
		return "", err
	}
	defer file.Close()

	bitmap, err := utilities.ReadBytes(file, int(superBlock.InodesCount), int64(superBlock.BmInodeStart))
	if err != nil {
		return "", fmt.Errorf("error al leer bitmap de inodos: %v", err)
	}

	var lastInodeIndex int32 = -1

	var sb strings.Builder
	sb.WriteString("digraph G { rankdir=LR; node [shape=plaintext];")

	for i, bit := range bitmap {
		if bit != '1' {
			continue
		}

		currentInodeIndex := int32(i)
		var inode structures.Inode
		offset := int64(superBlock.InodeStart + currentInodeIndex*superBlock.InodeSize)
		if err := utilities.ReadObject(file, &inode, offset); err != nil {
			fmt.Printf("Advertencia: no se pudo leer el inodo %d, se omitirá: %v\n", i, err)
			continue
		}
		sb.WriteString(inode.GenerateTable(currentInodeIndex))

		if lastInodeIndex != -1 {
			sb.WriteString(fmt.Sprintf("inode%d -> inode%d;", lastInodeIndex, currentInodeIndex))
		}
		lastInodeIndex = currentInodeIndex
	}

	sb.WriteString("}")
	return sb.String(), nil
}

func (r *Rep) generateBlocksReport() (string, error) {
	superBlock, file, _, err := stores.GetSuperBlock(r.Id)
	if err != nil {
		return "", err
	}
	defer file.Close()

	inodeBitmap, err := utilities.ReadBytes(file, int(superBlock.InodesCount), int64(superBlock.BmInodeStart))
	if err != nil {
		return "", fmt.Errorf("error al leer bitmap de inodos: %v", err)
	}

	var sb strings.Builder
	sb.WriteString("digraph G { rankdir=LR; node [shape=plaintext];")

	processedBlocks := make(map[int32]bool) // Para no dibujar el mismo bloque dos veces
	var lastBlockIndex int32 = -1           // Para enlazar los nodos en orden de descubrimiento

	// Recorremos los inodos para obtener el contexto
	for i, bit := range inodeBitmap {
		if bit != '1' {
			continue
		}

		var inode structures.Inode
		inodeOffset := int64(superBlock.InodeStart + int32(i)*superBlock.InodeSize)
		if err := utilities.ReadObject(file, &inode, inodeOffset); err != nil {
			continue // Omitir inodos ilegibles
		}

		// Recorremos los punteros de cada inodo
		for k, blockIndex := range inode.Blocks {
			if blockIndex == -1 || processedBlocks[blockIndex] {
				continue // Omitir punteros vacíos o bloques ya procesados
			}

			// Leemos el bloque y determinamos su tipo para llamar al GenerateTable correcto
			var tableCode string
			blockOffset := int64(superBlock.BlockStart + blockIndex*superBlock.BlockSize)

			if k >= 12 { // Es un bloque de punteros (indirecto simple en adelante)
				var pointerBlock structures.PointerBlock
				if err := utilities.ReadObject(file, &pointerBlock, blockOffset); err == nil {
					tableCode = pointerBlock.GenerateTable(blockIndex)
				}
			} else { // Es un bloque de datos (archivo o carpeta)
				switch inode.Type[0] {
				case '0': // Carpeta
					var folderBlock structures.FolderBlock
					if err := utilities.ReadObject(file, &folderBlock, blockOffset); err == nil {
						tableCode = folderBlock.GenerateTable(blockIndex)
					}
				case '1': // Archivo
					var fileBlock structures.FileBlock
					if err := utilities.ReadObject(file, &fileBlock, blockOffset); err == nil {
						tableCode = fileBlock.GenerateTable(blockIndex)
					}
				}
			}

			// Si se generó la tabla, la añadimos y enlazamos
			if tableCode != "" {
				sb.WriteString(tableCode)

				if lastBlockIndex != -1 {
					sb.WriteString(fmt.Sprintf("block%d -> block%d;", lastBlockIndex, blockIndex))
				}
				lastBlockIndex = blockIndex
				processedBlocks[blockIndex] = true
			}
		}
	}

	sb.WriteString("}")
	return sb.String(), nil
}

func (r *Rep) generateInodeBitmapReport() (string, error) {
	superBlock, file, _, err := stores.GetSuperBlock(r.Id)
	if err != nil {
		return "", err
	}
	defer file.Close()

	inodeBitmap, err := utilities.ReadBytes(file, int(superBlock.InodesCount), int64(superBlock.BmInodeStart))
	if err != nil {
		return "", fmt.Errorf("error al leer bitmap de inodos: %w", err)
	}

	var sb strings.Builder
	for i, bit := range inodeBitmap {
		sb.WriteByte(bit)
		if (i+1)%20 == 0 && i < len(inodeBitmap)-1 {
			sb.WriteString("\n")
		} else if i < len(inodeBitmap)-1 {
			sb.WriteString("\t")
		}
	}
	return sb.String(), nil
}

func (r *Rep) generateBlockBitmapReport() (string, error) {
	superBlock, file, _, err := stores.GetSuperBlock(r.Id)
	if err != nil {
		return "", err
	}
	defer file.Close()

	blockBitmap, err := utilities.ReadBytes(file, int(superBlock.BlocksCount), int64(superBlock.BmBlockStart))
	if err != nil {
		return "", fmt.Errorf("error al leer bitmap de bloques: %w", err)
	}

	var sb strings.Builder
	for i, bit := range blockBitmap {
		sb.WriteByte(bit)
		if (i+1)%20 == 0 && i < len(blockBitmap)-1 {
			sb.WriteString("\n")
		} else if i < len(blockBitmap)-1 {
			sb.WriteString("\t")
		}
	}
	return sb.String(), nil
}

func (r *Rep) generateFileReport() error {
	if r.PathFileLs == "" {
		return fmt.Errorf("la ruta del archivo a extraer (-ruta) no está especificada")
	}

	superBlock, file, _, err := stores.GetSuperBlock(r.Id)
	if err != nil {
		return err
	}
	defer file.Close()

	fileSystem := structures.NewFileSystem(file, superBlock)

	fileInode, fileInodeIndex, err := fileSystem.GetInodeByPath(r.PathFileLs)
	if err != nil {
		return err
	}
	if fileInode.Type[0] != '1' {
		return fmt.Errorf("la ruta especificada no es un archivo: %s", r.PathFileLs)
	}
	content, err := fileSystem.ReadFileContent(fileInode)
	if err != nil {
		return fmt.Errorf("error al leer el contenido del archivo: %w", err)
	}

	outputDir := filepath.Dir(r.Path)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("error al crear directorio de salida '%s': %w", outputDir, err)
	}

	if err := os.WriteFile(r.Path, []byte(content), 0644); err != nil {
		return fmt.Errorf("error al escribir el archivo de reporte '%s': %w", r.Path, err)
	}

	fileInode.UpdateAccessTime()
	offset := int64(superBlock.InodeStart + fileInodeIndex*superBlock.InodeSize)
	return utilities.WriteObject(file, *fileInode, offset)
}

func (r *Rep) generateLsReport() (string, error) {
	if r.PathFileLs == "" {
		return "", fmt.Errorf("la ruta del archivo para realizar el reporte ls no está especificada")
	}

	superBlock, file, _, err := stores.GetSuperBlock(r.Id)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileSystem := structures.NewFileSystem(file, superBlock)

	dotCode, err := fileSystem.GenerateLsDOT(r.PathFileLs)
	if err != nil {
		return "", fmt.Errorf("error al generar el código DOT: %w", err)
	}

	return dotCode, nil
}

func (r *Rep) generateTreeReport() (string, error) {
	superBlock, file, _, err := stores.GetSuperBlock(r.Id)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileSystem := structures.NewFileSystem(file, superBlock)

	dotCode, err := fileSystem.GenerateTreeDOT()
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
