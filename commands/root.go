//
// Copyright:: Copyright 2018 Chef Software, Inc.
// License:: Apache License, Version 2.0
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
//

package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	chef_load "github.com/lbakerchef/chef-load/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "chef-load",
	Short: "A tool for simulating loading chef data",
	Long: `A tool for simulating load on a Chef Server and/or a Chef Automate Server.
Complete documentation is available at https://github.com/lbakerchef/chef-load`,
	TraverseChildren: true,
	Run:              func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.chef-load.toml)")
	rootCmd.PersistentFlags().StringP("data_collector_url", "d", "", "The data-collector url")
	// TODO: add the token flag for the data collector
	rootCmd.PersistentFlags().StringP("chef_server_url", "s", "", "The chef-server url")
	rootCmd.PersistentFlags().StringP("node_name_prefix", "p", "chef-load", "The nodes name prefix")
	rootCmd.PersistentFlags().IntP("num_nodes", "n", 0, "The number of nodes to simulate")
	rootCmd.PersistentFlags().IntP("num_actions", "a", 0, "The number of actions to generate")
	rootCmd.PersistentFlags().IntP("matrix.simulation.nodes", "N", 0, "The number of inspec nodes to generate")
	rootCmd.PersistentFlags().IntP("matrix.simulation.days", "D", 0, "The number of days worth of inspec nodes to generate")
	rootCmd.PersistentFlags().IntP("matrix.simulation.max_scans", "M", 0, "The number of max scans per day for inspec nodes to generate")
	rootCmd.PersistentFlags().IntP("matrix.simulation.total_max_scans", "T", 0, "The number of max scans per day for inspec nodes to generate")
	rootCmd.PersistentFlags().StringP("matrix.simulation.format", "F", "full", "Format of incoming inspec nodes to generate")
	rootCmd.PersistentFlags().BoolP("random_data", "r", false, "Generates random data")
	rootCmd.PersistentFlags().BoolP("liveness_agent", "l", false, "Generates liveness agent data")
	rootCmd.PersistentFlags().IntP("interval", "i", 30, "Interval between a node's chef-client runs, in minutes")
	// TODO unless these apply globally (they don't yet) they should be added to the "start" command.
	rootCmd.PersistentFlags().Float64P("download_cookbooks_scale_factor", "C", 1.0, "What probability (0.0 - 1.0) that any given cookbook will need to be downloaded")
	rootCmd.PersistentFlags().BoolP("skip_client_creation", "S", false, "Skips creation of client during each node's initial chef-client run")
	rootCmd.PersistentFlags().Float64P("node_replacement_rate", "R", 0.0, "How frequently (0.0 - 1.0) are new nodes generated and old ones no longer run. Default 0.0")
	viper.BindPFlags(rootCmd.PersistentFlags())
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Unable read config:", err)
			os.Exit(1)
		}
	} else {
		// TODO: @afiune if the user doesn't provide a config file
		// should we load it from the home directory or somewhere else
		// also, should we instead as them to run `chef-load init`?
		viper.SetConfigName(".chef-load")
		viper.SetConfigType("toml")
		viper.AddConfigPath("$HOME")
	}
}

func configFromViper() (*chef_load.Config, error) {
	cfg := chef_load.Default()
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.ChefServerURL == "" && cfg.DataCollectorURL == "" {
		return nil, errors.New("You must set chef_server_url or data_collector_url or both")
	}

	if cfg.ChefServerURL != "" {
		cfg.RunChefClient = true
		if !strings.HasSuffix(cfg.ChefServerURL, "/") {
			cfg.ChefServerURL = cfg.ChefServerURL + "/"
		}
		if cfg.ClientName == "" || cfg.ClientKey == "" {
			return nil, errors.New("You must set client_name and client_key if chef_server_url is set")
		}
	}

	if cfg.DataCollectorURL != "" && cfg.ChefServerURL == "" {
		// make sure cfg.ChefServerURL is set to something because it is used
		// even when only in data-collector mode
		cfg.ChefServerURL = "https://chef.example.com/organizations/demo/"
	}

	return &cfg, nil
}
