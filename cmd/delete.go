package main

import (
	"context"
	"flag"
	"peakbagger-tools/pbtools/config"
	"peakbagger-tools/pbtools/peakbagger"
	"peakbagger-tools/pbtools/terminal"

	"github.com/google/subcommands"
)

type deleteCmd struct {
	ascentID string
}

func (*deleteCmd) Name() string     { return "delete" }
func (*deleteCmd) Synopsis() string { return "Delete ascent(s) from peakbagger.com." }
func (*deleteCmd) Usage() string {
	return `delete [-id] <ascentId>
	Deletes peakbagger ascent(s).
  `
}

func (c *deleteCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.ascentID, "id", "", "peakbagger ascent id")
}

func (c *deleteCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	cfg := args[0].(*config.Config)

	pb := peakbagger.NewClient(cfg.PeakBaggerUsername, cfg.PeakBaggerPassword)

	// login to peakbagger
	o := terminal.NewOperation("Login to peakbagger.com with username '%s'", pb.Username)
	_, err := pb.Login()
	if err != nil {
		o.Error(err, "Failed to login to peakbagger.com")
		return 1
	}
	o.Success("Successfully logged in as '%s'", pb.Username)

	// delete ascent
	o = terminal.NewOperation("Deleting ascent id '%s'", c.ascentID)
	err = pb.DeleteAscent(c.ascentID)
	if err != nil {
		o.Error(err, "Failed to delete ascent id '%s'", c.ascentID)
		return 1
	}
	o.Success("Successfully deleted ascent id '%s'", c.ascentID)

	return 0
}
