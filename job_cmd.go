package bus

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/webx-top/com"
)

var cmdPreParams []string
var DefaultCmdJob = &CmdJob{}

func init() {
	if com.IsWindows {
		cmdPreParams = []string{"cmd.exe", "/c"}
		//cmdPreParams = []string{"bash.exe", "-c"}
	} else {
		shell := os.Getenv("SHELL")
		if len(shell) == 0 {
			shell = "/bin/bash"
		}
		cmdPreParams = []string{shell, "-c"}
	}
}

func CmdParams(command string) []string {
	params := append([]string{}, cmdPreParams...)
	params = append(params, command)
	return params
}

type CmdJob struct {
}

type CmdParam struct {
	Command  string   `json:"command"`
	Workdir  string   `json:"workdir"`
	Environs []string `json:"environs"`
}

func (c *CmdJob) Execute(ctx context.Context, params string) (string, error) {
	p := &CmdParam{}
	if strings.HasPrefix(params, `{`) && strings.HasSuffix(params, `}`) {
		if err := json.Unmarshal([]byte(params), p); err != nil {
			return ``, err
		}
	} else {
		p.Command = params
	}
	cParams := CmdParams(p.Command)
	cmd := exec.Command(cParams[0], cParams[1:]...)
	cmd.Dir = p.Workdir
	cmd.Env = append(os.Environ(), p.Environs...)
	// cmd.Stdout = bufOut
	// cmd.Stderr = bufErr
	err := cmd.Start()
	if err != nil {
		return "failed", err
	}
	return "ok", err
}
