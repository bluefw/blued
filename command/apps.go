package command

import (
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"strings"
)

// InfoCommand is a Command implementation that queries a running
// Serf agent for various debugging statistics for operators
type AppsCommand struct {
	Ui cli.Ui
}

func (i *AppsCommand) Help() string {
	helpText := `
Usage: blued apps [options]

	Provides debugging information for operators

Options:

  -format                  If provided, output is returned in the specified
                           format. Valid formats are 'json', and 'text' (default)

  -rpc-addr=127.0.0.1:7373 RPC address of the Serf agent.

  -rpc-auth=""             RPC auth token of the Serf agent.
`
	return strings.TrimSpace(helpText)
}

func (i *AppsCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("service", flag.ContinueOnError)
	cmdFlags.Usage = func() { i.Ui.Output(i.Help()) }
	rpcAddr := RPCAddrFlag(cmdFlags)
	rpcAuth := RPCAuthFlag(cmdFlags)
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	client, err := RPCClient(*rpcAddr, *rpcAuth)
	if err != nil {
		i.Ui.Error(fmt.Sprintf("Error connecting to Serf agent: %s", err))
		return 1
	}
	defer client.Close()

	stats, err := client.ListMicroApps()
	if err != nil {
		i.Ui.Error(fmt.Sprintf("Error querying agent: %s", err))
		return 1
	}

	output, err := formatOutput(stats, "json")
	if err != nil {
		i.Ui.Error(fmt.Sprintf("Encoding error: %s", err))
		return 1
	}

	i.Ui.Output(string(output))
	return 0
}

func (i *AppsCommand) Synopsis() string {
	return "Provides debugging information for operators"
}
