package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/asticode/go-astiav"
	"github.com/davecgh/go-spew/spew"
	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/avpipeline"
	"github.com/xaionaro-go/avpipeline/codec"
	"github.com/xaionaro-go/avpipeline/kernel"
	mathcondition "github.com/xaionaro-go/avpipeline/math/condition"
	"github.com/xaionaro-go/avpipeline/packet"
	packetcondition "github.com/xaionaro-go/avpipeline/packet/condition"
	"github.com/xaionaro-go/avpipeline/processor"
	"github.com/xaionaro-go/avpipeline/types"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/recoder/libav/grpc/go/recoder_grpc"
	"github.com/xaionaro-go/recoder/libav/grpc/goconv"
	"github.com/xaionaro-go/secret"
	"github.com/xaionaro-go/xcontext"
	"github.com/xaionaro-go/xsync"
	"google.golang.org/grpc"
)

type ContextID uint64
type EncoderID uint64
type InputID uint64
type OutputID uint64

type Context struct {
	InputNode        *avpipeline.Node[*processor.FromKernel[*kernel.Input]]
	OutputNode       *avpipeline.Node[*processor.FromKernel[*kernel.Output]]
	CloseOnce        sync.Once
	RecoderCancelFn  context.CancelFunc
	RecordingEndChan chan struct{}
}

func (context *Context) Close() error {
	ok := false
	context.CloseOnce.Do(func() {
		close(context.RecordingEndChan)
		ok = true
	})
	if !ok {
		return io.ErrClosedPipe
	}
	return nil
}

type GRPCServer struct {
	recoder_grpc.UnimplementedRecoderServer

	GRPCServer *grpc.Server
	IsStarted  bool

	BeltLocker xsync.Mutex
	Belt       *belt.Belt

	ContextLocker xsync.Mutex
	Context       map[ContextID]*Context
	ContextNextID atomic.Uint64

	InputLocker xsync.Mutex
	Input       map[InputID]*kernel.Input
	InputNextID atomic.Uint64

	OutputLocker xsync.Mutex
	Output       map[OutputID]*kernel.Output
	OutputNextID atomic.Uint64
}

func NewServer() *GRPCServer {
	srv := &GRPCServer{
		GRPCServer: grpc.NewServer(),
		Context:    make(map[ContextID]*Context),
		Input:      make(map[InputID]*kernel.Input),
		Output:     make(map[OutputID]*kernel.Output),
	}
	recoder_grpc.RegisterRecoderServer(srv.GRPCServer, srv)
	return srv
}

func (srv *GRPCServer) Serve(
	ctx context.Context,
	listener net.Listener,
) error {
	if srv.IsStarted {
		panic("this GRPC server was already started at least once")
	}
	srv.IsStarted = true
	srv.Belt = belt.CtxBelt(ctx)
	logger.FromBelt(srv.Belt).Debugf("srv.GRPCServer.Serve")
	return srv.GRPCServer.Serve(listener)
}

func (srv *GRPCServer) belt() *belt.Belt {
	ctx := context.TODO()
	return xsync.DoR1(xsync.WithNoLogging(ctx, true), &srv.BeltLocker, func() *belt.Belt {
		return srv.Belt
	})
}

func (srv *GRPCServer) ctx(ctx context.Context) context.Context {
	ctx = belt.CtxWithBelt(ctx, srv.belt())
	ctx = xsync.WithNoLogging(ctx, true)
	return ctx
}

func (srv *GRPCServer) SetLoggingLevel(
	ctx context.Context,
	req *recoder_grpc.SetLoggingLevelRequest,
) (*recoder_grpc.SetLoggingLevelReply, error) {
	ctx = srv.ctx(ctx)
	logger.Debugf(ctx, "SetLoggingLevel: %#+v", req)
	srv.BeltLocker.Do(ctx, func() {
		logLevel := logLevelProtobuf2Go(req.GetLevel())
		l := logger.FromBelt(srv.Belt).WithLevel(logLevel)
		srv.Belt = srv.Belt.WithTool(logger.ToolID, l)
	})
	return &recoder_grpc.SetLoggingLevelReply{}, nil
}

func (srv *GRPCServer) NewInput(
	ctx context.Context,
	req *recoder_grpc.NewInputRequest,
) (*recoder_grpc.NewInputReply, error) {
	ctx = srv.ctx(ctx)
	logger.Debugf(ctx, "NewInput: %#+v", req)
	switch path := req.Path.GetResourcePath().(type) {
	case *recoder_grpc.ResourcePath_Url:
		return srv.newInputByURL(ctx, path, req.Config)
	default:
		return nil, fmt.Errorf("the support of path type '%T' is not implemented", path)
	}
}

func (srv *GRPCServer) newInputByURL(
	ctx context.Context,
	path *recoder_grpc.ResourcePath_Url,
	_ *recoder_grpc.InputConfig,
) (*recoder_grpc.NewInputReply, error) {
	config := kernel.InputConfig{}
	input, err := kernel.NewInputFromURL(
		xcontext.DetachDone(ctx),
		path.Url.Url,
		secret.New(path.Url.AuthKey),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to initialize an input using URL '%s' and config %#+v",
			path.Url,
			config,
		)
	}

	inputID := xsync.DoR1(ctx, &srv.InputLocker, func() InputID {
		inputID := InputID(srv.InputNextID.Add(1))
		srv.Input[inputID] = input
		return inputID
	})
	return &recoder_grpc.NewInputReply{
		Id: uint64(inputID),
	}, nil
}

func (srv *GRPCServer) CloseContext(
	ctx context.Context,
	req *recoder_grpc.CloseContextRequest,
) (*recoder_grpc.CloseContextReply, error) {
	contextID := ContextID(req.GetContextID())
	err := xsync.DoR1(ctx, &srv.ContextLocker, func() error {
		context := srv.Context[contextID]
		if context == nil {
			return fmt.Errorf("there is no open context with ID %d", contextID)
		}
		err := context.Close()
		if err != nil {
			logger.Errorf(ctx, "unable to close the context: %v", err)
		}
		delete(srv.Context, contextID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &recoder_grpc.CloseContextReply{}, nil
}

func (srv *GRPCServer) CloseInput(
	ctx context.Context,
	req *recoder_grpc.CloseInputRequest,
) (*recoder_grpc.CloseInputReply, error) {
	inputID := InputID(req.GetInputID())
	err := xsync.DoA2R1(ctx, &srv.InputLocker, srv.closeInput, ctx, inputID)
	if err != nil {
		return nil, err
	}
	return &recoder_grpc.CloseInputReply{}, nil
}

func (srv *GRPCServer) closeInput(
	ctx context.Context,
	inputID InputID,
) error {
	input := srv.Input[inputID]
	if input == nil {
		return fmt.Errorf("there is no open input with ID %d", inputID)
	}
	err := input.Close(ctx)
	if err != nil {
		logger.Errorf(ctx, "unable to close the input: %v", err)
	}
	delete(srv.Input, inputID)
	return nil
}

func (srv *GRPCServer) NewOutput(
	ctx context.Context,
	req *recoder_grpc.NewOutputRequest,
) (*recoder_grpc.NewOutputReply, error) {
	ctx = srv.ctx(ctx)
	logger.Debugf(ctx, "NewOutput: %#+v", req)
	switch path := req.Path.GetResourcePath().(type) {
	case *recoder_grpc.ResourcePath_Url:
		return srv.newOutputByURL(ctx, path, req.Config)
	default:
		return nil, fmt.Errorf("the support of path type '%T' is not implemented", path)
	}
}

func (srv *GRPCServer) GetStream(
	ctx context.Context,
	streamIndex int,
) *astiav.Stream {
	return xsync.DoR1(ctx, &srv.InputLocker, func() *astiav.Stream {
		if len(srv.Input) != 1 {
			logger.Errorf(ctx, "currently we support only copying of one input into one output")
			return nil
		}

		var input *kernel.Input
		for _, _input := range srv.Input {
			input = _input
			break
		}

		for _, stream := range input.FormatContext.Streams() {
			if stream.Index() == streamIndex {
				return stream
			}
		}

		logger.Errorf(ctx, "have not found a stream %d", streamIndex)
		return nil
	})
}

func (srv *GRPCServer) newOutputByURL(
	ctx context.Context,
	path *recoder_grpc.ResourcePath_Url,
	_ *recoder_grpc.OutputConfig,
) (*recoder_grpc.NewOutputReply, error) {
	config := kernel.OutputConfig{}
	output, err := kernel.NewOutputFromURL(
		xcontext.DetachDone(ctx),
		path.Url.Url,
		secret.New(path.Url.AuthKey),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to initialize an output using URL '%s' and config %#+v: %w",
			path.Url,
			config,
			err,
		)
	}

	outputID := xsync.DoR1(ctx, &srv.OutputLocker, func() OutputID {
		outputID := OutputID(srv.OutputNextID.Add(1))
		srv.Output[outputID] = output
		return outputID
	})
	return &recoder_grpc.NewOutputReply{
		Id: uint64(outputID),
	}, nil
}

func (srv *GRPCServer) CloseOutput(
	ctx context.Context,
	req *recoder_grpc.CloseOutputRequest,
) (*recoder_grpc.CloseOutputReply, error) {
	outputID := OutputID(req.GetOutputID())
	err := xsync.DoA2R1(ctx, &srv.InputLocker, srv.closeOutput, ctx, outputID)
	if err != nil {
		return nil, err
	}
	return &recoder_grpc.CloseOutputReply{}, nil
}

func (srv *GRPCServer) closeOutput(
	ctx context.Context,
	outputID OutputID,
) error {
	output := srv.Output[outputID]
	if output == nil {
		return fmt.Errorf("there is no open output with ID %d", outputID)
	}
	err := output.Close(ctx)
	if err != nil {
		logger.Errorf(ctx, "unable to close the output: %v", err)
	}
	delete(srv.Output, outputID)
	return nil
}

func (srv *GRPCServer) NewContext(
	ctx context.Context,
	req *recoder_grpc.NewContextRequest,
) (*recoder_grpc.NewContextReply, error) {
	ctx = srv.ctx(ctx)
	logger.Debugf(ctx, "NewContext")
	contextInstance := &Context{
		RecordingEndChan: make(chan struct{}),
	}
	contextID := xsync.DoR1(ctx, &srv.ContextLocker, func() ContextID {
		contextID := ContextID(srv.ContextNextID.Add(1))
		srv.Context[contextID] = contextInstance
		return contextID
	})
	return &recoder_grpc.NewContextReply{
		Id: uint64(contextID),
	}, nil
}

func (srv *GRPCServer) StartRecoding(
	ctx context.Context,
	req *recoder_grpc.StartRecodingRequest,
) (*recoder_grpc.StartRecodingReply, error) {
	ctx = srv.ctx(ctx)
	logger.Debugf(ctx, "StartRecoding: %s", spew.Sdump(req))

	contextID := ContextID(req.GetContextID())
	inputID := InputID(req.GetInputID())
	outputID := OutputID(req.GetOutputID())

	srv.ContextLocker.ManualLock(ctx)
	srv.InputLocker.ManualLock(ctx)
	srv.OutputLocker.ManualLock(ctx)
	defer srv.ContextLocker.ManualUnlock(ctx)
	defer srv.InputLocker.ManualUnlock(ctx)
	defer srv.OutputLocker.ManualUnlock(ctx)

	srvContext := srv.Context[contextID]
	if srvContext == nil {
		return nil, fmt.Errorf("the recorder with ID '%v' does not exist", contextID)
	}
	if srvContext.RecoderCancelFn != nil {
		return nil, fmt.Errorf("the recoder is already started")
	}

	input := srv.Input[inputID]
	if input == nil {
		return nil, fmt.Errorf("the input with ID '%v' does not exist", inputID)
	}
	output := srv.Output[outputID]
	if output == nil {
		return nil, fmt.Errorf("the output with ID '%v' does not exist", outputID)
	}

	pipelineCtx, cancelFn := context.WithCancel(xcontext.DetachDone(ctx))
	srvContext.RecoderCancelFn = cancelFn

	inputNode := avpipeline.NewNodeFromKernel(
		pipelineCtx,
		input,
		processor.DefaultOptionsInput()...,
	)
	srvContext.InputNode = inputNode

	outputNode := avpipeline.NewNodeFromKernel(
		pipelineCtx,
		output,
		processor.DefaultOptionsOutput()...,
	)
	srvContext.OutputNode = outputNode

	hasRecoder := false
	if encoderCfg := req.GetConfig(); encoderCfg != nil {
		encoderCfg, isEnabled := goconv.EncoderConfigFromThrift(encoderCfg)
		if isEnabled {
			logger.Debugf(ctx, "encoding enabled")
			streamsMerger, err := newFrameStreamsMerger(encoderCfg)
			if err != nil {
				cancelFn()
				return nil, fmt.Errorf("unable to initialize frame streams merger: %w", err)
			}
			if len(encoderCfg.OutputVideoTracks) > 1 {
				cancelFn()
				return nil, fmt.Errorf("we currently support recoding to a single video track at most only, but requested %d video tracks", len(encoderCfg.OutputVideoTracks))
			}
			var vCodec string
			if len(encoderCfg.OutputVideoTracks) == 1 {
				videoTrack := encoderCfg.OutputVideoTracks[0]
				vCodec = optimalVideoCodec(videoTrack.Config.Codec)
			}
			var aCodec string
			if len(encoderCfg.OutputAudioTracks) > 1 {
				cancelFn()
				return nil, fmt.Errorf("we currently support recoding to a single audio track at most only, but requested %d audio tracks", len(encoderCfg.OutputAudioTracks))
			}
			if len(encoderCfg.OutputAudioTracks) == 1 {
				audioTrack := encoderCfg.OutputAudioTracks[0]
				aCodec = audioTrack.Config.Codec.String()
			}
			hasRecoder = true
			decoderNode := avpipeline.NewNodeFromKernel(
				pipelineCtx,
				kernel.NewDecoder(
					pipelineCtx,
					codec.NewNaiveDecoderFactory(
						pipelineCtx,
						0, "",
						nil, nil,
					),
				),
				processor.DefaultOptionsRecoder()...,
			)
			inputNode.AddPushPacketsTo(
				decoderNode,
				packetcondition.Function(func(ctx context.Context, pkt packet.Input) bool {
					outStreamID, err := streamsMerger.StreamIndexAssign(ctx, types.InputPacketOrFrameUnion{
						Packet: &pkt,
					})
					if err != nil {
						logger.Errorf(ctx, "unable to get the output stream ID: %w", err)
						return false
					}
					return outStreamID.IsSet()
				}),
			)
			videoOpts := astiav.NewDictionary()
			setFinalizerFree(ctx, videoOpts)
			videoOpts.Set("g", "10", 0)
			encoderNode := avpipeline.NewNodeFromKernel(
				pipelineCtx,
				kernel.NewEncoder(
					pipelineCtx,
					codec.NewNaiveEncoderFactory(
						pipelineCtx,
						vCodec, aCodec,
						0, "",
						nil, nil,
					),
					nil,
				),
				processor.DefaultOptionsRecoder()...,
			)
			mergerNode := avpipeline.NewNodeFromKernel(
				pipelineCtx,
				kernel.NewMapStreamIndices(
					pipelineCtx,
					streamsMerger,
				),
				processor.DefaultOptionsRecoder()...,
			)
			waiterNode := avpipeline.NewNodeFromKernel(
				pipelineCtx,
				kernel.NewWait(
					packetcondition.Not{packetcondition.SeenStreamCount(mathcondition.GreaterOrEqual(uint(
						len(encoderCfg.OutputVideoTracks) + len(encoderCfg.OutputAudioTracks),
					)))},
					nil,
					1000,
					1000,
				),
				processor.DefaultOptionsRecoder()...,
			)
			decoderNode.AddPushFramesTo(mergerNode)
			mergerNode.AddPushFramesTo(encoderNode)
			encoderNode.AddPushPacketsTo(waiterNode)
			waiterNode.AddPushPacketsTo(outputNode)
		}
	}
	if !hasRecoder {
		inputNode.AddPushPacketsTo(outputNode)
	}

	observability.Go(ctx, func(ctx context.Context) {
		defer logger.Debugf(ctx, "recoding ended")
		defer func() {
			srvContext.Close()
		}()
		defer cancelFn()
		errCh := make(chan avpipeline.ErrNode, 10)

		observability.Go(ctx, func(ctx context.Context) {
			defer logger.Debugf(ctx, "/errCh listener")
			for {
				select {
				case <-pipelineCtx.Done():
					return
				case err := <-errCh:
					logger.Errorf(ctx, "received error: %v", err)
					cancelFn()
					return
				}
			}
		})
		logger.Debugf(ctx, "pipeline.Serve: %s", inputNode.String())
		var pushTos []avpipeline.AbstractNode
		for _, pushTo := range inputNode.PushPacketsTo {
			pushTos = append(pushTos, pushTo.Node)
		}
		if err := avpipeline.NotifyAboutPacketSourcesRecursively(ctx, input, pushTos...); err != nil {
			logger.Errorf(ctx, "unable to notify about the stream formats: %v", err)
		}
		avpipeline.ServeRecursively(pipelineCtx, avpipeline.ServeConfig{}, errCh, inputNode)
	})

	return &recoder_grpc.StartRecodingReply{}, nil
}

func (srv *GRPCServer) RecodingEndedChan(
	req *recoder_grpc.RecodingEndedChanRequest,
	streamSrv recoder_grpc.Recoder_RecodingEndedChanServer,
) (_ret error) {
	ctx := srv.ctx(streamSrv.Context())
	contextID := ContextID(req.GetContextID())

	logger.Tracef(ctx, "RecodingEndedChan(%v)", contextID)
	defer func() { logger.Tracef(ctx, "/RecodingEndedChan(%v): %v", contextID, _ret) }()

	context := xsync.DoR1(ctx, &srv.ContextLocker, func() *Context {
		return srv.Context[contextID]
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-context.RecordingEndChan:
	}

	return streamSrv.Send(&recoder_grpc.RecodingEndedChanReply{})
}

func (srv *GRPCServer) GetStats(
	ctx context.Context,
	req *recoder_grpc.GetRecoderStatsRequest,
) (*recoder_grpc.GetRecoderStatsReply, error) {
	return xsync.DoR2(ctx, &srv.ContextLocker, func() (*recoder_grpc.GetRecoderStatsReply, error) {
		context := srv.Context[ContextID(req.GetContextID())]
		readStats := context.InputNode.GetStatistics().GetStats()
		writeStats := context.OutputNode.GetStatistics().GetStats()
		return &recoder_grpc.GetRecoderStatsReply{
			BytesCountRead:  readStats.BytesCountWrote,
			BytesCountWrote: writeStats.BytesCountRead,
		}, nil
	})
}

func (srv *GRPCServer) Die(
	ctx context.Context,
	req *recoder_grpc.DieRequest,
) (*recoder_grpc.DieReply, error) {
	srv.ContextLocker.ManualLock(ctx)
	srv.InputLocker.ManualLock(ctx)
	srv.OutputLocker.ManualLock(ctx)
	defer srv.ContextLocker.ManualUnlock(ctx)
	defer srv.InputLocker.ManualUnlock(ctx)
	defer srv.OutputLocker.ManualUnlock(ctx)

	for key, ctx := range srv.Context {
		ctx.RecoderCancelFn()
		ctx.RecoderCancelFn = nil
		delete(srv.Context, key)
	}

	var inputIDs []InputID
	for inputID := range srv.Input {
		inputIDs = append(inputIDs, inputID)
	}

	var outputIDs []InputID
	for outputID := range srv.Input {
		outputIDs = append(outputIDs, outputID)
	}

	for _, inputID := range inputIDs {
		srv.closeInput(ctx, inputID)
	}

	for _, outputID := range outputIDs {
		srv.closeInput(ctx, outputID)
	}

	srv.GRPCServer.Stop()
	return &recoder_grpc.DieReply{}, nil
}
