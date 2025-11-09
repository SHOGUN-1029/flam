package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const configFile = "config.json"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `View and modify configuration parameters such as retry count and backoff base.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration parameter (e.g., max-retries, backoff-base)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])
		value := args[1]

		switch key {
		case "max-retries", "max_retries", "maxretries":
			v, err := strconv.Atoi(value)
			if err != nil {
				fmt.Println(" Invalid value for max-retries, must be integer.")
				return
			}
			config.MaxRetries = v
			if err := saveConfig(); err != nil {
				fmt.Println(" Failed to save config:", err)
			} else {
				fmt.Printf(" Updated max-retries to %d\n", config.MaxRetries)
			}

		case "backoff-base", "backoff_base", "backoffbase":
			v, err := strconv.Atoi(value)
			if err != nil {
				fmt.Println(" Invalid value for backoff-base, must be integer.")
				return
			}
			config.BackoffBase = v
			if err := saveConfig(); err != nil {
				fmt.Println(" Failed to save config:", err)
			} else {
				fmt.Printf(" Updated backoff-base to %d\n", config.BackoffBase)
			}

		default:
			fmt.Printf(" Unknown config key: %s\n", key)
		}
	},
}

// Show subcommand
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration settings",
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadConfig(); err != nil {
			fmt.Println(" Could not load config:", err)
		}
		fmt.Println("\n️ Current Configuration:")
		fmt.Printf("   Max Retries  : %d\n", config.MaxRetries)
		fmt.Printf("   Backoff Base : %d\n", config.BackoffBase)
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}

// write the current config to config.json
func saveConfig() error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}

// loadConfig reads from config.json and overwrites `config` values if present
func loadConfig() error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}
	config = c
	return nil
}
