package command

import (
	"flag"
	"fmt"
	"github.com/bluefw/blued/client"
	"github.com/mitchellh/cli"
	"strings"
)

// SyncCommand is a Command implementation that tells a running Blued
// agent to sync router table from another.
type SyncCommand struct {
	Ui cli.Ui
}

func (c *SyncCommand) Help() string {
	helpText := `
Usage: blued sync [options] node ...

  Tells a running Blued agent (with "blued agent") to sync router table
  by specifying at least one existing member name.

Options:
  -rpc-addr=127.0.0.1:7373  RPC address of the Blued agent.
  -rpc-auth=""              RPC auth token of the Blued agent.
`
	return strings.TrimSpace(helpText)
}

func (c *SyncCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("sync", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	rpcAddr := RPCAddrFlag(cmdFlags)
	rpcAuth := RPCAuthFlag(cmdFlags)
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	nodes := cmdFlags.Args()
	if len(nodes) == 0 {
		c.Ui.Error("At least one member to sync must be specified.")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	client, err := RPCClient(*rpcAddr, *rpcAuth)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Serf agent: %s", err))
		return 1
	}
	defer client.Close()

	syncAddr := c.SyncRouters(client, *rpcAuth, nodes)
	if syncAddr != "" {
		c.Ui.Output(fmt.Sprintf("Successfully sync router table from %s", syncAddr))
	} else {
		c.Ui.Output("No router table is sync")
	}
	return 0
}

func (c *SyncCommand) SyncRouters(cl *client.RPCClient, auth string, nodes []string) string {
	syncAddr := ""
	respCh := make(chan client.NodeResponse, 64)
	params := client.QueryParam{
		Name:        "qr",
		FilterNodes: nodes,
		RequestAck:  false,
		RespCh:      respCh,
	}
	if err := cl.Query(&params); err != nil {
		c.Ui.Error(fmt.Sprintf("Error sending query: %s", err))
		return syncAddr
	}

OUTER:
	for {
		select {
		case r := <-respCh:
			if r.From == "" {
				break OUTER
			}
			if syncAddr == "" {
				err := c.updateRouters(cl, string(r.Payload), auth)
				if err == nil {
					syncAddr = string(r.Payload)
				}
			}
		}
	}

	return syncAddr
}

func (c *SyncCommand) updateRouters(cl *client.RPCClient, addr string, auth string) error {
	tc, err := RPCClient(addr, auth)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to %s: %s", addr, err))
		return err
	}
	defer tc.Close()

	rs, err := tc.ListRouters()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error getting router table from %s: %s", addr, err))
		return err
	}

	err = cl.UpdateRouters(rs)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error updating router table from %s: %s", addr, err))
	}
	return err
}

func (c *SyncCommand) Synopsis() string {
	return "Tell Serf agent to join cluster"
}
