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

type Chgrp struct {
	Username  string
	GroupName string
}

func NewChgrp(input string) (*Chgrp, error) {
	username, err := arguments.ParseUser(input)
	if err != nil {
		return nil, err
	}

	groupName, err := arguments.ParseGrp(input)
	if err != nil {
		return nil, err
	}

	return &Chgrp{
		Username:  username,
		GroupName: groupName,
	}, nil
}

func (c *Chgrp) Execute(session *session.Session) error {
	if !session.IsLoggedIn {
		return fmt.Errorf("no hay sesión activa: inicie sesión primero")
	}

	if strings.EqualFold(c.Username, "root") {
		return fmt.Errorf("no se puede cambiar el grupo del usuario 'root'")
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
	usersInode, uiOffset, err := fileSystem.GetInodeByPath("/user.txt")
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

	if !c.GroupExist(content) {
		return fmt.Errorf("el grupo '%s' no existe o ya fue eliminado", c.GroupName)
	}

	newContent, err := c.ChangeGroup(content)
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

	if err := utilities.WriteObject(file, *usersInode, uiOffset); err != nil {
		return err
	}

	if err := utilities.WriteObject(file, *superBlock, sbOffset); err != nil {
		return err
	}

	return nil
}

func (c *Chgrp) GroupExist(content string) bool {
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
		if strings.EqualFold(fileGroupName, c.GroupName) {
			return true
		}
	}

	return false
}

func (c *Chgrp) ChangeGroup(content string) (string, error) {
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
		if strings.EqualFold(username, c.Username) {
			userFound = true
			fields[2] = c.GroupName
		}

		newContent.WriteString(strings.Join(fields, ",") + "\n")
	}

	if !userFound {
		return "", fmt.Errorf("el usuario '%s' no existe o ya fue eliminado", c.Username)
	}

	return newContent.String(), nil
}
