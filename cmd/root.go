package cmd

import (
	"fmt"
	"marsx/internal/tui"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "marsx",
	Short: "MarsX - AI powered terminal assistant",
	Long:  `MarsX is a CLI tool that helps you generate shell commands using AI.`,
	Run: func(cmd *cobra.Command, args []string) {
		tui.Start(quickMode, chatMode)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var (
	cfgFile   string
	quickMode bool
	chatMode  bool
	Version   = "dev" // set by build script
	Commit    = "none"
	Date      = "unknown"
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.marsx.yaml)")
	rootCmd.Flags().BoolVarP(&quickMode, "quick", "q", false, "Quickly generate commit message from staged changes")
	rootCmd.Flags().BoolVarP(&chatMode, "chat", "c", false, "Start directly in chat mode")
	rootCmd.Version = Version
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".") // Also look in current directory
		viper.SetConfigType("yaml")
		viper.SetConfigName(".marsx")
	}

	viper.SetEnvPrefix("marsx")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
