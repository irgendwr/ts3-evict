package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/irgendwr/go-ts3"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const name = "ts3-evict"
const description = "Evicts clients from a TeamSpeak 3 server after a given time; useful for demo servers."

//const longDescription = ""

var defaultPorts = []int{9987}

const defaultQueryPort = ts3.DefaultPort
const defaultHost = "127.0.0.1"
const defaultUser = "serveradmin"
const defaultViolators = "violators.csv"
const defaultAction = "kick"
const defaultTimelimit = 5
const defaultKicklimit = 3
const defaultBanDuration = 0
const defaultMessage = "Timelimit exceeded."
const defaultDelay = 5

// defaultCfgFile is the default config file name without extention
const defaultCfgFile = ".ts3-evict"
const defaultCfgFileType = "yaml"
const envPrefix = "ts3_evict"

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

// cfgFile contains the config file path if set by a CLI flag
var cfgFile string

// printVersion is true when version flag is set
var printVersion bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   name,
	Short: description,
	//Long: longDescription, //TODO: write description
	Run: func(cmd *cobra.Command, args []string) {
		if printVersion {
			fmt.Println(description)
			fmt.Println(buildVersion(version, commit, date))
			return
		}

		var cfg config
		if err := viper.Unmarshal(&cfg); err != nil {
			log.Fatalf("Error: Unable to read config: %s\n", err)
		}

		// if cfg.DefaultUsername == "" || cfg.DefaultPassword == "" {
		// 	log.Fatalln("Please set your server-query username and password in the config file (" + defaultCfgFile + "." + defaultCfgFileType + ").")
		// }

		if !(cfg.Action == "kick" || cfg.Action == "ban" || cfg.Action == "kick or ban" || cfg.Action == "none") {
			log.Fatalln("Error: Please set a valid action: either kick, ban, 'kick or ban' or none.")
		}

		if len(cfg.KickMessage) > 80 {
			log.Fatalln("Error: Kick message too long. Use 80 chars or less.")
		}

		if len(cfg.BanMessage) > 80 {
			log.Fatalln("Error: Ban message too long. Use 80 chars or less.")
		}

		if err := evict(cfg); err != nil {
			log.Fatalf("Error: %s\n", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Define flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is "+defaultCfgFile+"."+defaultCfgFileType+" in program dir, CWD or $HOME)")
	rootCmd.PersistentFlags().StringP("action", "a", defaultAction, "action (none/kick/ban/kick or ban)")
	rootCmd.PersistentFlags().StringP("message", "m", defaultMessage, "message")
	rootCmd.Flags().BoolVarP(&printVersion, "version", "v", false, "show version and exit")

	// Bind flags to config and set default values
	viper.SetDefault("defaultqueryport", defaultQueryPort)
	viper.SetDefault("defaultports", defaultPorts)
	viper.SetDefault("defaultusername", defaultUser)
	viper.SetDefault("defaultviolators", defaultViolators)
	viper.BindPFlag("action", rootCmd.PersistentFlags().Lookup("action"))
	viper.SetDefault("action", defaultAction)
	viper.SetDefault("timelimit", defaultTimelimit)
	viper.SetDefault("kicklimit", defaultKicklimit)
	viper.SetDefault("banduration", defaultBanDuration)
	viper.BindPFlag("message", rootCmd.PersistentFlags().Lookup("message"))
	viper.SetDefault("message", defaultMessage)
	viper.SetDefault("delay", defaultDelay)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if printVersion {
		// skip reading config when printVersionis set
		return
	}
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in directory of program with name ".sb-spotify.yaml"
		viper.SetConfigName(defaultCfgFile)
		viper.SetConfigType(defaultCfgFileType)

		if ex, err := os.Executable(); err == nil {
			viper.AddConfigPath(ex)
		}

		// And also in CWD and home dir
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
	}

	// Read in environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func buildVersion(version, commit, date string) string {
	var result = fmt.Sprintf("version: %s", version)
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}
	if builtBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, builtBy)
	}
	return result
}
