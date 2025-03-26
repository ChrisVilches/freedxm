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
	currSessions *model.CurrentSessions
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

	log.Println("Session started")

	time.AfterFunc(time.Duration(req.TimeSeconds)*time.Second, func() {
		log.Println("Session finished")
		s.currSessions.Remove(sessionID)
	})

	return &emptypb.Empty{}, nil
}

// TODO: Implement returning "time left" as well.
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

func GRPCServerStart(port int, currSessions *model.CurrentSessions) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterServiceServer(grpcServer, &service{currSessions: currSessions})

	log.Printf("gRPC server is running on port %d...", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
