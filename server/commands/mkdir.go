package commands

import (
	"fmt"
	"path"
	"server/arguments"
	"server/session"
	"server/stores"
	"server/structures"
)

type Mkdir struct {
	Path string
	P    bool
}

func NewMkdir(input string) (*Mkdir, error) {
	path, err := arguments.ParsePath(input, false)
	if err != nil {
		return nil, err
	}

	p, err := arguments.ParseP(input)
	if err != nil {
		return nil, err
	}

	return &Mkdir{
		Path: path,
		P:    p,
	}, nil
}

func (m *Mkdir) Execute(session *session.Session) error {
	if !session.IsLoggedIn {
		return fmt.Errorf("no hay sesión activa: inicie sesión primero")
	}

	cleanPath := path.Clean(m.Path)

	if cleanPath == "" {
		return fmt.Errorf("la ruta de la carpeta no puede estar vacía")
	}

	parentPath := path.Dir(cleanPath)
	if parentPath == "." || parentPath == "" {
		parentPath = "/"
	}

	folderName := path.Base(cleanPath)
	if folderName == "" || folderName == "." || folderName == ".." {
		return fmt.Errorf("nombre de carpeta inválido")
	}

	if len(folderName) > 11 {
		return fmt.Errorf("el nombre de carpeta '%s' es demasiado largo (máximo 11 caracteres)", folderName)
	}

	superBlock, file, _, err := stores.GetSuperBlock(session.PartitionID)
	if err != nil {
		return err
	}
	defer file.Close()

	if superBlock.Magic != 0xEF53 {
		return fmt.Errorf("la partición '%s' no tiene un sistema de archivos ext2 (magic number incorrecto)", session.PartitionID)
	}

	fileSystem := structures.NewFileSystem(file, superBlock)

	if m.P {
		_, _, err := fileSystem.EnsurePathExist(cleanPath, session.UserID, session.GroupID)
		if err != nil {
			return fmt.Errorf("error al crear directorios recursivamente: %w", err)
		}
	} else {
		parentInode, parentInodeIndex, err := fileSystem.GetInodeByPath(parentPath)
		if err != nil {
			return err
		}
		parentInode.UpdateAccessTime()

		existingInodeIndex, err := fileSystem.GetInodeIndexByName(parentInode, folderName)
		if err != nil {
			return err
		}

		if existingInodeIndex != -1 {
			return fmt.Errorf("la carpeta '%s' ya existe", folderName)
		}

		newFolderInodeIndex, err := fileSystem.CreateNewFolder(parentInodeIndex, session.UserID, session.GroupID)
		if err != nil {
			return err
		}

		if err := fileSystem.AddEntryToParent(parentInode, parentInodeIndex, folderName, newFolderInodeIndex); err != nil {
			return err
		}
	}

	return nil
}
