package config

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type Command struct {
	Program       string
	Arguments     []string
	AppendPayload bool
	AppendHeaders bool
}

func (c Command) Execute(payload string, header http.Header) ([]byte, error) {
	arguments := make([]string, 0)
	copy(c.Arguments, arguments)

	if c.AppendPayload {
		arguments = append(arguments, payload)
	}

	if c.AppendHeaders {
		var header_builder strings.Builder;
		header.Write(&header_builder);

		arguments = append(arguments, header_builder.String())
	}

	return exec.Command(c.Program, arguments...).Output()
}

func (c Command) String() string {
	return fmt.Sprintf(
		"<Command cmd=%v AppendPayload=%v>",
		append([]string{c.Program}, c.Arguments...),
		c.AppendPayload,
	)
}
