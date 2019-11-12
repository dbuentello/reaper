package eventvwr

import (
	"fmt"
	"strconv"
)

// Parse is the initial entry point for all extended modules. All validation checks and processing will be performed here
// The function input types are limited to strings and therefore require additional processing
func Parse(options map[string]string) ([]string, error) {

	command, errCommand := GetJob()
	if errCommand != nil {
		return nil, fmt.Errorf("there was an error getting the EventVwr job:\r\n%s", errCommand.Error())
	}

	return command, nil
}

// GetJob returns a string array containing the commands, in the proper order, to be used with agents.AddJob
func GetJob(command string, ExecLocation string) ([]string, error) {
	return []string{"EventVwr", command, ExecLocation}, nil
}
