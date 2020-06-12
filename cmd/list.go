package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"peakbagger-tools/pbtools/config"
	"peakbagger-tools/pbtools/convert"
	"peakbagger-tools/pbtools/peakbagger"
	"peakbagger-tools/pbtools/terminal"
	"strconv"

	"github.com/google/subcommands"
)

type listCmd struct {
	format     string
	outputFile string
}

const (
	jsonF = "json"
	textF = "text"
	csvF  = "csv"
)

const dateFormat = "01/02/2006"

func (*listCmd) Name() string     { return "list" }
func (*listCmd) Synopsis() string { return "List personal ascent(s) from peakbagger.com." }
func (*listCmd) Usage() string {
	return `list
	List personal peakbagger ascents.
  `
}

func (c *listCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.format, "format", "text", "format to display ascents (json, text, csv)")
	f.StringVar(&c.outputFile, "output", "", "output file")
}

func (c *listCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	cfg := args[0].(*config.Config)

	pb := peakbagger.NewClient(cfg.PeakBaggerUsername, cfg.PeakBaggerPassword)

	// validate parameters
	switch c.format {
	case jsonF, textF, csvF:
	default:
		terminal.Error(nil, "Invalid format '%s'", c.format)
		return 1
	}

	// login to peakbagger
	o := terminal.NewOperation("Login to peakbagger.com with username '%s'", pb.Username)
	_, err := pb.Login()
	if err != nil {
		o.Error(err, "Failed to login to peakbagger.com")
		return 1
	}
	o.Success("Successfully logged in as '%s'", pb.Username)

	// list ascent
	o = terminal.NewOperation("Listing ascents from peakbagger.com")
	ascents, err := pb.ListAscents()
	if err != nil {
		o.Error(err, "Failed to list ascents")
		return 1
	}
	o.Success("Successfully listed %d ascents", len(ascents))

	// get a file writer if needed
	var w io.Writer = os.Stdout
	var op *terminal.Operation
	if c.outputFile != "" {
		f, err := os.Create(c.outputFile)
		w = f
		defer f.Close()
		if err != nil {
			terminal.Error(err, "Could not open file '%s'", c.outputFile)
			return 1
		}

		op = terminal.NewOperation("Exporting list to '%s' in %s format", c.outputFile, c.format)
	}

	// print result
	switch c.format {
	case textF:
		for _, a := range ascents {
			fmt.Fprintf(w, "%s - %s (%d') - %s\n", a.Date.Format(dateFormat), a.PeakName, int(convert.ToFeet(a.Elevation)), a.Location)
		}
	case jsonF:
		elts := make([]map[string]interface{}, len(ascents))
		for i, a := range ascents {
			jsonMap := map[string]interface{}{}
			jsonMap["ascent_id"] = a.AscentID
			jsonMap["peak_id"] = a.PeakID
			jsonMap["peak_name"] = a.PeakName
			jsonMap["date"] = a.Date.Format(dateFormat)
			jsonMap["elevation"] = int(convert.ToFeet(a.Elevation))
			jsonMap["location"] = a.Location
			elts[i] = jsonMap
		}
		jsonStr, _ := json.MarshalIndent(elts, "", "  ")
		fmt.Fprint(w, string(jsonStr))
	case csvF:
		csvW := csv.NewWriter(w)
		csvW.Write([]string{"date", "peak id", "peak name", "elevation(feet)", "location", "ascent id"})
		for _, a := range ascents {
			csvW.Write([]string{a.Date.Format(dateFormat), a.PeakID, a.PeakName, strconv.Itoa(int(convert.ToFeet(a.Elevation))), a.Location, a.AscentID})
		}
		csvW.Flush()
	}

	if op != nil {
		op.Success("List exported to %s", c.outputFile)
	}

	return 0
}
