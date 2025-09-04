package arguments

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func ParsePath(input string, isDisk bool) (string, error) {
	re := regexp.MustCompile(`-path=(?:"([^"]+)"|([^ ]+))`)
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

	if isDisk {
		if ext := filepath.Ext(path); ext != ".mia" {
			return "", fmt.Errorf("el archivo '%s' debe tener extensión .mia", path)
		}
	}

	return path, nil
}

func ParseName(input string) (string, error) {
	re := regexp.MustCompile(`-name=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "", fmt.Errorf("no se encontró un nombre válido")
	}

	return match[1], nil
}

func ParseSize(input string, isMandatory bool) (int, error) {
	re := regexp.MustCompile(`-size=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		if isMandatory {
			return 0, fmt.Errorf("no se encontró un size válido")
		}
		return 0, nil
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
	re := regexp.MustCompile(`-unit=([^ ]+)`)
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
	re := regexp.MustCompile(`-fit=([^ ]+)`)
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
	re := regexp.MustCompile(`-type=([^ ]+)`)
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

func ParseId(input string) (string, error) {
	re := regexp.MustCompile(`-id=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "", fmt.Errorf("no se encontró un id válido")
	}

	return match[1], nil
}

func ParseFsType(input string) (string, error) {
	re := regexp.MustCompile(`-type=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "full", nil
	}

	fsType := strings.ToLower(match[1])
	if fsType != "full" {
		return "", fmt.Errorf("tipo de formateo inválido: %s (solo se permite full)", fsType)
	}

	return fsType, nil
}

func ParsePathFileLs(input string) (string, error) {
	re := regexp.MustCompile(`-path_file_ls=(?:"([^"]+)"|([^ ]+))`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "", nil
	}

	var pathFileLs string
	if match[1] != "" {
		pathFileLs = match[1]
	} else {
		pathFileLs = match[2]
	}

	if !strings.HasPrefix(pathFileLs, "/") {
		return "", fmt.Errorf("la ruta '%s' debe ser absoluta (empezar con /)", pathFileLs)
	}

	return pathFileLs, nil
}

func ParseUser(input string) (string, error) {
	re := regexp.MustCompile(`-user=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "", fmt.Errorf("no se encontró un user válido")
	}

	return match[1], nil
}

func ParsePass(input string) (string, error) {
	re := regexp.MustCompile(`-pass=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "", fmt.Errorf("no se encontró un pass válido")
	}

	return match[1], nil
}

func ParseFilePaths(input string) ([]string, error) {
	re := regexp.MustCompile(`-file\d+=(?:"([^"]+)"|([^ ]+))`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("no se encontró ningún parámetro -file<N> válido")
	}

	var paths []string
	for _, match := range matches {
		var path string
		if match[1] != "" {
			path = match[1]
		} else {
			path = match[2]
		}

		if !strings.HasPrefix(path, "/") {
			return nil, fmt.Errorf("la ruta '%s' debe ser absoluta", path)
		}

		paths = append(paths, path)
	}

	return paths, nil
}

func ParseGrp(input string) (string, error) {
	re := regexp.MustCompile(`-grp=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return "", fmt.Errorf("no se encontró un grp válido")
	}

	return match[1], nil
}

func ParseCont(input string) string {
	re := regexp.MustCompile(`-cont=([^ ]+)`)
	match := re.FindStringSubmatch(input)

	if match == nil {
		return ""
	}

	return match[1]
}

func ParseR(input string) (bool, error) {
	reWithValue := regexp.MustCompile(`-r=`)
	if reWithValue.MatchString(input) {
		return false, fmt.Errorf("la bandera -r no debe llevar valor")
	}

	reFlag := regexp.MustCompile(`(^|\s)-r(\s|$)`)
	if reFlag.MatchString(input) {
		return true, nil
	}

	return false, nil
}

func ParseP(input string) (bool, error) {
	reWithValue := regexp.MustCompile(`-p=`)
	if reWithValue.MatchString(input) {
		return false, fmt.Errorf("la bandera -p no debe llevar valor")
	}

	reFlag := regexp.MustCompile(`(^|\s)-p(\s|$)`)
	if reFlag.MatchString(input) {
		return true, nil
	}

	return false, nil
}
