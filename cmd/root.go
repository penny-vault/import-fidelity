// Copyright 2022
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "import-fidelity",
	Short: "Download stock data from Fidelity website",
	Long: `import-fidelity contains a collection of tools to scrape various bits of
stock information from fidelities website. Most actions require authentication to complete.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLog)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is import-fidelity.toml)")

	rootCmd.PersistentFlags().Bool("log.json", false, "print logs as json to stderr")
	if err := viper.BindPFlag("log.json", rootCmd.PersistentFlags().Lookup("log.json")); err != nil {
		log.Error().Err(err).Msg("bind log.json")
	}

	rootCmd.PersistentFlags().Bool("log.debug", false, "use debug log level")
	if err := viper.BindPFlag("log.debug", rootCmd.PersistentFlags().Lookup("log.debug")); err != nil {
		log.Error().Err(err).Msg("bind log.debug")
	}

	rootCmd.PersistentFlags().String("backblaze-application-id", "<not-set>", "backblaze application id")
	if err := viper.BindPFlag("backblaze.application_id", rootCmd.PersistentFlags().Lookup("backblaze-application-id")); err != nil {
		log.Error().Err(err).Msg("bind backblaze.application_id")
	}
	rootCmd.PersistentFlags().String("backblaze-application-key", "<not-set>", "backblaze application key")
	if err := viper.BindPFlag("backblaze.application_key", rootCmd.PersistentFlags().Lookup("backblaze-application-key")); err != nil {
		log.Error().Err(err).Msg("bind backblaze.application_key")
	}
	rootCmd.PersistentFlags().String("backblaze-bucket", "ticker-info", "backblaze bucket")
	if err := viper.BindPFlag("backblaze.bucket", rootCmd.PersistentFlags().Lookup("backblaze-bucket")); err != nil {
		log.Error().Err(err).Msg("bind backblaze.bucket")
	}

	rootCmd.PersistentFlags().Bool("show-browser", false, "don't run the browser in headless mode")
	if err := viper.BindPFlag("show_browser", rootCmd.PersistentFlags().Lookup("show-browser")); err != nil {
		log.Error().Err(err).Msg("bind show_browser")
	}

	rootCmd.PersistentFlags().String("parquet-file", "tickers.parquet", "save results to parquet")
	if err := viper.BindPFlag("parquet_file", rootCmd.PersistentFlags().Lookup("parquet-file")); err != nil {
		log.Error().Err(err).Msg("bind parquet_file")
	}

	rootCmd.PersistentFlags().StringP("username", "u", "", "encrypted fidelity username")
	if err := viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username")); err != nil {
		log.Error().Err(err).Msg("bind username")
	}

	rootCmd.PersistentFlags().StringP("pin", "p", "", "encrypted fidelity password")
	if err := viper.BindPFlag("pin", rootCmd.PersistentFlags().Lookup("pin")); err != nil {
		log.Error().Err(err).Msg("bind pin")
	}

	rootCmd.PersistentFlags().String("state-file", "state.json", "store session state in the speficied file")
	if err := viper.BindPFlag("state_file", rootCmd.PersistentFlags().Lookup("state-file")); err != nil {
		log.Error().Err(err).Msg("bind state_file")
	}

	rootCmd.PersistentFlags().String("user-agent", "", "user agent to use")
	if err := viper.BindPFlag("user_agent", rootCmd.PersistentFlags().Lookup("user-agent")); err != nil {
		log.Error().Err(err).Msg("bind user_agent")
	}
}

func initLog() {
	if !viper.GetBool("log.json") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if viper.GetBool("log.debug") {
		log.Level(zerolog.DebugLevel)
		log.Info().Msg("set log level to debug")
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".import-fidelity" (without extension).
		viper.AddConfigPath("/etc/") // path to look for the config file in
		viper.AddConfigPath(fmt.Sprintf("%s/.config", home))
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
		viper.SetConfigName("import-fidelity")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info().Str("ConfigFile", viper.ConfigFileUsed()).Msg("Loaded config file")
	} else {
		log.Error().Err(err).Msg("error reading config file")
	}
}
