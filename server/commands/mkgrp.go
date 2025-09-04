package commands

import (
	"fmt"
	"server/arguments"
	"server/session"
	"server/stores"
	"server/structures"
	"server/utilities"
	"strconv"
	"strings"
)

type Mkgrp struct {
	GroupName string
}

func NewMkgrp(input string) (*Mkgrp, error) {
	name, err := arguments.ParseName(input)
	if err != nil {
		return nil, err
	}

	return &Mkgrp{
		GroupName: name,
	}, nil
}

func (m *Mkgrp) Execute(session *session.Session) error {
	if !session.IsLoggedIn {
		return fmt.Errorf("no hay sesión activa: inicie sesión primero")
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

	newGid, err := m.GetNewGID(content)
	if err != nil {
		return err
	}

	newContent := fmt.Sprintf("%s%d,G,%s\n", content, newGid, m.GroupName)

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

	offset := int64(fileSystem.Sb.InodeStart + usersInodeIndex*superBlock.InodeSize)
	if err := utilities.WriteObject(file, *usersInode, offset); err != nil {
		return err
	}

	if err := utilities.WriteObject(file, *superBlock, sbOffset); err != nil {
		return err
	}

	return nil
}

func (m *Mkgrp) GetNewGID(content string) (int, error) {
	highestGid := 0
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		fields := strings.Split(trimmedLine, ",")

		if fields[0] == "0" || strings.TrimSpace(fields[1]) != "G" || len(fields) != 3 {
			continue
		}

		fileGroupName := strings.TrimSpace(fields[2])
		if strings.EqualFold(fileGroupName, m.GroupName) {
			return 0, fmt.Errorf("el grupo '%s' ya existe", m.GroupName)
		}

		gid, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err != nil {
			return 0, fmt.Errorf("error al analizar el GID: %v", err)
		}

		if gid > highestGid {
			highestGid = gid
		}
	}

	return highestGid + 1, nil
}
