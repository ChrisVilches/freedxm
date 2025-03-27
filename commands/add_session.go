package commands

import (
	"context"

	rpc "github.com/ChrisVilches/freedxm/rpc/implementation"
	"github.com/urfave/cli/v3"
)

const secondsInMinute = 60

func AddSession(_ context.Context, cmd *cli.Command) error {
	port := cmd.Int("port")
	seconds := int(cmd.Int("minutes")) * secondsInMinute
	blockListNames := cmd.StringSlice("block-lists")
	return rpc.CreateSession(int(port), seconds, blockListNames)
}
