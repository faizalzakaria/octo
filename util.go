package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"gopkg.in/yaml.v2"
)

var debugMode = false

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func startProgram(command string, args []string) error {
	if runtime.GOOS != "windows" {
		p, err := exec.LookPath(command)
		if err != nil {
			log.Printf("Error finding path to %q: %s\n", command, err)
			os.Exit(2)
		}
		command = p
	}

	if debugMode {
		fmt.Printf("Running Command %s with (%s)\n", command, args)
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// Get instances
func getInstances(asgConfig *AsgConfig) []*ec2.Instance {
	sess, err := session.NewSession()
	svc := ec2.New(sess, aws.NewConfig().WithRegion(asgConfig.Region))

	var asgNames []*string
	for _, asgName := range asgConfig.AsgNames {
		asgNames = append(asgNames, aws.String(asgName))
	}

	filters := []*ec2.Filter{
		{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running"), aws.String("pending")},
		},
		{
			Name:   aws.String("tag:Name"),
			Values: asgNames,
		},
	}

	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	result, err := svc.DescribeInstances(input)

	if err != nil {
		fmt.Println("there was an error listing instances in", err.Error())
		log.Fatal(err.Error())
	}

	var instances []*ec2.Instance
	for _, res := range result.Reservations {
		for _, inst := range res.Instances {
			instances = append(instances, inst)
		}
	}

	return instances
}

func loadAsgConfigs(configFile string, asgConfigs *map[string]map[string]*AsgConfig) {
	fullFilePath, err := homedir.Expand(configFile)
	check(err)

	data, err := ioutil.ReadFile(fullFilePath)
	check(err)

	if debugMode {
		fmt.Print(string(data))
	}

	err = yaml.Unmarshal([]byte(data), asgConfigs)
	check(err)
}

func uniqueStrings(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
