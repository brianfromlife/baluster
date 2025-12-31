package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/brianfromlife/baluster/internal/core/cli"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var deployReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Build and deploy all applications (services and web)",
	Long:  "Build and push Docker images for REST and gRPC services, and deploy the web application to Azure Static Web Apps.",
	RunE:  runDeployApplication,
}

func runDeployApplication(cmd *cobra.Command, args []string) error {
	// Prompt for environment first
	env, err := cli.PromptEnvironment()
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

	// Read values from environment variables
	resourceGroup := strings.TrimSpace(os.Getenv("AZURE_RESOURCE_GROUP"))
	if resourceGroup == "" {
		return fmt.Errorf("AZURE_RESOURCE_GROUP environment variable is not set")
	}

	// Get registry server and static web app name for display in summary
	registryServer, err := getContainerRegistryLoginServer(resourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get container registry login server: %w", err)
	}

	staticWebAppName := fmt.Sprintf("baluster-web-%s", env)

	// deployment summary
	displayReleaseDeploymentSummary(env, resourceGroup, registryServer, staticWebAppName)

	// Run bubbletea workflow
	confirmed, err := cli.Run(env)
	if err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}
	if !confirmed {
		color.New(color.FgYellow).Println("Deployment cancelled.")
		return nil
	}

	projectRoot := getProjectRoot()

	// Deploy backend services
	border := strings.Repeat("=", 60)
	color.New(color.FgCyan).Println("\n" + border)
	color.New(color.FgCyan, color.Bold).Println("Deploying Backend Services")
	color.New(color.FgCyan).Println(border)

	if err := loginToACR(registryServer); err != nil {
		return fmt.Errorf("failed to login to container registry: %w", err)
	}

	if err := buildAndPushImage(projectRoot, "build/rest.dockerfile", fmt.Sprintf("%s/baluster-rest:latest", registryServer)); err != nil {
		return fmt.Errorf("failed to build and push REST image: %w", err)
	}

	if err := buildAndPushImage(projectRoot, "build/grpc.dockerfile", fmt.Sprintf("%s/baluster-grpc:latest", registryServer)); err != nil {
		return fmt.Errorf("failed to build and push gRPC image: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Println("\n✓ Backend services deployed successfully.")

	// Deploy web
	color.New(color.FgCyan).Println("\n" + border)
	color.New(color.FgCyan, color.Bold).Println("Deploying Web Application")
	color.New(color.FgCyan).Println(border)

	// Get REST API URL for the vite envvar
	restApiUrl, err := getRestApiUrl(resourceGroup, env)
	if err != nil {
		return fmt.Errorf("failed to get REST API URL: %w", err)
	}

	// Build the web application
	if err := buildWebApplication(projectRoot, restApiUrl); err != nil {
		return fmt.Errorf("failed to build web application: %w", err)
	}

	if err := verifyStaticWebAppExists(staticWebAppName, resourceGroup); err != nil {
		return fmt.Errorf("static web app verification failed: %w", err)
	}

	// Deploy the application
	deploymentToken, err := getStaticWebAppDeploymentToken(staticWebAppName, resourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get deployment token: %w", err)
	}

	webDir := filepath.Join(projectRoot, "web")
	if err := deployToStaticWebApp(webDir, deploymentToken); err != nil {
		return fmt.Errorf("failed to deploy to static web app: %w", err)
	}

	// Display the URL
	url, err := getStaticWebAppURL(staticWebAppName, resourceGroup)
	if err != nil {
		color.New(color.FgYellow).Printf("Warning: failed to retrieve Static Web App URL: %v\n", err)
	} else {
		color.New(color.FgCyan).Printf("\nStatic Web App URL: ")
		color.New(color.FgHiCyan, color.Bold).Printf("https://%s\n", url)
	}

	color.New(color.FgGreen, color.Bold).Println("\n✓ Web application deployed successfully.")
	finalBorder := strings.Repeat("=", 60)
	color.New(color.FgGreen).Println("\n" + finalBorder)
	color.New(color.FgGreen, color.Bold).Println("✓ All deployments completed successfully!")
	color.New(color.FgGreen).Println(finalBorder)
	return nil
}

func displayReleaseDeploymentSummary(env, resourceGroup, registryServer, staticWebAppName string) {
	headerColor := color.New(color.FgCyan, color.Bold)
	borderColor := color.New(color.FgCyan)
	labelColor := color.New(color.FgHiWhite)
	valueColor := color.New(color.FgWhite)

	border := strings.Repeat("=", 60)
	borderColor.Println("\n" + border)
	headerColor.Println("Application Deployment Configuration Summary")
	borderColor.Println(border)

	labelColor.Printf("Environment:      ")
	valueColor.Printf("%s\n", env)
	labelColor.Printf("Resource Group:   ")
	valueColor.Printf("%s\n", resourceGroup)
	labelColor.Printf("Registry Server:  ")
	valueColor.Printf("%s\n", registryServer)
	labelColor.Printf("REST Image:       ")
	valueColor.Printf("%s/baluster-rest:latest\n", registryServer)
	labelColor.Printf("gRPC Image:       ")
	valueColor.Printf("%s/baluster-grpc:latest\n", registryServer)
	labelColor.Printf("Static Web App:   ")
	valueColor.Printf("%s\n", staticWebAppName)
	borderColor.Println(border)
}

func buildWebApplication(projectRoot string, restApiUrl string) error {
	color.New(color.FgCyan).Println("\nBuilding web application...")
	webDir := filepath.Join(projectRoot, "web")

	// Check if pnpm is available
	if _, err := exec.LookPath("pnpm"); err != nil {
		return fmt.Errorf("pnpm is not installed - install it with: npm install -g pnpm")
	}

	// Run pnpm install
	color.New(color.FgCyan).Println("Installing dependencies...")
	installCmd := exec.Command("pnpm", "install")
	installCmd.Dir = webDir
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	color.New(color.FgCyan).Printf("Building application with VITE_API_URL=%s...\n", restApiUrl)

	// Run pnpm build with VITE_API_URL environment variable
	buildCmd := exec.Command("pnpm", "build")
	buildCmd.Dir = webDir
	buildCmd.Env = append(os.Environ(), fmt.Sprintf("VITE_API_URL=%s", restApiUrl))
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build application: %w", err)
	}

	distDir := filepath.Join(webDir, "dist")
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		return fmt.Errorf("build failed - dist directory not found")
	}

	color.New(color.FgGreen).Println("✓ Build completed successfully.")
	return nil
}

func verifyStaticWebAppExists(staticWebAppName, resourceGroup string) error {
	color.New(color.FgCyan).Printf("\nVerifying Static Web App exists: %s\n", staticWebAppName)
	args := []string{
		"staticwebapp", "show",
		"--name", staticWebAppName,
		"--resource-group", resourceGroup,
		"--output", "none",
	}

	cmd := exec.Command("az", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("static web app '%s' not found in resource group '%s' - make sure you've deployed infrastructure first", staticWebAppName, resourceGroup)
	}

	color.New(color.FgGreen).Println("✓ Static Web App verified.")
	return nil
}

func getStaticWebAppDeploymentToken(staticWebAppName, resourceGroup string) (string, error) {
	color.New(color.FgCyan).Println("Retrieving deployment token...")
	args := []string{
		"staticwebapp", "secrets", "list",
		"--name", staticWebAppName,
		"--resource-group", resourceGroup,
		"--query", "properties.apiKey",
		"--output", "tsv",
	}

	cmd := exec.Command("az", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve deployment token: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("deployment token is empty")
	}

	color.New(color.FgGreen).Println("✓ Deployment token retrieved.")
	return token, nil
}

func deployToStaticWebApp(webDir, deploymentToken string) error {
	color.New(color.FgCyan).Println("\nDeploying to Azure Static Web Apps...")

	var cmd *exec.Cmd
	// Check if swa CLI is available, otherwise use npx
	if _, err := exec.LookPath("swa"); err == nil {
		// Use swa CLI directly
		args := []string{"deploy", "dist", "--deployment-token", deploymentToken, "--env", "production"}
		cmd = exec.Command("swa", args...)
	} else {
		// Use npx to run @azure/static-web-apps-cli
		args := []string{"@azure/static-web-apps-cli", "deploy", "dist", "--deployment-token", deploymentToken, "--env", "production"}
		cmd = exec.Command("npx", args...)
	}

	cmd.Dir = webDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	return nil
}

func getStaticWebAppURL(staticWebAppName, resourceGroup string) (string, error) {
	args := []string{
		"staticwebapp", "show",
		"--name", staticWebAppName,
		"--resource-group", resourceGroup,
		"--query", "defaultHostname",
		"--output", "tsv",
	}

	cmd := exec.Command("az", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get static web app URL: %w", err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("static web app URL is empty")
	}

	return url, nil
}

func getContainerRegistryLoginServer(resourceGroup string) (string, error) {
	args := []string{
		"acr", "list",
		"--resource-group", resourceGroup,
		"--query", "[0].loginServer",
		"--output", "tsv",
	}

	cmd := exec.Command("az", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list container registries: %w", err)
	}

	server := strings.TrimSpace(string(output))
	if server == "" {
		return "", fmt.Errorf("no container registry found in resource group: %s", resourceGroup)
	}

	return server, nil
}

func getRestApiUrl(resourceGroup string, env string) (string, error) {
	restAppName := fmt.Sprintf("baluster-rest-%s", env)
	color.New(color.FgCyan).Printf("Retrieving REST API URL for %s...\n", restAppName)

	args := []string{
		"containerapp", "show",
		"--name", restAppName,
		"--resource-group", resourceGroup,
		"--query", "properties.configuration.ingress.fqdn",
		"--output", "tsv",
	}

	cmd := exec.Command("az", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get REST API URL: %w", err)
	}

	fqdn := strings.TrimSpace(string(output))
	if fqdn == "" {
		return "", fmt.Errorf("REST API FQDN is empty for container app: %s", restAppName)
	}

	// Prepend https:// to get the full URL
	restApiUrl := fmt.Sprintf("https://%s", fqdn)
	color.New(color.FgGreen).Printf("✓ REST API URL: %s\n", restApiUrl)

	return restApiUrl, nil
}

func loginToACR(registryServer string) error {
	color.New(color.FgCyan).Printf("Logging in to container registry: %s\n", registryServer)

	// Extract registry name from login server (e.g., "myregistry.azurecr.io" -> "myregistry")
	registryName := registryServer
	if strings.Contains(registryServer, ".") {
		registryName = strings.Split(registryServer, ".")[0]
	}

	args := []string{
		"acr", "login",
		"--name", registryName,
	}

	cmd := exec.Command("az", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to login to ACR: %w", err)
	}

	return nil
}

func buildAndPushImage(projectRoot, dockerfile, imageTag string) error {
	color.New(color.FgCyan).Printf("\nBuilding Docker image: %s\n", imageTag)
	buildArgs := []string{
		"build",
		"-t", imageTag,
		"-f", dockerfile,
		projectRoot,
	}

	// Build docker image
	buildCmd := exec.Command("docker", buildArgs...)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	color.New(color.FgGreen).Printf("✓ Image built successfully: %s\n", imageTag)

	// Push docker image
	color.New(color.FgCyan).Printf("Pushing Docker image: %s\n", imageTag)
	pushArgs := []string{
		"push",
		imageTag,
	}

	pushCmd := exec.Command("docker", pushArgs...)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr

	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push Docker image: %w", err)
	}

	color.New(color.FgGreen).Printf("✓ Image pushed successfully: %s\n", imageTag)
	return nil
}
