package config

import (
	"fmt"
	"os/exec"
)

type Command struct {
	Program       string
	Arguments     []string
	AppendPayload bool
}

func (c Command) Execute(payload string) ([]byte, error) {
	arguments := make([]string, 0)
	copy(c.Arguments, arguments)
	if c.AppendPayload {
		arguments = append(arguments, payload)
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
