package commands

import (
	"context"
	"fmt"
	"time"

	rpc "github.com/ChrisVilches/freedxm/rpc/implementation"
	"github.com/ChrisVilches/freedxm/rpc/pb"
	"github.com/ChrisVilches/freedxm/util"
	"github.com/urfave/cli/v3"
)

func printBlockLists(blockLists []*pb.BlockList) {
	for _, blockList := range blockLists {
		fmt.Printf("  %s\n", blockList.Name)
		for _, domain := range blockList.Domains {
			fmt.Printf("    %s\n", domain)
		}
		for _, proc := range blockList.Processes {
			fmt.Printf("    %s\n", proc)
		}
	}
}

func printSessionTimeInfo(session *pb.Session) {
	createdAt := session.CreatedAt.AsTime()
	diff := int32(time.Now().Sub(createdAt).Seconds())
	format := "2006-01-02 15:04:05"
	elapsed := util.SecondsToHHMMSS(diff)
	totalFmt := util.SecondsToHHMMSS(session.TimeSeconds)

	fmt.Printf("Since %v", createdAt.In(time.Local).Format(format))
	percentage := float64(diff) * 100 / float64(session.TimeSeconds)
	fmt.Printf(" (%.2f%%, %s / %s)\n", percentage, elapsed, totalFmt)
}

func ListSessions(_ context.Context, cmd *cli.Command) error {
	port := int(cmd.Int("port"))
	sessionList, err := rpc.ListSessions(port)
	if err != nil {
		return err
	}

	for idx, session := range sessionList.Sessions {
		if idx > 0 {
			fmt.Println()
		}

		printSessionTimeInfo(session)
		printBlockLists(session.BlockLists)
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
