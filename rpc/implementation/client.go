package rpc

import (
	"context"
	"fmt"

	"github.com/ChrisVilches/freedxm/rpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func getClient(port int) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func ListSessions(port int) (*pb.SessionList, error) {
	conn, err := getClient(port)

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewServiceClient(conn)

	return client.FetchSessions(context.Background(), &emptypb.Empty{})
}

func CreateSession(port, timeSeconds int, blockListNames []string) error {
	conn, err := getClient(port)

	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewServiceClient(conn)

	_, err = client.CreateSession(context.Background(), &pb.NewSessionRequest{
		BlockLists:  blockListNames,
		TimeSeconds: int32(timeSeconds),
	})

	return err
}

func FetchConfigFileContent(port int) (string, error) {
	conn, err := getClient(port)

	if err != nil {
		return "", err
	}

	defer conn.Close()

	client := pb.NewServiceClient(conn)

	res, err := client.FetchConfigFileContent(
		context.Background(),
		&emptypb.Empty{},
	)

	return res.Value, err
}
