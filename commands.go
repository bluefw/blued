package main

import (
	"github.com/bluefw/blued/command"
	"github.com/bluefw/blued/command/agent"
	"github.com/mitchellh/cli"
	"os"
	"os/signal"
)

// Commands is the mapping of all the available Serf commands.
var Commands map[string]cli.CommandFactory

func init() {
	ui := &cli.BasicUi{Writer: os.Stdout}

	Commands = map[string]cli.CommandFactory{
		"agent": func() (cli.Command, error) {
			return &agent.Command{
				Ui:         ui,
				ShutdownCh: make(chan struct{}),
			}, nil
		},

		"event": func() (cli.Command, error) {
			return &command.EventCommand{
				Ui: ui,
			}, nil
		},

		"query": func() (cli.Command, error) {
			return &command.QueryCommand{
				ShutdownCh: makeShutdownCh(),
				Ui:         ui,
			}, nil
		},

		"force-leave": func() (cli.Command, error) {
			return &command.ForceLeaveCommand{
				Ui: ui,
			}, nil
		},

		"join": func() (cli.Command, error) {
			return &command.JoinCommand{
				Ui: ui,
			}, nil
		},

		"keygen": func() (cli.Command, error) {
			return &command.KeygenCommand{
				Ui: ui,
			}, nil
		},

		"keys": func() (cli.Command, error) {
			return &command.KeysCommand{
				Ui: ui,
			}, nil
		},

		"apps": func() (cli.Command, error) {
			return &command.AppsCommand{
				Ui: ui,
			}, nil
		},

		"routers": func() (cli.Command, error) {
			return &command.RoutersCommand{
				Ui: ui,
			}, nil
		},

		"sync": func() (cli.Command, error) {
			return &command.SyncCommand{
				Ui: ui,
			}, nil
		},
		
		"leave": func() (cli.Command, error) {
			return &command.LeaveCommand{
				Ui: ui,
			}, nil
		},

		"members": func() (cli.Command, error) {
			return &command.MembersCommand{
				Ui: ui,
			}, nil
		},

		"monitor": func() (cli.Command, error) {
			return &command.MonitorCommand{
				ShutdownCh: makeShutdownCh(),
				Ui:         ui,
			}, nil
		},

		"tags": func() (cli.Command, error) {
			return &command.TagsCommand{
				Ui: ui,
			}, nil
		},

		"reachability": func() (cli.Command, error) {
			return &command.ReachabilityCommand{
				ShutdownCh: makeShutdownCh(),
				Ui:         ui,
			}, nil
		},

		"rtt": func() (cli.Command, error) {
			return &command.RTTCommand{
				Ui: ui,
			}, nil
		},

		"info": func() (cli.Command, error) {
			return &command.InfoCommand{
				Ui: ui,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Revision:          GitCommit,
				BluedVer:          BluedVer,
				SerfVer:           SerfVer,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
	}
}

// makeShutdownCh returns a channel that can be used for shutdown
// notifications for commands. This channel will send a message for every
// interrupt received.
func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}
