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
	if os.Getenv(EnvKeyIsEncoder) != "" {
		runEncoder()
		belt.Flush(context.TODO())
		os.Exit(0)
	}
}

func runEncoder() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Errorf("failed to listen: %w", err))
	}
	defer listener.Close()

	d := ReturnedData{
		ListenAddr: listener.Addr().String(),
	}
	b, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", b)

	loggingLevelStr := os.Getenv(EnvKeyLogLevel)

	var loggerLevel logger.Level
	err = loggerLevel.Set(loggingLevelStr)
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
		logger.Errorf(context.Background(), "unable to parse the logging level '%s': %v", loggingLevelStr, err)
	}

	logger.Debugf(ctx, "logging level: %s", loggerLevel)

	srv := server.NewServer()
	err = srv.Serve(ctx, listener)
	panic(err)
}
