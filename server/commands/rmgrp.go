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

type Rmgrp struct {
	GroupName string
}

func NewRmgrp(input string) (*Rmgrp, error) {
	if err := arguments.ValidateParams(input, []string{"name"}); err != nil {
		return nil, err
	}

	name, err := arguments.ParseName(input)
	if err != nil {
		return nil, err
	}

	return &Rmgrp{
		GroupName: name,
	}, nil
}

func (m *Rmgrp) Execute(session *session.Session) error {
	if !session.IsLoggedIn {
		return fmt.Errorf("no hay sesión activa: inicie sesión primero")
	}

	if strings.EqualFold(m.GroupName, "root") {
		return fmt.Errorf("no se puede eliminar el grupo 'root'")
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
	usersInode, usersInodeIndex, err := fileSystem.GetInodeByPath("/users.txt")
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

	newContent, err := m.RemoveGroup(content)
	if err != nil {
		return err
	}

	if err := fileSystem.FreeFileInode(usersInode); err != nil {
		return err
	}

	AllocatedBlocks, err := fileSystem.AllocateFileBlocks([]byte(newContent))
	if err != nil {
		return fmt.Errorf("error al asignar bloques para el nuevo contenido de /users.txt: %v", err)
	}

	usersInode.Blocks = AllocatedBlocks
	usersInode.Size = int32(len(newContent))
	usersInode.UpdateAccessTime()
	usersInode.UpdateModificationTime()

	offset := int64(fileSystem.Sb.InodeStart + usersInodeIndex*superBlock.InodeSize)
	if err := utilities.WriteObject(file, *usersInode, offset); err != nil {
		return err
	}

	if err := utilities.WriteObject(file, *superBlock, sbOffset); err != nil {
		return err
	}

	return nil
}

func (m *Rmgrp) RemoveGroup(content string) (string, error) {
	lines := strings.Split(content, "\n")
	var newContent strings.Builder
	groupFound := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		fields := strings.Split(trimmedLine, ",")
		if len(fields) < 3 {
			newContent.WriteString(trimmedLine + "\n")
			continue
		}

		id := strings.TrimSpace(fields[0])

		if id == "0" {
			newContent.WriteString(trimmedLine + "\n")
			continue
		}

		groupName := strings.TrimSpace(fields[2])
		if strings.EqualFold(groupName, m.GroupName) {
			groupFound = true
			fields[0] = "0"
		}

		newContent.WriteString(strings.Join(fields, ",") + "\n")
	}

	if !groupFound {
		return "", fmt.Errorf("el grupo '%s' no existe o ya fue eliminado", m.GroupName)
	}

	return newContent.String(), nil
}
