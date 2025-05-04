package commands

import (
	"context"
	"fmt"
	"time"

	rpc "github.com/ChrisVilches/freedxm/rpc/implementation"
	"github.com/urfave/cli/v3"
)

func ListSessions(_ context.Context, cmd *cli.Command) error {
	port := int(cmd.Int("port"))
	sessionList, err := rpc.ListSessions(port)
	if err != nil {
		return err
	}

	for _, session := range sessionList.Sessions {
		createdAt := session.CreatedAt.AsTime()
		diff := int(time.Now().Sub(createdAt).Seconds())
		format := "2006-01-02 15:04:05"

		fmt.Printf("Created at: %v\n", createdAt.In(time.Local).Format(format))
		fmt.Printf("%d/%ds\n", diff, session.TimeSeconds)
		fmt.Println(session.BlockLists)
	}

	return nil
}

func ShowConfigFileContent(_ context.Context, cmd *cli.Command) error {
	port := int(cmd.Int("port"))
	result, err := rpc.FetchConfigFileContent(port)
	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}
