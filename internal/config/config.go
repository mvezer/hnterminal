package config

import (
	arg "github.com/alexflint/go-arg"
)

var cliArgs struct {
	Count   int    `arg:"-c,--count" help:"number of items to show"`
	Command string `arg:"positional" help:"command to execute"`
}

type Config struct {
	Count   int
	Command string
}

func ParseArgs() {
	arg.MustParse(&cliArgs)
}

func IsTUI() bool {
	return cliArgs.Command == ""
}

func GetConfig() *Config {
	return &Config{
		Count:   cliArgs.Count,
		Command: cliArgs.Command,
	}
}
