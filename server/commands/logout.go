package commands

import (
	"fmt"
	"server/session"
)

func Logout(input string, session *session.Session) error {
	if input != "" {
		return fmt.Errorf("comando 'logout' no requiere argumentos")
	}

	if session.IsAuthenticated {
		session.Logout()
	} else {
		return fmt.Errorf("no hay sesión activa")
	}

	return nil
}
