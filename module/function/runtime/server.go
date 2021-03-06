package runtime

import (
	"net"

	"github.com/juju/errors"

	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Handle function handle
type Handle func(context.Context, *Message) (*Message, error)

// Server runtime server to handle message
type Server struct {
	Address string
	config  ServerConfig
	server  *grpc.Server
	handle  Handle
}

// NewServer creates a new server
func NewServer(c ServerConfig, handle Handle) (*Server, error) {
	lis, err := net.Listen("tcp", c.Address)
	if err != nil {
		return nil, errors.Trace(err)
	}
	server := grpc.NewServer(
		grpc.ConnectionTimeout(c.Timeout),
		grpc.MaxRecvMsgSize(int(c.Message.Length.Max)),
		grpc.MaxSendMsgSize(int(c.Message.Length.Max)),
	)
	s := &Server{config: c, handle: handle, server: server, Address: lis.Addr().String()}
	RegisterRuntimeServer(server, s)
	reflection.Register(server)
	go s.server.Serve(lis)
	return s, nil
}

// Handle handles messages
func (s *Server) Handle(c context.Context, m *Message) (*Message, error) {
	if s.handle == nil {
		return nil, errors.NotImplementedf("Handle")
	}
	return s.handle(c, m)
}

// Close closes server
func (s *Server) Close() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}
