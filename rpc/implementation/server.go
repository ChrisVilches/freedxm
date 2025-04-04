package rpc

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
	"google.golang.org/protobuf/types/known/wrapperspb"
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
			return nil, fmt.Errorf("'%s' not found", b)
		}

		blockLists = append(blockLists, *blockList)
	}

	sessionID := s.currSessions.Add(model.Session{
		TimeSeconds: int(req.TimeSeconds),
		BlockLists:  blockLists,
	})

	s.sessionsUpdatedCh <- struct{}{}

	log.Printf("Session started (%ds, %v)", req.TimeSeconds, req.BlockLists)

	time.AfterFunc(time.Duration(req.TimeSeconds)*time.Second, func() {
		log.Printf("Session ended (%ds, %v)", req.TimeSeconds, req.BlockLists)
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

func (*service) FetchConfigFileContent(
	_ context.Context,
	_ *emptypb.Empty,
) (*wrapperspb.StringValue, error) {
	res, err := config.ReadConfigFileRaw()
	if err != nil {
		return nil, err
	}
	return &wrapperspb.StringValue{Value: res}, nil
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
