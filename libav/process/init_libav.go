//go:build with_libav
// +build with_libav

package process

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/facebookincubator/go-belt/tool/logger/implementation/logrus"
	"github.com/xaionaro-go/recoder/libav/process/server"
)

const (
	EnvKeyIsEncoder = "IS_STREAMPANEL_RECODER"
	EnvKeyLogLevel  = "LOG_LEVEL"
)

func init() {
	if os.Getenv(EnvKeyIsEncoder) == "" {
		return
	}

	loggingLevelStr := os.Getenv(EnvKeyLogLevel)

	var loggerLevel logger.Level
	err := loggerLevel.Set(loggingLevelStr)
	l := logrus.Default().WithLevel(loggerLevel)
	ctx := logger.CtxWithLogger(context.Background(), l)
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()
	logger.Default = func() logger.Logger {
		return l
	}
	belt.Default = func() *belt.Belt {
		return belt.CtxBelt(ctx)
	}
	defer belt.Flush(ctx)
	if err != nil {
		logger.Errorf(ctx, "unable to parse logging level '%s': %v", loggingLevelStr, err)
	}
	logger.Debugf(ctx, "logging level: %s", loggerLevel)

	runEncoderServer(ctx, func(ctx context.Context, addr net.Addr) error {
		d := ReturnedData{
			ListenAddr: addr.String(),
		}
		b, err := json.Marshal(d)
		if err != nil {
			return fmt.Errorf("unable to serialize %#+v: %w", d, err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", b)
		return nil
	})
	belt.Flush(ctx)
	os.Exit(0)
}

func runEncoderServer(
	ctx context.Context,
	reportListener func(context.Context, net.Addr) error,
) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	if err := reportListener(ctx, listener.Addr()); err != nil {
		return fmt.Errorf("unable to report the listener: %w", err)
	}

	srv := server.NewServer()
	err = srv.Serve(ctx, listener)
	if err != nil {
		return fmt.Errorf("unable to serve: %w", err)
	}
	return nil
}
