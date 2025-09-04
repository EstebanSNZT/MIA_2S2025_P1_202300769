package commands

import (
	"fmt"
	"os"
	"path"
	"server/arguments"
	"server/session"
	"server/stores"
	"server/structures"
	"server/utilities"
	"strings"
)

type Mkfile struct {
	Path string
	R    bool
	Size int
	Cont string
}

func NewMkfile(input string) (*Mkfile, error) {
	path, err := arguments.ParsePath(input, false)
	if err != nil {
		return nil, err
	}

	r, err := arguments.ParseR(input)
	if err != nil {
		return nil, err
	}

	size, err := arguments.ParseSize(input, false)
	if err != nil {
		return nil, err
	}

	cont := arguments.ParseCont(input)

	return &Mkfile{
		Path: path,
		R:    r,
		Size: size,
		Cont: cont,
	}, nil
}

func (m *Mkfile) Execute(session *session.Session) error {
	if !session.IsLoggedIn {
		return fmt.Errorf("no hay sesión activa: inicie sesión primero")
	}

	cleanPath := path.Clean(m.Path)

	if cleanPath == "/" {
		return fmt.Errorf("no se puede crear un archivo en la raíz")
	}

	if cleanPath == "" {
		return fmt.Errorf("la ruta del archivo no puede estar vacía")
	}

	parentPath := path.Dir(cleanPath)
	if parentPath == "." || parentPath == "" {
		parentPath = "/"
	}

	fileName := path.Base(cleanPath)
	if fileName == "" || fileName == "." || fileName == ".." {
		return fmt.Errorf("nombre de archivo no válido")
	}

	if len(fileName) > 11 {
		return fmt.Errorf("el nombre de archivo '%s' es demasiado largo (máximo 11 caracteres)", fileName)
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

	var parentInode *structures.Inode
	var parentInodeIndex int32

	if m.R {
		var err error
		parentInode, parentInodeIndex, err = fileSystem.EnsurePathExist(parentPath, session.UserID, session.GroupID)
		if err != nil {
			return fmt.Errorf("error creando directorios padres: %w", err)
		}
	} else {
		var err error
		parentInode, parentInodeIndex, err = fileSystem.GetInodeByPath(parentPath)
		if err != nil {
			return fmt.Errorf("no se puede crear el archivo: el directorio padre '%s' no existe", parentPath)
		}
	}

	existingInodeIndex, err := fileSystem.GetInodeIndexByName(parentInode, fileName)
	if err != nil {
		return err
	}
	if existingInodeIndex != -1 {
		return fmt.Errorf("no se puede crear '%s': el archivo ya existe", fileName)
	}

	var contentBytes []byte
	if m.Cont != "" {
		fileContent, err := os.ReadFile(m.Cont)
		if err != nil {
			return fmt.Errorf("error al leer el archivo de contenido '%s': %w", m.Cont, err)
		}
		contentBytes = fileContent
	} else if m.Size > 0 {
		contentBuilder := strings.Builder{}
		contentBuilder.Grow(m.Size)
		for i := 0; i < m.Size; i++ {
			contentBuilder.WriteByte(byte('0' + (i % 10)))
		}
		contentBytes = []byte(contentBuilder.String())
	}

	fileInodeIndex, err := superBlock.GetFreeInodeIndex(file)
	if err != nil {
		return err
	}

	AllocatedBlocks, err := fileSystem.AllocateFileBlocks(contentBytes)
	if err != nil {
		return fmt.Errorf("error al asignar bloques para el contenido del archivo: %w", err)
	}

	fileInode := structures.NewInode(session.UserID, session.GroupID, int32(len(contentBytes)), [1]byte{'1'}, [3]byte{'6', '4', '4'})
	fileInode.Blocks = AllocatedBlocks
	fileInode.UpdateModificationTime()

	fileNodeOffset := int64(superBlock.InodeStart + fileInodeIndex*superBlock.InodeSize)
	if err := utilities.WriteObject(file, *fileInode, fileNodeOffset); err != nil {
		return err
	}

	if err := fileSystem.AddEntryToParent(parentInode, parentInodeIndex, fileName, fileInodeIndex); err != nil {
		return err
	}

	if err := superBlock.UpdateInodeBitmap(fileInodeIndex, [1]byte{'1'}, file); err != nil {
		return err
	}

	return nil
}
