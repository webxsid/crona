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
	wg        sync.WaitGroup
}

func NewServer(transport, endpoint string, handler Handler, logger *runtime.Logger) *Server {
	return &Server{
		transport: transport,
		endpoint:  endpoint,
		handler:   handler,
		logger:    logger,
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
	if s.listener == nil {
		return nil
	}
	err := s.listener.Close()
	s.wg.Wait()
	if removeErr := localipc.CleanupEndpoint(s.endpoint); removeErr != nil && err == nil {
		err = removeErr
	}
	return err
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			s.logger.Error("ipc accept failed", err)
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
			if err := streamHandler.Stream(context.Background(), req, writer); err != nil {
				s.logger.Error("ipc event stream failed", err)
			}
			return
		}

		resp := s.handler.Handle(context.Background(), req)
		if err := writer.Encode(resp); err != nil {
			s.logger.Error("ipc write failed", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error("ipc read failed", err)
	}
}
