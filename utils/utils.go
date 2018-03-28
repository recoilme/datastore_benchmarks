package utils

import (
	"os"
	"os/exec"
	"fmt"
)

const KiB = 1 << 10
const MiB = 1 << 20

func ExportPprofOutput(profPath string, format string) (string, error) {
	return RunShellCommand("go", "tool", "pprof", "-lines",
		fmt.Sprintf("--%s", format),
		"--output", fmt.Sprintf("%s.%s", profPath, format),
		profPath)
}


// Copied from gx-go
func RunShellCommand(cmdName string, args ...string) (string, error) {

	cmdOutput, err := exec.Command(cmdName, args...).CombinedOutput()
	if err != nil {
		return string(cmdOutput), fmt.Errorf("error in cmd %s: %s", cmdName, err)
	}

	fmt.Fprintf(os.Stdout, string(cmdOutput))
	return string(cmdOutput), nil
}
