package api

import (
	"fmt"
	"net"

	"github.com/inenagl/anti-brute-force/internal/app"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	host   string
	port   int
	logger zap.Logger
	app    app.Application
	server *grpc.Server
}

func NewServer(host string, port int, logger zap.Logger, app app.Application) *Server {
	return &Server{
		host:   host,
		port:   port,
		logger: logger,
		app:    app,
	}
}

func (s *Server) Start() error {
	lsn, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingInterceptor(&s.logger)),
	)
	service := NewService(s.app, &s.logger)
	RegisterAntiBruteForceServer(server, service)
	s.server = server

	s.logger.Debug(fmt.Sprintf("starting api server on %s", lsn.Addr().String()))
	if err := server.Serve(lsn); err != nil {
		s.logger.Error(err.Error())
	}

	return nil
}

func (s *Server) Stop() error {
	s.logger.Debug("stopping api")
	s.server.GracefulStop()
	return nil
}
