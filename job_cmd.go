package bus

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

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
			if _, err := os.Stat(shell); err != nil {
				shell = "/bin/sh"
			}
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
	Command string   `json:"command"`
	Workdir string   `json:"workdir"`
	Env     []string `json:"env"`
	Timeout string   `json:"timeout"`
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
	var duration time.Duration
	if len(p.Timeout) > 0 {
		var err error
		duration, err = time.ParseDuration(p.Timeout)
		if err != nil {
			return "failed", err
		}
	}
	var cancel context.CancelFunc
	if duration > 0 {
		ctx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, cParams[0], cParams[1:]...)
	cmd.Dir = p.Workdir
	cmd.Env = append(os.Environ(), p.Env...)
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return "failed", err
	}
	defer errReader.Close()
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return "failed", err
	}
	defer outReader.Close()
	err = cmd.Start()
	if err != nil {
		return "failed", err
	}
	reader := io.MultiReader(errReader, outReader)
	b, err := io.ReadAll(io.LimitReader(reader, 2000))
	cmd.Wait()
	return string(b), err
}
