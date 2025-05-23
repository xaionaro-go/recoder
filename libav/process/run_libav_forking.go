//go:build with_libav && !noforking
// +build with_libav,!noforking

package process

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/facebookincubator/go-belt/tool/experimental/errmon"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/recoder/libav/process/client"
	"github.com/xaionaro-go/xpath"
)

func run(
	ctx context.Context,
) (*Recoder, error) {
	execPath, err := xpath.GetExecPath(os.Args[0])
	if err != nil {
		return nil, fmt.Errorf("unable to get self-path: %w", err)
	}
	cmd := exec.Command(execPath)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize an stdout pipe: %w", err)
	}
	cmd.Env = append(os.Environ(), EnvKeyIsEncoder+"=1", EnvKeyLogLevel+"="+logger.FromCtx(ctx).Level().String())
	err = child_process_manager.ConfigureCommand(cmd)
	errmon.ObserveErrorCtx(ctx, err)
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("unable to start a forked process of a libav-recoder: %w", err)
	}
	err = child_process_manager.AddChildProcess(cmd.Process)
	if err != nil {
		if runtime.GOOS == "windows" {
			// this is actually an error, but I have no idea how to fix it, so demoting to a debug message
			logger.Debugf(ctx, "unable to register the command to be auto-killed: %v", err)
		} else {
			logger.Errorf(ctx, "unable to register the command to be auto-killed: %v", err)
		}
	}

	decoder := json.NewDecoder(stdout)
	var d ReturnedData
	err = decoder.Decode(&d)
	logger.Debugf(ctx, "got data: %#+v", d)
	if err != nil {
		return nil, fmt.Errorf("unable to un-JSON-ize the process output: %w", err)
	}

	c := client.New(d.ListenAddr)
	level := observability.LogLevelFilter.GetLevel()
	if err := c.SetLoggingLevel(ctx, level); err != nil {
		return nil, fmt.Errorf("unable to set the logging level to %s: %w", level, err)
	}

	return &Recoder{
		Client: c,
		Cmd:    cmd,
	}, nil
}

func (r *Recoder) Kill(ctx context.Context) error {
	return r.Cmd.Process.Kill()
}

func (r *Recoder) Wait(ctx context.Context) error {
	_, err := r.Cmd.Process.Wait()
	return err
}
