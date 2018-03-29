package utils

import (
	"os"
	"os/exec"
	"fmt"
	"path"
	"path/filepath"
)

const KiB = 1 << 10
const MiB = 1 << 20
const OutDir = "out"
var BadgerDBPath = path.Join(OutDir, "db", "badger")
var FlatfsDBPath = path.Join(OutDir, "db", "flatfs")
var BadgerProfPath = path.Join(OutDir, "prof", "badger", "cpu.prof")
var FlatfsProfPath = path.Join(OutDir, "prof", "flatfs", "cpu.prof")
var BadgerRepoPath = path.Join(OutDir, "repo", "badger")
var FlatfsRepoPath = path.Join(OutDir, "repo", "flatfs")
// TODO: This could be handled in a vector with callbacks and configuration.

// `go tool pprof -lines --svg --output cpu.prof.svg cpu.prof`
func ExportPprofOutput(profPath string, format string) (string) {
	return RunShellCommand("go", "tool", "pprof", "-lines",
		fmt.Sprintf("--%s", format),
		"--output", fmt.Sprintf("%s.%s", profPath, format),
		profPath)
}

// Adapted from gx-go
func RunShellCommand(cmdName string, args ...string) (string) {

	cmdOutput, err := exec.Command(cmdName, args...).CombinedOutput()
	if err != nil {
		panic(fmt.Errorf("error in cmd %s: %s, output: \n%s", cmdName, err, string(cmdOutput)))
	}

	fmt.Fprintf(os.Stdout, string(cmdOutput))
	return string(cmdOutput)
}

func CreateFileAndDirs(path string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}
