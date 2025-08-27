package commands

import (
	"fmt"
	"server/arguments"
	"server/session"
	"server/stores"
	"server/structures"
	"strings"
)

type Login struct {
	Username string
	Password string
	Id       string
}

func NewLogin(input string) (*Login, error) {
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
	if session.IsAuthenticated {
		if session.PartitionID == l.Id {
			return fmt.Errorf("ya hay una sesión activa en esta partición '%s' para el usuario '%s'", l.Id, l.Username)
		} else {
			return fmt.Errorf("ya hay una sesión activa en otra partición '%s' para el usuario '%s'", session.PartitionID, l.Username)
		}
	}

	superBlock, file, err := stores.GetSuperBlock(l.Id)
	if err != nil {
		return err
	}
	defer file.Close()

	if superBlock.Magic != 0xEF53 {
		return fmt.Errorf("la partición '%s' no tiene un sistema de archivos ext2 (magic number incorrecto)", l.Id)
	}

	fileSystem := structures.NewFileSystem(file, superBlock)
	usersInode, err := fileSystem.GetInodeByPath("/user.txt")
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

	if !l.AuthenticateUser(content) {
		return fmt.Errorf("usuario o contraseña incorrectos")
	}

	session.Login(l.Username, l.Id)

	return nil
}

func (l *Login) AuthenticateUser(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		fields := strings.Split(strings.TrimSpace(line), ",")

		if len(fields) != 5 || strings.TrimSpace(fields[1]) != "U" {
			continue
		}

		fileUsername := strings.TrimSpace(fields[3])
		filePassword := strings.TrimSpace(fields[4])

		if strings.EqualFold(fileUsername, l.Username) {
			return strings.EqualFold(filePassword, l.Password)
		}
	}
	return false
}
