package editor

import (
	"os"
	"os/exec"
	"strings"
)

type OSEditor struct{}

func NewOSEditor() *OSEditor {
	return &OSEditor{}
}

func (e *OSEditor) Open(content string) (string, error) {
	f, err := os.CreateTemp("", "nota-*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(content); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	name := os.Getenv("EDITOR")
	if name == "" {
		name = "vi"
	}
	c := exec.Command(name, f.Name())
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return "", err
	}

	b, err := os.ReadFile(f.Name())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
