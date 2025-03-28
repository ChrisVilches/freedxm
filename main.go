package main

import (
	"context"
	"github.com/ChrisVilches/freedxm/commands"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/status"
	"log"
	"os"
)

func getCmds() []*cli.Command {
	return []*cli.Command{
		{
			Name:    "serve",
			Aliases: []string{"s"},
			Usage:   "Start the server",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "port",
					Aliases: []string{"p"},
					Value:   int64(commands.ServeDefaultPort),
					Usage:   "Port number to listen on",
				},
			},
			Action: commands.Serve,
		},
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "Creates new blocking session",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:     "minutes",
					Aliases:  []string{"m"},
					Required: true,
					Usage:    "Number of minutes",
				},
				&cli.StringSliceFlag{
					Name:     "block-lists",
					Aliases:  []string{"b"},
					Required: true,
					Usage:    "Block lists to use",
				},
				&cli.IntFlag{
					Name:    "port",
					Aliases: []string{"p"},
					Value:   int64(commands.ServeDefaultPort),
					Usage:   "Port where the service is running on",
				},
			},
			Action: commands.AddSession,
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "List active sessions",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "port",
					Aliases: []string{"p"},
					Value:   int64(commands.ServeDefaultPort),
					Usage:   "Port where the service is running on",
				},
			},
			Action: commands.ListSessions,
		},
		{
			Name:    "show-config",
			Aliases: []string{"sc"},
			Usage:   "Show config file content",
			Action:  commands.ShowConfigFileContent,
		},
	}
}

func main() {
	cmd := &cli.Command{Commands: getCmds()}

	err := cmd.Run(context.Background(), os.Args)

	if err == nil {
		return
	}

	if grpcErr, ok := status.FromError(err); ok {
		mainMsg := "Error while communicating with the server"
		log.Fatalf("(%s) %s", mainMsg, grpcErr.Message())
	}

	log.Fatal(err)
}
