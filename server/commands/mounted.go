package commands

import (
	"fmt"
	"server/stores"
	"strings"
)

func Mounted(input string) (string, error) {
	if input != "" {
		return "", fmt.Errorf("comando 'mounted' no requiere argumentos")
	}

	if len(stores.MountedPartitions) == 0 {
		return "No hay particiones montadas.", nil
	}

	var sb strings.Builder
	sb.WriteString("Particiones montadas:\n")

	for id := range stores.MountedPartitions {
		sb.WriteString(fmt.Sprintf(" > %s\n", id))
	}

	return sb.String(), nil
}
