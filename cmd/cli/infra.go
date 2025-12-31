package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/brianfromlife/baluster/internal/core/cli"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	requiredEnvVars = []string{
		"AZURE_RESOURCE_GROUP",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
		"JWT_SECRET",
		"GITHUB_REDIRECT_URL",
	}
)

var deployInfraCmd = &cobra.Command{
	Use:   "infra",
	Short: "Deploy infrastructure to Azure",
	Long:  "Deploy Baluster infrastructure to Azure using Bicep templates. Requires confirmation based on environment.",
	RunE:  runDeployInfra,
}

func runDeployInfra(cmd *cobra.Command, args []string) error {
	// Prompt for environment first
	env, err := promptEnvironment()
	if err != nil {
		if err.Error() == "cancelled" {
			color.New(color.FgYellow).Println("Deployment cancelled.")
			return nil
		}
		return fmt.Errorf("failed to get environment: %w", err)
	}

	// Load the appropriate environment file
	if err := loadEnvironmentFile(env); err != nil {
		return fmt.Errorf("failed to load environment file: %w", err)
	}

	// Validate required environment variables after loading the file
	missing := validateRequiredEnvVars()
	if len(missing) > 0 {
		color.New(color.FgRed, color.Bold).Fprintf(os.Stderr, "Error: Missing required environment variables:\n")
		for _, v := range missing {
			color.New(color.FgRed).Fprintf(os.Stderr, "  - %s\n", v)
		}
		envFileName := fmt.Sprintf(".env.%s", env)
		color.New(color.FgYellow).Fprintf(os.Stderr, "\nPlease set these variables in your %s file.\n", envFileName)
		os.Exit(1)
	}

	// Read values from environment variables
	resourceGroup := strings.TrimSpace(os.Getenv("AZURE_RESOURCE_GROUP"))
	githubClientID := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID"))
	githubClientSecret := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_SECRET"))
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	cosmosDatabase := strings.TrimSpace(os.Getenv("COSMOS_DATABASE"))
	githubRedirectURL := strings.TrimSpace(os.Getenv("GITHUB_REDIRECT_URL"))
	customDomainName := strings.TrimSpace(os.Getenv("CUSTOM_DOMAIN_NAME"))

	// Validate GITHUB_REDIRECT_URL
	if githubRedirectURL == "" {
		return fmt.Errorf("GITHUB_REDIRECT_URL environment variable is required")
	}

	// Display deployment summary
	displayDeploymentSummary(env, resourceGroup, githubClientID, cosmosDatabase)

	// Get confirmation
	confirmed, err := cli.Run(env)
	if err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}
	if !confirmed {
		color.New(color.FgYellow).Println("Deployment cancelled.")
		return nil
	}

	projectRoot := getProjectRoot()
	bicepPath := filepath.Join(projectRoot, "infra", "main.bicep")

	// Ensure resource group exists
	if err := ensureResourceGroup(resourceGroup, env); err != nil {
		return fmt.Errorf("failed to ensure resource group: %w", err)
	}

	// Deploy Bicep template
	color.New(color.FgCyan, color.Bold).Printf("\nDeploying infrastructure for environment: %s\n", env)
	return deployBicep(bicepPath, env, resourceGroup, githubClientID, githubClientSecret, jwtSecret, githubRedirectURL, customDomainName)
}

// validateRequiredEnvVars checks for required environment variables and returns list of missing ones
func validateRequiredEnvVars() []string {
	var missing []string
	for _, v := range requiredEnvVars {
		value := os.Getenv(v)
		if strings.TrimSpace(value) == "" {
			missing = append(missing, v)
		}
	}
	return missing
}

// promptEnvironment prompts the user for the deployment environment
func promptEnvironment() (string, error) {
	return cli.PromptEnvironment()
}

// displayDeploymentSummary prints a summary of deployment configuration
func displayDeploymentSummary(env, resourceGroup, githubClientID, cosmoDbName string) {
	headerColor := color.New(color.FgCyan, color.Bold)
	borderColor := color.New(color.FgCyan)
	labelColor := color.New(color.FgHiWhite)
	valueColor := color.New(color.FgWhite)
	maskedColor := color.New(color.FgYellow)

	border := strings.Repeat("=", 60)
	borderColor.Println("\n" + border)
	headerColor.Println("Deployment Configuration Summary")
	borderColor.Println(border)

	labelColor.Printf("Environment:      ")
	valueColor.Printf("%s\n", env)
	labelColor.Printf("Resource Group:   ")
	valueColor.Printf("%s\n", resourceGroup)
	labelColor.Printf("Cosmos Database:   ")
	valueColor.Printf("%s\n", cosmoDbName)

	// Mask GitHub Client ID (show first 8 chars)
	labelColor.Printf("GitHub Client ID: ")
	if len(githubClientID) > 8 {
		maskedColor.Printf("%s***\n", githubClientID[:8])
	} else {
		maskedColor.Printf("****\n")
	}

	labelColor.Printf("GitHub Secret:    ")
	maskedColor.Printf("****\n")
	labelColor.Printf("JWT Secret:       ")
	maskedColor.Printf("****\n")
	borderColor.Println(border)
}

func ensureResourceGroup(resourceGroup, environment string) error {
	args := []string{
		"group",
		"create",
		"--name", resourceGroup,
		"--location", "centralus",
		"--tags", fmt.Sprintf("environment=%s", environment), "project=baluster-dev",
	}
	color.New(color.FgCyan).Printf("Ensuring resource group exists: %s...\n", resourceGroup)
	cmd := exec.Command("az", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "already exists") {
			return fmt.Errorf("failed to create resource group: %s: %w", string(output), err)
		}
	}
	color.New(color.FgGreen).Printf("Resource group %s exists.\n", resourceGroup)
	return nil
}

func deployBicep(bicepPath string, env string, resourceGroup string, githubClientID, githubClientSecret, jwtSecret, githubRedirectURL, customDomainName string) error {
	deploymentName := fmt.Sprintf("baluster-%s-%d", env, os.Getpid())

	args := []string{
		"deployment", "group", "create",
		"--resource-group", resourceGroup,
		"--template-file", bicepPath,
		"--name", deploymentName,
		"--parameters",
		fmt.Sprintf("environment=%s", env),
		fmt.Sprintf("githubClientId=%s", githubClientID),
		fmt.Sprintf("githubClientSecret=%s", githubClientSecret),
		fmt.Sprintf("jwtSecret=%s", jwtSecret),
		fmt.Sprintf("githubRedirectUrl=%s", githubRedirectURL),
	}

	// Add custom domain parameter if provided
	if customDomainName != "" {
		args = append(args, fmt.Sprintf("customDomainName=%s", customDomainName))
	}

	args = append(args, "--output", "none")

	cmd := exec.Command("az", args...)
	// Suppress stdout but keep stderr for errors
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Printf("\n✓ Deployment completed successfully for environment: %s.\n", env)
	return nil
}
