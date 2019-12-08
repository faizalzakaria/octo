package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"runtime"

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
	Long: `This will open the firewall for SSH from your IP address temporaritly (20 minutes), downloads the keys if you don't have them
and starts a SSH session.

You need to have the right access permissions to use this command.
You can use either the server name (ie lion) or the server IP (ie. 123.123.123.123) or the server role (ie. web)
with thie command.

If a role is specified the command will connect to the first server with that role.

Names are case insensitive and will work with the starting characters as well.

You should provide a key to your bastion server if it is deployed with a deploy gateway.

This command is only supported on Linux and OS X.

Examples:
$ octo ssh -s api -e staging
$ octo ssh -s web -e production
`,
}

func runSsh(c *cli.Context) error {
	if runtime.GOOS == "windows" {
		log.Fatal("Not supported on Windows")
		os.Exit(2)
	}

	stack := c.String("stack")
	environment := c.String("environment")
	configFile := c.String("config")
	sshUser := c.String("user")
	appName := c.String("app")

	if len(stack) <= 0 {
		fmt.Println("Stack is missing, using default stack")
		stack = "default"
	}

	if len(environment) <= 0 {
		fmt.Println("Environment is missing")
		return nil
	}

	if len(appName) > 0 {
		configFile = "~/.octo/config." + appName + ".yml"
		fmt.Printf("Config file is reconfigured to %s\n", configFile)
	}

	asgConfigs := map[string]map[string]*AsgConfig{}
	loadAsgConfigs(configFile, &asgConfigs)

	asgConfig := asgConfigs[environment][stack]

	if asgConfig == nil {
		fmt.Println("Invalid ASG Config, check your ~/.octo/config.yml file")
		return nil
	}

	printAsgConfig(asgConfig)

	instances := getInstances(asgConfig)

	if len(sshUser) <= 0 {
		if len(asgConfig.User) >= 0 {
			sshUser = asgConfig.User
		} else {
			user, _ := user.Current()
			sshUser = user.Username
		}
	}

	var instanceToSSH *ec2.Instance
	if len(instances) <= 0 {
		fmt.Println("No instance found")
		return nil
	} else if len(instances) == 1 {
		instanceToSSH = instances[0]
	} else {
		printInstances(instances)

		fmt.Println("\tWhich instance ? ")
		var i int
		fmt.Scanf("%d", &i)
		instanceToSSH = instances[i]
	}

	sshToServer(sshUser, instanceToSSH, 0)

	return nil
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
