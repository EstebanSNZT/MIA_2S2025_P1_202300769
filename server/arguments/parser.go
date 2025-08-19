package arguments

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func ParsePath(input string) (string, error) {
	re := regexp.MustCompile(`path=(?:"([^"]+)"|([^ ]+))`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "", fmt.Errorf("no se encontró un path válido")
	}

	var path string
	if match[1] != "" {
		path = match[1]
	} else {
		path = match[2]
	}

	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("la ruta '%s' debe ser absoluta (empezar con /)", path)
	}

	if !strings.HasSuffix(path, ".mia") {
		return "", fmt.Errorf("el archivo '%s' debe tener extensión .mia", path)
	}

	return path, nil
}

func ParseName(input string) (string, error) {
	re := regexp.MustCompile(`name=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "", fmt.Errorf("no se encontró un nombre válido")
	}

	return match[1], nil
}

func ParseSize(input string) (int, error) {
	re := regexp.MustCompile(`size=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return 0, fmt.Errorf("no se encontró un tamaño válido")
	}

	size, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("error al convertir el tamaño: %v", err)
	}

	if size <= 0 {
		return 0, fmt.Errorf("el tamaño debe ser mayor que cero")
	}

	return size, nil
}

func ParseUnit(input string, isDisk bool) (string, error) {
	re := regexp.MustCompile(`unit=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		if isDisk {
			return "M", nil
		}
		return "K", nil
	}

	unit := strings.ToUpper(match[1])

	if isDisk {
		if unit != "M" && unit != "K" {
			return "", fmt.Errorf("unidad inválida para disco: %s (solo se permite M o K)", unit)
		}
	} else {
		if unit != "M" && unit != "K" && unit != "B" {
			return "", fmt.Errorf("unidad inválida para partición: %s (solo se permite M, K o B)", unit)
		}
	}

	return unit, nil
}

func ParseFit(input string, isDisk bool) (string, error) {
	re := regexp.MustCompile(`fit=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		if isDisk {
			return "FF", nil
		}
		return "WF", nil
	}

	fit := strings.ToUpper(match[1])

	if fit != "BF" && fit != "FF" && fit != "WF" {
		return "", fmt.Errorf("ajuste inválido: %s", fit)
	}

	return fit, nil
}

func ParseType(input string) (string, error) {
	re := regexp.MustCompile(`type=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "P", nil
	}

	typeValue := strings.ToUpper(match[1])

	if typeValue != "P" && typeValue != "E" && typeValue != "L" {
		return "", fmt.Errorf("tipo inválido: %s (solo se permite P, E o L)", typeValue)
	}

	return typeValue, nil
}
