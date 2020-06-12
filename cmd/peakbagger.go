package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"peakbagger-tools/pbtools/config"
	t "peakbagger-tools/pbtools/terminal"

	"github.com/google/subcommands"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&addCmd{}, "")
	subcommands.Register(&deleteCmd{}, "")
	subcommands.Register(&listCmd{}, "")

	cfg, err := config.Load()
	if err != nil {
		t.Error(nil, "Failed to load config")
	}

	err = getPeakbaggerCredentials(&cfg.PeakBaggerUsername, &cfg.PeakBaggerPassword)
	if err != nil {
		t.Error(err, "Failed to get peakbagger credentials")
	}

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx, cfg)))
}

// Fetch peakbagger credentials from a config file located in the home directory.
// If the file doesn't exist, prompt user with username and password, and save it
// to this file.
func getPeakbaggerCredentials(username *string, password *string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fileName := home + "/.peakbagger"
	var user string
	var pwd string

	data, err := readLines(fileName)
	if err != nil {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("No peakbagger credentials file found.")
		fmt.Print("Enter Username: ")
		user, _ = reader.ReadString('\n')

		fmt.Print("Enter Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		pwd := string(bytePassword)
		fmt.Println()

		err = writeLines([]string{
			"username=" + user,
			"password=" + pwd,
		}, fileName)
		if err != nil {
			return err
		}
	} else {
		if len(data) == 2 && strings.HasPrefix("username=", data[0]) && strings.HasPrefix("password=", data[1]) {
			return errors.New("wrong credentials file format")
		}

		user = string(data[0][9:])
		pwd = string(data[1][9:])
	}

	*username = user
	*password = pwd

	return nil
}

// read lines from a file
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// write lines to a file
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		w.WriteString(line)
	}
	return w.Flush()
}
