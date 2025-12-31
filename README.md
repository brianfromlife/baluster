# Baluster

Baluster is a demo project that provides an access control and API key management system. It enables organizations to manage applications, API keys, and service keys with role-based access control. The system includes GitHub OAuth authentication, REST and gRPC APIs, and a React-based web interface. Data is stored in Azure Cosmos DB, and the entire infrastructure can be deployed to Azure using IaC with Bicep templates.

## Application Features

Baluster provides a comprehensive access control and API key management system with the following features:

- **Organization Management**: Create and manage organizations with role-based access control
- **Application Management**: Define applications with custom permission sets within organizations
- **API Key Management**: Generate and manage API keys for authenticating requests to Baluster APIs using Bearer token authentication
- **Service Key Management**: Create service keys that grant access to multiple applications with granular permissions
- **Access Validation**: Validate service key permissions via REST and gRPC APIs for microservice-to-microservice communication
- **Audit History**: Complete audit trail tracking all create, update, and delete operations on applications, API keys, and service keys, including user information and timestamps
- **GitHub OAuth Authentication**: User authentication via GitHub OAuth with JWT-based session management
- **Dual API Support**: Both REST and gRPC (Connect RPC) interfaces available for programmatic access

## CLI Demo

The cli tool in this project will deploy your infrastructure and application.

The CLI will:

- Prompts you to select an environment (prod or dev) and enter a confirmation code
- Deploys both backend services (REST and gRPC), and the Cosmos database and containers
- Builds the React application and deploys it to an Azure Static Web App

  ![Baluster CLI Demo](docs/baluster-cli-demo.webp)

## Environment Setup

To run and deploy Baluster locally, you'll need the following tools installed:

### Required Tools

- **Go 1.25+**

  - Download from [golang.org](https://golang.org/dl/)
  - Verify installation: `go version`

- **Node.js 20+**

  - Download from [nodejs.org](https://nodejs.org/)
  - Verify installation: `node --version`

- **pnpm**

  - Install: `npm install -g pnpm`
  - Verify installation: `pnpm --version`

- **Make** - Build automation tool (`Makefile`)

  - Make is typically pre-installed on macOS and Linux
  - Verify installation: `make --version`
  - On macOS, if not installed: `xcode-select --install`

- **Azure CLI** - Required for deploying infrastructure and managing Azure resources

  - Install: Follow instructions at [docs.microsoft.com/cli/azure/install-azure-cli](https://docs.microsoft.com/cli/azure/install-azure-cli)
  - Verify installation: `az --version`
  - Login: `az login`

- **buf CLI** - Protocol buffer compiler and code generator

  - The Makefile includes an `install-buf` target that will install buf automatically
  - Or install manually: Download from [buf.build](https://buf.build/docs/installation)
  - Verify installation: `buf --version`

- **Docker** - Required for building and pushing container images
  - Download from [docker.com](https://www.docker.com/get-started)
  - Verify installation: `docker --version`

### Before Getting Started

Before you can deploy Baluster to Azure, you'll need to set up an Azure subscription and authenticate with the Azure CLI:

1. **Create an Azure Subscription:**

   - Go to the [Azure Portal](https://portal.azure.com/)
   - Sign in with your Microsoft account
   - Navigate to Subscriptions → Create subscription
   - Follow the prompts to create a new subscription (you may be eligible for a free trial)
   - Note your subscription name and ID for reference

2. **Log in to Azure CLI:**

   ```bash
   az login
   ```

   This will open a browser window where you can authenticate with your Azure account.

3. **Set your active subscription:**

   If you have multiple subscriptions, set the active subscription you want to use:

   ```bash
   az account set --subscription "<subscription-name-or-id>"
   ```

   Verify your active subscription:

   ```bash
   az account list --output table
   ```

   Or view details of the current subscription:

   ```bash
   az account show
   ```

## Running Locally

To run Baluster locally, you'll need to deploy the development infrastructure to Azure first, then configure your local environment with the necessary connection details.

### Required Environment Variables

For local development, you'll need the following environment variables:

- `COSMOS_ENDPOINT` - Azure Cosmos DB account endpoint URL
- `COSMOS_KEY` - Azure Cosmos DB account primary key
- `COSMOS_DATABASE` - Cosmos DB database name (typically "baluster")
- `JWT_SECRET` - Secret key for JWT token signing (use a strong random string)
- `GITHUB_CLIENT_ID` - GitHub OAuth application client ID
- `GITHUB_CLIENT_SECRET` - GitHub OAuth application client secret
- `GITHUB_REDIRECT_URL` - OAuth callback URL (optional, falls back on `http://localhost:5173/auth/callback`)

### Deploying Development Resources

Before running locally, you need to deploy the development infrastructure to Azure:

1. **Set up a GitHub OAuth Application:**

   - Go to GitHub Settings → Developer settings → OAuth Apps
   - Create a new OAuth App
   - Set Authorization callback URL to `http://localhost:5173/auth/callback`
   - Copy the Client ID and Client Secret to your `.env.local` file (next step)

2. **Create a `.env.dev` file** in the project root with your deployment configuration:

   ```bash
   AZURE_RESOURCE_GROUP=baluster-rg-dev
   COSMOS_DATABASE=baluster-dev
   GITHUB_CLIENT_ID=<your-github-client-id>
   GITHUB_CLIENT_SECRET=<your-github-client-secret>
   GITHUB_REDIRECT_URL=http://localhost:5173/auth/callback
   JWT_SECRET=<your-jwt-secret>
   ```

   > **Note:** For local development, use `http://localhost:5173/auth/callback` as the redirect URL. If you're deploying to a dev environment in Azure, you'll need to update this with the actual Static Web App URL after the first deployment (similar to production).

3. **Build the CLI tool:**

   ```bash
   make build-cli
   ```

4. **Deploy development infrastructure:**

   ```bash
   ./bin/cli deploy infra # or `make infra`
   ```

   - Select `dev` as the environment
   - Type `confirm` when prompted

5. **Update environment variables:**

   In the Azure portal, navigate to your new Cosmos database and grab the `URI` and `PRIMARY KEY` from Settings > Keys.
   Alternatively you can grab the information from the command line:

   ```bash
   # Get the Cosmos DB account name
   COSMOS_ACCOUNT=$(az cosmosdb list --resource-group baluster-rg-dev --query "[?contains(name, 'dev')].name" -o tsv)

   # Get the Cosmos DB endpoint
   COSMOS_ENDPOINT=$(az cosmosdb show --resource-group baluster-rg-dev --name $COSMOS_ACCOUNT --query documentEndpoint -o tsv)

   # Get the Cosmos DB primary key
   COSMOS_KEY=$(az cosmosdb keys list --resource-group baluster-rg-dev --name $COSMOS_ACCOUNT --query primaryMasterKey -o tsv)

   echo "COSMOS_ENDPOINT=$COSMOS_ENDPOINT"
   echo "COSMOS_KEY=$COSMOS_KEY"
   ```

   New environment variables:

   ```bash
    COSMOS_ENDPOINT="<add-here>"
    COSMOS_KEY="<add-here>"
   ```

### Running the Services

Once your environment variables are set, you can run the services locally:

1. **Install backend dependencies:**

   ```bash
   go mod download
   ```

2. **Install frontend dependencies:**

   ```bash
   cd web && pnpm install
   ```

3. **Build the applications**

   ```bash
   make build # this will generate the proto files and build all of the executables

   ```

4. **Run the REST server** (in one terminal):

   ```bash
   make rest
   ```

   The REST server will start on `http://localhost:8080`

5. **Run the gRPC server** (in another terminal, optional):

   ```bash
   make grpc
   ```

6. **Run the frontend development server** (in another terminal):
   ```bash
   make web
   ```
   The frontend will start on `http://localhost:5173`

You can now access the application at `http://localhost:5173` and authenticate using GitHub OAuth.

## Deploying

Baluster includes a CLI tool that simplifies deployment to Azure. The CLI provides two main deployment commands: infrastructure deployment and service deployment.

### Building the CLI

First, build the CLI tool:

```bash
make build-cli
```

This will create a binary at `bin/cli`.

### CLI Commands

The CLI tool provides the following commands:

#### `deploy infra`

Deploys the Azure infrastructure using Bicep templates. This command:

- Prompts you to select an environment (prod or dev)
- Validates required environment variables
- Displays a deployment summary
- Requires confirmation (production requires typing a random 12-character code, development requires typing "confirm")
- Creates or ensures the Azure resource group exists
- Deploys the Bicep template with the specified parameters

**Required Environment Variables:**

- `AZURE_RESOURCE_GROUP` - Azure resource group name
- `GITHUB_CLIENT_ID` - GitHub OAuth application client ID
- `GITHUB_CLIENT_SECRET` - GitHub OAuth application client secret
- `JWT_SECRET` - Secret key for JWT token signing

**Usage:**

```bash
make infra
```

The CLI reads environment variables from a `.env.[environment]` file in the project root. Create this file with your configuration:

```bash
AZURE_RESOURCE_GROUP=baluster-rg
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
JWT_SECRET=your-jwt-secret
COSMOS_DATABASE=baluster
```

#### `deploy release`

Builds and deploys all applications (backend services and web application). This command:

**Required Environment Variables:**

- `AZURE_RESOURCE_GROUP` - Azure resource group name (must contain a container registry and Static Web App)

**Usage:**

```bash
./bin/cli deploy release # or `make release`
```

### Deployment Workflow

A typical deployment workflow would be:

1. **Deploy Infrastructure** (first time or when infrastructure changes):

   ```bash
   ./bin/cli deploy infra
   ```

   Select your environment and confirm the deployment. This creates all Azure resources including Cosmos DB, Container Registry, and App Services.

2. **Deploy Applications** (when code changes):

   ```bash
   ./bin/cli deploy release
   ```

   Or using Make:

   ```bash
   make release
   ```

   Select your environment and confirm. This builds and deploys both backend services (Docker images) and the web application (React app) to Azure.

### Deploying to Production

**Before you begin:** Create a `.env.prod` file in the project root with all required environment variables:

```bash
AZURE_RESOURCE_GROUP=baluster-rg-prod
GITHUB_CLIENT_ID=<placeholder-initially>
GITHUB_CLIENT_SECRET=<placeholder-initially>
GITHUB_REDIRECT_URL=<placeholder-initially>
JWT_SECRET=<your-jwt-secret>
COSMOS_DATABASE=baluster-prod
```

> **Note:** You can use placeholder values for `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` initially. You'll update these with real values after the first deployment (see step 3 below).
>
> **Important:** After the first deployment, you'll also need to update the `GITHUB_REDIRECT_URL` environment variable in your `.env.prod` file with the actual Static Web App URL (e.g., `https://your-app.azurestaticapps.net/auth/callback`). This URL is only known after the first deployment creates the Static Web App.
>
> **Using a Custom Domain:** If you already have a custom domain registered (e.g., `myapp.com`), you can set `GITHUB_REDIRECT_URL=https://myapp.com/auth/callback` in your `.env.prod` file from the start. This eliminates the need to update it after the first deployment, since you know your domain name beforehand. You'll still need to configure your Static Web App to use the custom domain after deployment, but the OAuth configuration can be set up correctly from the beginning.

When deploying to production for the first time, you'll need to set up GitHub OAuth in a two-step process:

1. **Initial Deployment** (GitHub OAuth will fail):

   First, deploy the application with placeholder GitHub OAuth credentials:

   ```bash
   make release
   ```

   Select `prod` as the environment and confirm. After deployment completes, the CLI will display the Static Web App URL (e.g., `https://your-app.azurestaticapps.net`).

2. **Configure GitHub OAuth Application**:

   - Go to GitHub Settings → Developer settings → OAuth Apps
   - Create a new OAuth App for production
   - Set the Authorization callback URL to: `https://<your-static-web-app-url>/auth/callback`
     - Replace `<your-static-web-app-url>` with the URL displayed after deployment
   - Copy the Client ID and Client Secret

3. **Update Environment Variables**:

   Update your `.env.prod` file with the production GitHub OAuth credentials and redirect URL:

   ```bash
   GITHUB_CLIENT_ID=<your-production-github-client-id>
   GITHUB_CLIENT_SECRET=<your-production-github-client-secret>
   GITHUB_REDIRECT_URL=https://<your-static-web-app-url>/auth/callback
   ```

   Replace `<your-static-web-app-url>` with the actual Static Web App URL from step 1.

   > **Note:** The `GITHUB_REDIRECT_URL` must be updated after the first deployment because the Static Web App URL is only known after the infrastructure is created. This URL will be used in subsequent infrastructure deployments via the Bicep template.

4. **Redeploy Applications**:

   After updating the environment variables, redeploy the applications:

   ```bash
   make release
   ```

   Select `prod` as the environment and confirm. GitHub OAuth should now work correctly with the production URL.
