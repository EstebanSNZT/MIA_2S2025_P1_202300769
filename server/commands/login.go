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

type Login struct {
	Username string
	Password string
	Id       string
}

func NewLogin(input string) (*Login, error) {
	allowed := []string{"user", "pass", "id"}
	if err := arguments.ValidateParams(input, allowed); err != nil {
		return nil, err
	}

	username, err := arguments.ParseUser(input)
	if err != nil {
		return nil, err
	}

	password, err := arguments.ParsePass(input)
	if err != nil {
		return nil, err
	}

	id, err := arguments.ParseId(input)
	if err != nil {
		return nil, err
	}

	return &Login{
		Username: username,
		Password: password,
		Id:       id,
	}, nil
}

func (l *Login) Execute(session *session.Session) error {
	if session.IsLoggedIn {
		if session.PartitionID == l.Id {
			return fmt.Errorf("ya hay una sesión activa en esta partición '%s' para el usuario '%s'", l.Id, l.Username)
		} else {
			return fmt.Errorf("ya hay una sesión activa en otra partición '%s' para el usuario '%s'", session.PartitionID, l.Username)
		}
	}

	superBlock, file, _, err := stores.GetSuperBlock(l.Id)
	if err != nil {
		return err
	}
	defer file.Close()

	if superBlock.Magic != 0xEF53 {
		return fmt.Errorf("la partición '%s' no tiene un sistema de archivos ext2 (magic number incorrecto)", l.Id)
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

	usersInode.UpdateAccessTime()

	offset := int64(superBlock.InodeStart + usersInodeIndex*superBlock.InodeSize)
	if err := utilities.WriteObject(file, *usersInode, offset); err != nil {
		return err
	}

	UID, GID, err := l.AuthenticateUser(content)
	if err != nil {
		return err
	}

	session.Login(l.Username, UID, GID, l.Id)
	fmt.Printf("sesión iniciada correctamente en la partición '%s' para el usuario '%s'\n", l.Id, l.Username)
	fmt.Printf("UID: %d, GID: %d\n", UID, GID)
	return nil
}

func (l *Login) AuthenticateUser(content string) (UID int32, GID int32, err error) {
	lines := strings.Split(content, "\n")
	groups := make(map[string]int32)
	var userUID int32 = -1
	var userGroupName string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		fields := strings.Split(trimmedLine, ",")
		idStr := strings.TrimSpace(fields[0])
		lineType := strings.TrimSpace(fields[1])

		if idStr == "0" {
			continue
		}

		if lineType == "G" && len(fields) == 3 {
			groupName := strings.ToLower(strings.TrimSpace(fields[2]))
			gid, _ := strconv.ParseInt(idStr, 10, 32)
			groups[groupName] = int32(gid)
		}

		if userUID == -1 && lineType == "U" && len(fields) == 5 {
			fileUsername := strings.TrimSpace(fields[3])
			if strings.EqualFold(fileUsername, l.Username) {
				filePassword := strings.TrimSpace(fields[4])
				if strings.EqualFold(filePassword, l.Password) {
					uid, _ := strconv.ParseInt(idStr, 10, 32)
					userUID = int32(uid)
					userGroupName = strings.ToLower(strings.TrimSpace(fields[2]))
				}
			}
		}

		if userUID != -1 {
			if _, ok := groups[userGroupName]; ok {
				break
			}
		}
	}

	if userUID == -1 {
		return -1, -1, fmt.Errorf("usuario o contraseña incorrectos")
	}

	userGID, ok := groups[userGroupName]
	if !ok {
		return -1, -1, fmt.Errorf("el grupo '%s' asignado al usuario '%s' no fue encontrado", userGroupName, l.Username)
	}

	return userUID, userGID, nil
}
