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

type Rmusr struct {
	Username string
}

func NewRmusr(input string) (*Rmusr, error) {
	name, err := arguments.ParseUser(input)
	if err != nil {
		return nil, err
	}

	return &Rmusr{
		Username: name,
	}, nil
}

func (m *Rmusr) Execute(session *session.Session) error {
	if !session.IsLoggedIn {
		return fmt.Errorf("no hay sesión activa: inicie sesión primero")
	}

	if strings.EqualFold(m.Username, "root") {
		return fmt.Errorf("no se puede eliminar el usuario 'root'")
	}

	superBlock, file, sbOffset, err := stores.GetSuperBlock(session.PartitionID)
	if err != nil {
		return err
	}
	defer file.Close()

	if superBlock.Magic != 0xEF53 {
		return fmt.Errorf("la partición '%s' no tiene un sistema de archivos ext2 (magic number incorrecto)", session.PartitionID)
	}

	fileSystem := structures.NewFileSystem(file, superBlock)
	usersInode, usersInodeIndex, err := fileSystem.GetInodeByPath("/user.txt")
	if err != nil {
		return err
	}

	if usersInode == nil {
		return fmt.Errorf("el archivo de usuarios no existe")
	}

	content, err := fileSystem.ReadFileContent(usersInode)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	newContent, err := m.RemoveUser(content)
	if err != nil {
		return err
	}

	if err := fileSystem.FreeFileInode(usersInode); err != nil {
		return err
	}

	AllocatedBlocks, err := fileSystem.AllocateFileBlocks([]byte(newContent))
	if err != nil {
		return fmt.Errorf("error al asignar bloques para el nuevo contenido de /user.txt: %v", err)
	}

	usersInode.Blocks = AllocatedBlocks
	usersInode.Size = int32(len(newContent))
	usersInode.UpdateAccessTime()
	usersInode.UpdateModificationTime()

	offset := int64(superBlock.InodeStart + usersInodeIndex*superBlock.InodeSize)
	if err := utilities.WriteObject(file, *usersInode, offset); err != nil {
		return err
	}

	if err := utilities.WriteObject(file, *superBlock, sbOffset); err != nil {
		return err
	}

	return nil
}

func (m *Rmusr) RemoveUser(content string) (string, error) {
	lines := strings.Split(content, "\n")
	var newContent strings.Builder
	userFound := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		fields := strings.Split(trimmedLine, ",")
		if len(fields) < 5 || strings.TrimSpace(fields[1]) != "U" {
			newContent.WriteString(trimmedLine + "\n")
			continue
		}

		id := strings.TrimSpace(fields[0])

		if id == "0" {
			newContent.WriteString(trimmedLine + "\n")
			continue
		}

		username := strings.TrimSpace(fields[3])
		if strings.EqualFold(username, m.Username) {
			userFound = true
			fields[0] = "0"
		}

		newContent.WriteString(strings.Join(fields, ",") + "\n")
	}

	if !userFound {
		return "", fmt.Errorf("el usuario '%s' no existe o ya fue eliminado", m.Username)
	}

	return newContent.String(), nil
}
