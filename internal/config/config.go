package config

import (
	"fmt"
	utils "hnterminal/internal/utils"
	"os"

	arg "github.com/alexflint/go-arg"
)

const DEFAULT_STORY_COUNT = 10
const DEFAULT_COMMAND = ""
const DEFAULT_DB_PATH = ""

var ValidCommands = [...]string{"top", "comment"}

var cliArgs struct {
	StoryCount int    `arg:"-c,--count" help:"Number of strories to show"`
	Command    string `arg:"positional" help:"Command to execute (top)"`
}

type Config struct {
	StoryCount int
	Command    string
	DbPath     string
}

var isConfigInitialized = false

var currentConfig = Config{}

func New() *Config {
	// set default config values
	currentConfig = Config{
		DEFAULT_STORY_COUNT,
		DEFAULT_COMMAND,
		getDefaultDbPath(),
	}
	parseConfig()
	parseArgs()
	isConfigInitialized = true
	return GetConfig()
}

/*
Parses and validates command line args
*/
func parseArgs() {
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

func parseConfig() {
	currentConfig.DbPath = getDefaultDbPath()
}

func IsTUI() bool {
	return currentConfig.Command == ""
}

func getDefaultDbPath() string {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		stateHome = fmt.Sprintf("%s/.state", os.Getenv("HOME"))
	}

	return fmt.Sprintf("%s/hacker-news-terminal", stateHome)
}

func GetConfig() *Config {
	if !isConfigInitialized {
		utils.HandleError(fmt.Errorf("Config is not initialized yet!"), utils.ErrorSeverityFatal)
	}
	return &currentConfig
}
