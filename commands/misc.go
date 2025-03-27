package commands

import (
	"context"
	"fmt"
	"github.com/ChrisVilches/freedxm/config"
	rpc "github.com/ChrisVilches/freedxm/rpc/implementation"
	"github.com/urfave/cli/v3"
)

func ListSessions(_ context.Context, cmd *cli.Command) error {
	port := int(cmd.Int("port"))
	sessions, err := rpc.ListSessions(port)
	if err != nil {
		return err
	}

	fmt.Println(sessions)
	return nil
}

func ShowConfigFileContent(_ context.Context, _ *cli.Command) error {
	content, err := config.ReadConfigFileRaw()
	if err != nil {
		return err
	}

	fmt.Println(content)
	return nil
}
