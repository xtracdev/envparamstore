# envparamstore

envparamstore provides a convenient way to populate environment variable
values from the [AWS SSM Parameter Store](https://aws.amazon.com/ec2/systems-manager/parameter-store/)

This allows processes that read  their configuration from environment variables
to use values stored in the AWS parameter store without being aware of the
Parameter store.

Usage:

<pre>
go run envparamstore.go --help
usage: envparamstore --cmd=CMD [<flags>]

Flags:
  --help             Show context-sensitive help (also try --help-long and
                     --help-man).
  --prefix=PREFIX    Inject keys that match the given prefix
  --leave-encrypted  Do not decrypt parameter values
  --cmd=CMD ...      Command to run
  --pristine         Do not inject surrounding process's environment
</pre>

## Configuration

The configuration for AWS is taken from the environment. You can specify the
profile and region using the AWS\_PROFILE and AWS\_REGION environment
variables, or you can export AWS\_SDK\_LOAD\_CONFIG=1 and the 
credentials will be obtained in the same manner as the AWS CLI, for
example using AWS\_DEFAULT\_PROFILE and AWS\_DEFAULT\_REGION.


## Trying it out

Inject some parameter store variables

<pre>
aws ssm put-parameter --name demo.PARAM1 --value 'Param 1 Value' --type String
aws ssm put-parameter --name demo.PARAM2 --value 'Param 2 Value' --type String

go run envparamstore.go --prefix demo. --pristine --cmd env
INFO[0000] Injecting demo.PARAM1 as PARAM1              
INFO[0000] Injecting demo.PARAM2 as PARAM2              
INFO[0001] [PARAM1=Param 1 Value PARAM2=Param 2 Value]  
PARAM1=Param 1 Value
PARAM2=Param 2 Value
</pre>

For encrypting parameter values, an encryption key id is needed. You can
get the id for a key via `aws kms list-keys` and `aws kms describe-key`

<pre>
aws ssm put-parameter --name demo.my_secret --value 'loose lips sink ships' --type SecureString --key-id <key id>

go run envparamstore.go --prefix demo. --pristine --cmd env
INFO[0001] Injecting demo.PARAM1 as PARAM1              
INFO[0001] Injecting demo.PARAM2 as PARAM2              
INFO[0001] Injecting demo.my_secret as my_secret        
INFO[0001] [my_secret=loose lips sink ships PARAM1=Param 1 Value PARAM2=Param 2 Value] 
my_secret=loose lips sink ships
PARAM1=Param 1 Value
PARAM2=Param 2 Value

go run envparamstore.go --prefix demo. --pristine --leave-encrypted --cmd env
INFO[0001] Injecting demo.PARAM1 as PARAM1              
INFO[0001] Injecting demo.PARAM2 as PARAM2              
INFO[0001] Injecting demo.my_secret as my_secret        
INFO[0001] [PARAM1=Param 1 Value PARAM2=Param 2 Value my_secret=AQECAHgDl5254172VojbuHtk5XKZan719pdjNXEuV+C4K+004AAAAHMwcQYJKoZIhvcNAQcGoGQwYgIBADBdBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDMx+Ncz907Yf7Dx6eAIBEIAwZkiudD8Thv+yGjFm/jjvWRp/hludgBcLngi5432MTL90JCk5vGkatLkpINFYZSUR] 
PARAM1=Param 1 Value
PARAM2=Param 2 Value
my_secret=AQECAHgDl5254172VojbuHtk5XKZan719pdjNXEuV+C4K+004AAAAHMwcQYJKoZIhvcNAQcGoGQwYgIBADBdBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDMx+Ncz907Yf7Dx6eAIBEIAwZkiudD8Thv+yGjFm/jjvWRp/hludgBcLngi5432MTL90JCk5vGkatLkpINFYZSUR
</pre>





