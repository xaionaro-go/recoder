package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/facebookincubator/go-belt/tool/logger/implementation/logrus"
	"github.com/spf13/pflag"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/recoder"
	"github.com/xaionaro-go/recoder/libav"
)

func main() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "syntax: %s <URL-from> <URL-to>\n", os.Args[0])
	}

	loggerLevel := logger.LevelWarning
	pflag.Var(&loggerLevel, "log-level", "Log level")
	netPprofAddr := pflag.String("net-pprof-listen-addr", "", "an address to listen for incoming net/pprof connections")
	pflag.Parse()
	if len(pflag.Args()) != 2 {
		pflag.Usage()
		os.Exit(1)
	}

	l := logrus.Default().WithLevel(loggerLevel)
	ctx := logger.CtxWithLogger(context.Background(), l)
	ctx, cancelFn := context.WithCancel(ctx)
	logger.Default = func() logger.Logger {
		return l
	}
	defer belt.Flush(ctx)

	if *netPprofAddr != "" {
		observability.Go(ctx, func() { l.Error(http.ListenAndServe(*netPprofAddr, nil)) })
	}

	fromURL := pflag.Arg(0)
	toURL := pflag.Arg(1)

	f, err := libav.NewFactory(ctx)
	if err != nil {
		l.Fatal(err)
	}

	l.Debugf("opening '%s' as the input...", fromURL)
	input, err := f.NewInputFromURL(ctx, fromURL, "", libav.InputConfig{})
	if err != nil {
		l.Fatal(err)
	}

	l.Debugf("opening '%s' as the output...", toURL)
	output, err := f.NewOutputFromURL(
		ctx,
		toURL, "",
		libav.OutputConfig{},
	)
	if err != nil {
		l.Fatal(err)
	}

	l.Debugf("initializing an encoder...")
	encoder, err := f.NewEncoder(ctx, recoder.EncodersConfig{})
	if err != nil {
		l.Fatal(err)
	}

	l.Debugf("initializing an recoder...")
	recoder, err := f.NewRecoder(ctx)
	if err != nil {
		l.Fatal(err)
	}

	l.Debugf("starting recoding...")
	err = recoder.Start(ctx, encoder, input, output)
	if err != nil {
		l.Fatal(err)
	}

	observability.Go(ctx, func() {
		err := recoder.Wait(ctx)
		if err != nil {
			logger.Error(ctx, "unable to wait: %v", err)
		}
		cancelFn()
	})

	l.Debugf("started recoding...")
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		defer cancelFn()
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			stats, err := recoder.GetStats(ctx)
			if err != nil {
				l.Fatal(err)
			}
			fmt.Printf("r:%d w:%d\n", stats.BytesCountRead, stats.BytesCountWrote)
		}
	}
}
