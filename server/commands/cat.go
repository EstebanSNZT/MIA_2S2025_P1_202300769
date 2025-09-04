package commands

import (
	"fmt"
	"server/arguments"
	"server/session"
	"server/stores"
	"server/structures"
	"server/utilities"
	"strings"
)

type Cat struct {
	Files []string
}

func NewCat(input string) (*Cat, error) {
	files, err := arguments.ParseFilePaths(input)
	if err != nil {
		return nil, err
	}

	return &Cat{
		Files: files,
	}, nil
}

func (c *Cat) Execute(session *session.Session) (string, error) {
	if !session.IsLoggedIn {
		return "", fmt.Errorf("no hay sesión activa: inicie sesión primero")
	}

	superBlock, file, _, err := stores.GetSuperBlock(session.PartitionID)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if superBlock.Magic != 0xEF53 {
		return "", fmt.Errorf("la partición '%s' no tiene un sistema de archivos ext2 (magic number incorrecto)", session.PartitionID)
	}

	fileSystem := structures.NewFileSystem(file, superBlock)

	var result strings.Builder
	for _, filePath := range c.Files {
		fileInode, fileInodeIndex, err := fileSystem.GetInodeByPath(filePath)
		if err != nil {
			return "", err
		}

		if fileInode == nil {
			return "", fmt.Errorf("el archivo '%s' no existe", filePath)
		}

		content, err := fileSystem.ReadFileContent(fileInode)
		if err != nil {
			return "", fmt.Errorf("error al leer el contenido del archivo '%s': %w", filePath, err)
		}

		fileInode.UpdateAccessTime()

		offset := int64(superBlock.InodeStart + fileInodeIndex*superBlock.InodeSize)
		if err := utilities.WriteObject(file, *fileInode, offset); err != nil {
			return "", err
		}

		result.WriteString(content + "\n")
	}

	return result.String(), nil
}
