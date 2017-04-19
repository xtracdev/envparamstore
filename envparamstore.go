package main

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"strings"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
)

var (
	prefix = kingpin.Flag("prefix", "Inject keys that match the given prefix").String()
	leaveEncrypted = kingpin.Flag("leave-encrypted", "Do not decrypt parameter values").Bool()
	command = kingpin.Flag("cmd", "Command to run").Required().Strings()
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

	cmdEnv,err := extractEnv(*prefix, *leaveEncrypted, svc)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Infof("%v", cmdEnv)
}

func extractEnv(keyPrefix string, doNotDecrypt bool, svc *ssm.SSM) ([]string,error) {
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