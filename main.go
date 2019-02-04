package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
	"slurm-client-go/command"
)

func main() {
	app := cli.NewApp()
	app.Name = "slurm_client_go"
	app.Usage = "slurm client"
	app.Action = func(c *cli.Context) error {
		command.QueryCommand()
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
