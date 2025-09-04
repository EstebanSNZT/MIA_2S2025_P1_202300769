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

type Mkusr struct {
	Username  string
	Password  string
	GroupName string
}

func NewMkusr(input string) (*Mkusr, error) {
	username, err := arguments.ParseUser(input)
	if err != nil {
		return nil, err
	}

	password, err := arguments.ParsePass(input)
	if err != nil {
		return nil, err
	}

	groupName, err := arguments.ParseGrp(input)
	if err != nil {
		return nil, err
	}

	return &Mkusr{
		Username:  username,
		Password:  password,
		GroupName: groupName,
	}, nil
}

func (m *Mkusr) Execute(session *session.Session) error {
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

	if !m.GroupExist(content) {
		return fmt.Errorf("el grupo '%s' no existe", m.GroupName)
	}

	newUid, err := m.GetNewUID(content)
	if err != nil {
		return err
	}

	newContent := fmt.Sprintf("%s%d,U,%s,%s,%s\n", content, newUid, m.GroupName, m.Username, m.Password)

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

func (m *Mkusr) GroupExist(content string) bool {
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
			return true
		}
	}

	return false
}

func (m *Mkusr) GetNewUID(content string) (int, error) {
	highestUid := 0
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		fields := strings.Split(trimmedLine, ",")

		if fields[0] == "0" || strings.TrimSpace(fields[1]) != "U" || len(fields) != 5 {
			continue
		}

		fileUsername := strings.TrimSpace(fields[3])
		if strings.EqualFold(fileUsername, m.Username) {
			return 0, fmt.Errorf("el usuario '%s' ya existe", m.Username)
		}

		uid, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err != nil {
			return 0, fmt.Errorf("error al analizar el UID: %v", err)
		}

		if uid > highestUid {
			highestUid = uid
		}
	}

	return highestUid + 1, nil
}
