package main

import (
	"log"
	"os"

	"github.com/takutakahashi/container-api-gateway/pkg/api"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "container gateway"
	app.Usage = "api generator for execute container"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "config.yaml filepath",
		},
	}
	app.Action = action
	app.Run(os.Args)
}

func action(c *cli.Context) error {
	configPath := c.String("config")
	if configPath == "" {
		cli.ShowAppHelp(c)
		return nil
	}
	server := api.Server{}
	err := server.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	server.Start()
	return nil
}
