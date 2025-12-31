package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "baluster",
	Short: "Baluster infrastructure deployment CLI",
	Long:  "A CLI tool for deploying Baluster infrastructure and services to Azure.",
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy infrastructure or services to Azure",
	Long:  "Deploy Baluster infrastructure or services to Azure.",
}

func init() {
	deployCmd.AddCommand(deployInfraCmd)
	deployCmd.AddCommand(deployReleaseCmd)
	rootCmd.AddCommand(deployCmd)
}

// loadEnvironmentFile loads the appropriate .env file based on the environment
func loadEnvironmentFile(env string) error {
	projectRoot := getProjectRoot()
	envFileName := fmt.Sprintf(".env.%s", env)
	envPath := filepath.Join(projectRoot, envFileName)

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		panic(err)
	}

	// Use Overload to override any existing environment variables (e.g., from .env.local)
	// This ensures the environment-specific file takes precedence
	if err := godotenv.Overload(envPath); err != nil {
		return fmt.Errorf("failed to load %s: %w", envPath, err)
	}

	return nil
}

func main() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	if err := rootCmd.Execute(); err != nil {
		color.New(color.FgRed).Printf("Command failed: %v", err)
		os.Exit(1)
	}
}

func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: assume we're in project root or cmd/cli
	if strings.HasSuffix(wd, "cmd/cli") {
		return filepath.Join(wd, "../..")
	}
	return wd
}
