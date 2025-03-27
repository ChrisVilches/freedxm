package rpc

// TODO: I think reading this is important.
// For semantics around ctx use and closing/ending streaming RPCs, please refer
// to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
// TODO: Clean every file in this directory. Messy as hell.
import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ChrisVilches/freedxm/config"
	"github.com/ChrisVilches/freedxm/model"
	"github.com/ChrisVilches/freedxm/rpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type service struct {
	pb.UnimplementedServiceServer
	currSessions      *model.CurrentSessions
	sessionsUpdatedCh chan<- struct{}
}

func (s *service) CreateSession(
	_ context.Context,
	req *pb.NewSessionRequest,
) (*emptypb.Empty, error) {
	blockLists := make([]model.BlockList, 0)

	for _, b := range req.BlockLists {
		blockList, err := config.GetBlockListByName(b)

		if err != nil {
			return nil, err
		}

		if blockList == nil {
			return nil, fmt.Errorf("%s not found", b)
		}

		blockLists = append(blockLists, *blockList)
	}

	sessionID := s.currSessions.Add(model.Session{
		TimeSeconds: int(req.TimeSeconds),
		BlockLists:  blockLists,
	})

	s.sessionsUpdatedCh <- struct{}{}

	log.Println("Session started")

	time.AfterFunc(time.Duration(req.TimeSeconds)*time.Second, func() {
		log.Println("Session finished")
		s.currSessions.Remove(sessionID)
		s.sessionsUpdatedCh <- struct{}{}
	})

	return &emptypb.Empty{}, nil
}

func (s *service) FetchSessions(
	_ context.Context,
	_ *emptypb.Empty,
) (*pb.SessionList, error) {
	result := make([]*pb.Session, 0)

	for _, s := range s.currSessions.GetAll() {
		newSess := pb.Session{TimeSeconds: int32(s.TimeSeconds)}

		for _, b := range s.BlockLists {
			newSess.BlockLists = append(newSess.BlockLists, &pb.BlockList{
				Name:      b.Name,
				Domains:   b.Domains,
				Processes: b.Processes,
			})
		}

		result = append(result, &newSess)
	}

	return &pb.SessionList{Sessions: result}, nil
}

func GRPCServerStart(
	ctx context.Context,
	port int,
	currSessions *model.CurrentSessions,
	sessionsUpdatedCh chan<- struct{},
) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	serv := &service{
		currSessions:      currSessions,
		sessionsUpdatedCh: sessionsUpdatedCh,
	}

	pb.RegisterServiceServer(grpcServer, serv)

	log.Printf("gRPC server is running on port %d...", port)

	serverErr := make(chan error)

	go func() { serverErr <- grpcServer.Serve(listener) }()

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		return ctx.Err()
	case err := <-serverErr:
		return err
	}
}
