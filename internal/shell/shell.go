package shell

import (
	"errors"
	"os/exec"
)

func Run(dir, command string, args ...string) (int, []byte, error) {
	cmd := exec.Command("sh", append([]string{"-c", command, "sh"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), out, nil
	}
	return 0, out, err
}
