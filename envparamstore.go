package main

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"strings"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"os/exec"
	"os"
)

var (
	prefix = kingpin.Flag("prefix", "Inject keys that match the given prefix").String()
	leaveEncrypted = kingpin.Flag("leave-encrypted", "Do not decrypt parameter values").Bool()
	command = kingpin.Flag("cmd", "Command to run").Required().Strings()
	pristine = kingpin.Flag("pristine", "Do not inject surrounding process's environment").Bool()
)

func main() {
	//Parse the command line arguments
	kingpin.Parse()

	//Establish the session with AWS
	session, err := session.NewSession()
	if err != nil {
		log.Fatal(err.Error())
	}

	svc := ssm.New(session)

	paramStoreEnv,err := extractParamStoreEnv(*prefix, *leaveEncrypted, svc)
	if err != nil {
		log.Fatal(err.Error())
	}

	cmdEnv := getCommandEnv(*pristine, paramStoreEnv)

	log.Infof("%v", cmdEnv)

	err = runCommand(*command, cmdEnv)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getCommandEnv(pristine bool, paramStoreEnv []string) []string {
	newEnv := make(map[string]string)

	if !pristine {
		for _, v := range os.Environ() {
			list := strings.SplitN(v, "=", 2)
			newEnv[list[0]] = list[1]
		}
	}

	for _, v := range paramStoreEnv {
		list := strings.SplitN(v, "=", 2)
		newEnv[list[0]] = list[1]
	}

	var finalEnv []string
	for k, v := range newEnv {
		finalEnv = append(finalEnv, fmt.Sprintf("%s=%s", k, v))
	}

	return finalEnv
}

func extractParamStoreEnv(keyPrefix string, doNotDecrypt bool, svc *ssm.SSM) ([]string,error) {
	var cmdEnv []string
	var decryptFlag = !doNotDecrypt

	params := &ssm.DescribeParametersInput{}
	for {
		resp, err := svc.DescribeParameters(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil,err
		}

		parameterMetadata := resp.Parameters
		for _, pmd := range parameterMetadata {
			if keyPrefix != "" && !strings.HasPrefix(*pmd.Name, keyPrefix) {
				log.Infof("skipping %s", *pmd.Name)
				continue
			}

			keyMinusPrefix := (*pmd.Name)[len(keyPrefix):]
			log.Infof("Injecting %s as %s", *pmd.Name, keyMinusPrefix)

			params := &ssm.GetParametersInput{
				Names: []*string{
					pmd.Name,
				},
				WithDecryption: aws.Bool(decryptFlag),
			}

			resp, err := svc.GetParameters(params)
			if err != nil {
				return nil,err
			}

			paramVal := resp.Parameters[0].Value
			cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", keyMinusPrefix, *paramVal))
		}

		nextToken := resp.NextToken
		if nextToken == nil {
			break
		}

		params = &ssm.DescribeParametersInput{
			NextToken:nextToken,
		}
	}

	return cmdEnv,nil
}

func runCommand(commandArgs []string, cmdEnv []string) error {
	log.Infof("Run command with env %v", cmdEnv)
	cmd := exec.Command(commandArgs[0], commandArgs[1:]...)
	cmd.Env = cmdEnv
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}