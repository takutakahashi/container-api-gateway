package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "container gateway"
	app.Usage = "api generator for execute container"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "lang, l",
			Value: "english",
			Usage: "language for the greeting",
		},
		cli.StringFlag{
			Name:  "meridian, m",
			Value: "AM",
			Usage: "meridian for the greeting",
		},
		cli.StringFlag{
			Name:  "time, t",
			Value: "07:00",
			// ``で囲むとhelp時のPlaceholderとしても使える
			// https://github.com/urfave/cli#placeholder-values
			Usage: "`your time` for the greeting",
		},
		cli.StringFlag{
			Name:  "aaa, a",
			Value: "sample",
			// default値をValueからではなくEnvから取る
			EnvVar: "SAMPLE_ENV",
		},
	}
	app.Action = func(c *cli.Context) error {
		fmt.Println("-- Action --")

		fmt.Printf("c.NArg()        : %+v\n", c.NArg())
		fmt.Printf("c.Args()        : %+v\n", c.Args())
		fmt.Printf("c.Args().Get(0) : %+v\n", c.Args().Get(0))
		fmt.Printf("c.Args()[0]     : %+v\n", c.Args()[0])
		fmt.Printf("c.FlagNames     : %+v\n", c.FlagNames())

		// version表示
		cli.ShowVersion(c)
		return nil
	}

	app.Run(os.Args)
}
