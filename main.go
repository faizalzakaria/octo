package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
)

type Command struct {
	Name  string
	Run   func(c *cli.Context) error
	Flags []cli.Flag
	Short string
	Long  string
}

type AsgConfig struct {
	User     string   `yaml:"user"`
	Name     string   `yaml:"name"`
	AsgNames []string `yaml:"asgNames"`
	Region   string   `yaml:"region"`
}

var commands = []*Command{
	cmdSsh,
}

func main() {
	app := cli.NewApp()
	cmds := []*cli.Command{}

	setupCloseHandler()

	for _, cmd := range commands {
		if cmd.Name == "" {
			log.Fatal("No Name is specified for %s", cmd)
		}

		cliCommand := buildBasicCommand()
		cliCommand.Name = cmd.Name
		cliCommand.Usage = cmd.Short
		cliCommand.Description = cmd.Long
		cliCommand.Action = cmd.Run
		cliCommand.Flags = cmd.Flags
		cmds = append(cmds, &cliCommand)
	}

	app.Commands = cmds
	app.Name = "octo"
	app.Version = "1.1.0"
	app.Compiled = time.Now()
	app.Authors = []*cli.Author{
		&cli.Author{
			Name:  "Faizal Zakaria",
			Email: "fai@code3.io",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func buildBasicCommand() cli.Command {
	return cli.Command{}
}

func setupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()
}
