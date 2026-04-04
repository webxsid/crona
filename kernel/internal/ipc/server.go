package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"

	"crona/kernel/internal/runtime"
	"crona/shared/localipc"
	"crona/shared/protocol"
)

type Handler interface {
	Handle(ctx context.Context, req protocol.Request) protocol.Response
}

type EventStreamHandler interface {
	Stream(ctx context.Context, req protocol.Request, writer *json.Encoder) error
}

type Server struct {
	transport string
	endpoint  string
	handler   Handler
	logger    *runtime.Logger
	listener  net.Listener
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewServer(transport, endpoint string, handler Handler, logger *runtime.Logger) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		transport: transport,
		endpoint:  endpoint,
		handler:   handler,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (s *Server) Start() error {
	ln, err := localipc.Listen(s.endpoint)
	if err != nil {
		return err
	}
	s.listener = ln
	s.wg.Add(1)
	go s.acceptLoop()
	return nil
}

func (s *Server) Close() error {
	if s.cancel != nil {
		s.cancel()
	}
	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}
	s.wg.Wait()
	if s.listener != nil {
		if removeErr := localipc.CleanupEndpoint(s.endpoint); removeErr != nil && err == nil {
			err = removeErr
		}
	}
	return err
}

func (s *Server) requestContext() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

func (s *Server) logError(msg string, err error) {
	if s.logger != nil {
		s.logger.Error(msg, err)
	}
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(s.requestContext().Err(), context.Canceled) {
				return
			}
			s.logError("ipc accept failed", err)
			continue
		}

		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer func() {
		_ = conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	writer := json.NewEncoder(conn)

	for scanner.Scan() {
		var req protocol.Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			_ = writer.Encode(protocol.Response{
				Error: &protocol.Error{
					Code:    "invalid_request",
					Message: "failed to decode request",
				},
			})
			continue
		}

		if req.Method == protocol.MethodEventsSubscribe {
			streamHandler, ok := s.handler.(EventStreamHandler)
			if !ok {
				_ = writer.Encode(protocol.Response{
					ID: req.ID,
					Error: &protocol.Error{
						Code:    "not_implemented",
						Message: "event streaming not supported",
					},
				})
				return
			}
			if err := streamHandler.Stream(s.requestContext(), req, writer); err != nil && !errors.Is(err, context.Canceled) {
				s.logError("ipc event stream failed", err)
			}
			return
		}

		resp := s.handler.Handle(s.requestContext(), req)
		if err := writer.Encode(resp); err != nil {
			s.logError("ipc write failed", err)
			return
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(s.requestContext().Err(), context.Canceled) {
		s.logError("ipc read failed", err)
	}
}
