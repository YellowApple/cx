// +build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdServerSet = &Command{
	Run:        runServerSet,
	Usage:      "server-set <server name>|<server ip>|<server role> <setting> <value>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "sets the value of a setting on a server",
	Long: `This sets and applies the value of a setting on a server. Applying some settings might require more
work and therefore this command will return immediately after the setting operation has started.

Examples:
$ cx server-set -s mystack lion server.name tiger
`,
}

func runServerSet(cmd *Command, args []string) {

	stack := mustStack()

	if len(args) != 3 {
		cmd.printUsage()
		os.Exit(2)
	}

	// get the server
	serverName := args[0]

	servers, err := client.Servers(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}

	server, err := findServer(servers, serverName)
	if err != nil {
		printFatal(err.Error())
	}

	if server == nil {
		printFatal("Server '" + serverName + "' not found")
	}

	fmt.Printf("Server: %s\n", server.Name)

	executeServerSet(*stack, *server, args)
}

func executeServerSet(stack cloud66.Stack, server cloud66.Server, args []string) {

	// filter out the server name
	args = args[1:len(args)]

	key := args[0]
	value := args[1]

	settings, err := client.ServerSettings(stack.Uid, server.Uid)
	must(err)

	// check to see if it's a valid setting
	for _, i := range settings {
		if key == i.Key {
			// yup. it's a good one
			fmt.Printf("Please wait while your setting is applied...\n")

			asyncId, err := startServerSet(stack.Uid, server.Uid, key, value)
			if err != nil {
				printFatal(err.Error())
			}
			genericRes, err := endServerSet(*asyncId, stack.Uid)
			if err != nil {
				printFatal(err.Error())
			}
			printGenericResponse(*genericRes)

			return
		}
	}

	printFatal(key + " is not a valid setting or does not apply to this server")
}

func startServerSet(stackUid string, serverUid string, key string, value string) (*int, error) {
	asyncRes, err := client.ServerSet(stackUid, serverUid, key, value)
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endServerSet(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}
