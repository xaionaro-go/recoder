package xaionarogortmp

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/go-rtmp"
	rtmpmsg "github.com/xaionaro-go/go-rtmp/message"
	"github.com/xaionaro-go/recoder"
	"github.com/xaionaro-go/recoder/xaionaro-go-rtmp/yutoppgortmp"
	"github.com/xaionaro-go/xsync"
)

type Input struct {
	xsync.Mutex
	Client *rtmp.ClientConn
	Pubsub *yutoppgortmp.Pubsub
}

var _ recoder.Input = (*Input)(nil)

func NewInputFromPublisher(
	ctx context.Context,
	publisherIface recoder.Publisher,
	cfg recoder.InputConfig,
) (recoder.Input, error) {
	publisher, ok := publisherIface.(*yutoppgortmp.Pubsub)
	if !ok {
		return nil, fmt.Errorf(
			"expected a publisher or type %T, but received %T",
			publisherIface,
			publisher,
		)
	}

	return &Input{
		Pubsub: publisher,
	}, nil
}

func NewInputFromURL(
	ctx context.Context,
	urlString string,
	authKey string,
	cfg recoder.InputConfig,
) (_ *Input, _err error) {
	inClient, err := newRTMPClient(ctx, urlString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to the input endpoint '%s': %w", urlString, err)
	}
	defer func() {
		if _err != nil {
			err := inClient.Close()
			if err != nil {
				logger.Errorf(ctx, "unable to close the client for the input endpoint: %v", err)
			}
		}
	}()

	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL '%s': %w", urlString, err)
	}

	if url.Scheme == "rtmp" && url.Port() == "1935" {
		url.Host = url.Hostname()
	}
	err = inClient.Connect(ctx, &rtmpmsg.NetConnectionConnect{
		Command: rtmpmsg.NetConnectionConnectCommand{
			App:      strings.Trim(url.Path, "/"),
			Type:     "nonprivate",
			FlashVer: "StreamPanel",
			TCURL:    url.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf(
			"got an error on command 'Connect' to the input endpoint '%s': %w",
			urlString,
			err,
		)
	}

	return nil, fmt.Errorf("not implemented, yet")
}

func (input *Input) Close() error {
	var err error
	ctx := context.TODO()
	input.Do(ctx, func() {
		if input.Client == nil {
			return
		}
		err = input.Client.Close()
		input.Client = nil
	})
	return err
}
