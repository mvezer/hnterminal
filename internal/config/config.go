package config

import (
	"fmt"
	utils "hnterminal/internal/utils"

	arg "github.com/alexflint/go-arg"
)

const DEFAULT_STORY_COUNT = 10
const DEFAULT_COMMAND = ""

var ValidCommands = [...]string{"top", "comment"}

var cliArgs struct {
	StoryCount int    `arg:"-c,--count" help:"Number of strories to show"`
	Command    string `arg:"positional" help:"Command to execute (top)"`
}

type Config struct {
	StoryCount int
	Command    string
}

var currentConfig = Config{
	StoryCount: DEFAULT_STORY_COUNT,
	Command:    DEFAULT_COMMAND,
}

/*
Parses and validates command line args
*/
func ParseArgs() {
	arg.MustParse(&cliArgs)
	isValidCommand := false
	if cliArgs.Command != "" {
		i := 0
		for !isValidCommand && i < len(ValidCommands) {
			if ValidCommands[i] == cliArgs.Command {
				isValidCommand = true
			}
			i++
		}
	} else {
		isValidCommand = true
	}
	if !isValidCommand {
		utils.HandleError(fmt.Errorf("unknown command \"%s\"\n", cliArgs.Command), utils.ErrorSeverityFatal)
	}

	currentConfig.Command = cliArgs.Command
	currentConfig.StoryCount = cliArgs.StoryCount
}

func IsTUI() bool {
	return cliArgs.Command == ""
}

func GetConfig() *Config {
	return &currentConfig
}
