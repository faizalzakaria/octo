package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"runtime"
	"sort"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/urfave/cli/v2"
)

var cmdSsh = &Command{
	Name: "ssh",
	Run:  runSsh,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "stack",
			Aliases: []string{"s"},
			Usage:   "",
		},
		&cli.StringFlag{
			Name:    "environment",
			Aliases: []string{"e"},
			Usage:   "full or partial environment name",
		},
		&cli.StringFlag{
			Name:        "user",
			Aliases:     []string{"u"},
			Usage:       "User to ssh with",
			DefaultText: "user as per in the ~/.octo/config.yml file",
		},
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Config file",
			Value:   "~/.octo/config.yml",
		},
		&cli.StringFlag{
			Name:    "app",
			Aliases: []string{"a"},
			Usage:   "App Name, this will read the file ~/.octo.<app anme>.yml",
			Value:   "",
		},
	},
	Short: "starts a ssh shell into the server",
	Long: `This will prompt some action needed, with option to choose which environment and stack
that you want to connect.

Alternatively, you have option to set a specific environment and stack from command line.

This is only works in Linux based OS.

Examples:
$ octo ssh -s api -e staging
$ octo ssh -s web -e production
`,
}

func runSsh(ctx *cli.Context) error {
	if runtime.GOOS == "windows" {
		log.Fatal("Not supported on Windows")
		os.Exit(2)
	}

	stack := ctx.String("stack")
	environment := ctx.String("environment")
	configFile := ctx.String("config")
	sshUser := ctx.String("user")
	appName := ctx.String("app")

	// load asgConfigs
	if len(appName) > 0 {
		configFile = "~/.octo/config." + appName + ".yml"
		fmt.Printf("Config file is reconfigured to %s\n", configFile)
	}
	asgConfigs := map[string]map[string]*AsgConfig{}
	loadAsgConfigs(configFile, &asgConfigs)

	// Get environments & stacks
	environments := []string{}
	stacks := []string{}
	for k, v := range asgConfigs {
		environments = append(environments, k)
		for s := range v {
			stacks = append(stacks, s)
		}
	}
	environments = uniqueStrings(environments)
	sort.Strings(environments)
	stacks = uniqueStrings(stacks)
	sort.Strings(stacks)

	// If no environment set,
	if len(environment) <= 0 {
		selectOption("Environment", &environment, environments)
	}

	// If no stack set,
	if len(stack) <= 0 {
		selectOption("Stack", &stack, stacks)
	}

	// Get the asg config for a given environment & stack
	asgConfig := asgConfigs[environment][stack]
	if asgConfig == nil {
		fmt.Println("Invalid ASG Config, check your ~/.octo/config.yml file")
		return nil
	}

	printAsgConfig(asgConfig)

	// Get instances
	instances := getInstances(asgConfig)

	// Get the ssh user
	if len(sshUser) <= 0 {
		if len(asgConfig.User) >= 0 {
			sshUser = asgConfig.User
		} else {
			user, _ := user.Current()
			sshUser = user.Username
		}
	}

	// Show list of instances to choose
	var instanceToSSH *ec2.Instance
	selectInstance(instanceToSSH, instances)

	// ssh to the server
	sshToServer(sshUser, instanceToSSH, 0)

	return nil
}

func selectInstance(instance *ec2.Instance, instances []*ec2.Instance) {
	if len(instances) <= 0 {
		fmt.Println("No instance found")
	} else if len(instances) == 1 {
		instance = instances[0]
	} else {
		printInstances(instances)

		fmt.Println("\tWhich instance ? ")
		var i int
		fmt.Scanf("%d", &i)
		instance = instances[i]
	}
}

func selectOption(label string, option *string, options []string) {
	if len(options) <= 0 {
		fmt.Printf("No %s found", label)
	} else if len(options) == 1 {
		*option = options[0]
	} else {
		fmt.Printf("\tselect %s\n", label)

		printStringList(options)

		fmt.Printf("\tWhich %s ? ", label)
		var i int
		fmt.Scanf("%d", &i)
		fmt.Println("\t", options[i])
		*option = options[i]
	}

}

func sshToServer(sshUser string, instance *ec2.Instance, verbosity int) error {
	instanceIp := *instance.PrivateIpAddress

	fmt.Printf("\nConnecting to %s@%s ...\n", sshUser, instanceIp)

	vflag := ""
	if verbosity == 1 {
		vflag = "-v"
	} else if verbosity == 2 {
		vflag = "-vv"
	} else if verbosity == 3 {
		vflag = "-vvv"
	}

	return startProgram("ssh", []string{
		sshUser + "@" + instanceIp,
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "CheckHostIP=no",
		"-o", "StrictHostKeyChecking=no",
		"-o", "LogLevel=QUIET",
		"-o", "IdentitiesOnly=yes",
		"-A",
		"-p", "22",
		vflag,
	})
}

func printStringList(list []string) {
	fmt.Println("\t------------------")
	for idx, str := range list {
		fmt.Printf("\t%d: %s\n", idx, str)
	}
	fmt.Println("\t------------------\n")
}

func printInstances(instances []*ec2.Instance) {
	fmt.Println("\t------------------")
	for idx, inst := range instances {
		fmt.Printf("\t%d: %s\n", idx, *inst.PrivateIpAddress)
	}
	fmt.Println("\t------------------\n")
}

func printAsgConfig(asgConfig *AsgConfig) {
	fmt.Println("\t------------------")
	fmt.Println("\tASG Config")
	fmt.Println("\t------------------")
	fmt.Println("\tName: ", asgConfig.Name)
	fmt.Println("\tRegion: ", asgConfig.Region)
	fmt.Println("\tUser: ", asgConfig.User)
	fmt.Println("\tAsgNames: ", asgConfig.AsgNames)
	fmt.Println("\t------------------\n")
}
