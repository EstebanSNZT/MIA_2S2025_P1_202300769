package handler

import (
	"fmt"
	"server/analyzer"
	"server/session"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Request struct {
	Script string `json:"script"`
}

type Response struct {
	Output string `json:"output"`
}

func Execute(c *fiber.Ctx, session *session.Session) error {
	req := new(Request)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Output: err.Error(),
		})
	}

	fmt.Println("Received script:", req.Script)

	script := strings.Split(req.Script, "\n")
	var output strings.Builder

	for i, command := range script {

		trimmedCommand := strings.TrimSpace(command)

		if trimmedCommand == "" || strings.HasPrefix(trimmedCommand, "#") {
			continue
		}

		output.WriteString(fmt.Sprint("Resultado línea ", i+1, " — "))

		result, err := analyzer.Analyzer(trimmedCommand, session)
		if err != nil {
			output.WriteString(fmt.Sprintf("%s Error%s\n", result, err.Error()))
		} else {
			output.WriteString(fmt.Sprintf("%s\n", result))
		}
	}

	return c.JSON(Response{
		Output: output.String(),
	})
}
