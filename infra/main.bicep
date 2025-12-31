@description('Name of the resource group')
param resourceGroupName string = 'baluster-rg'

@description('Azure region')
param location string = 'centralus'

@description('Environment name (prod or dev)')
@allowed(['prod', 'dev'])
param environment string = 'prod'

@description('Name of the Cosmos DB account')
param cosmosDbAccountName string = ''

@description('Name of the Cosmos DB database')
param cosmosDbDatabaseName string = ''

@description('Name of the Container Apps environment')
param containerAppEnvironmentName string = ''

@description('Name of the REST API container app')
param restAppName string = ''

@description('Name of the gRPC API container app')
param grpcAppName string = ''

@description('Name of the Static Web App')
param staticWebAppName string = ''

@description('GitHub OAuth Client ID')
@secure()
param githubClientId string

@description('GitHub OAuth Client Secret')
@secure()
param githubClientSecret string

@description('JWT secret for token signing')
@secure()
param jwtSecret string

@description('GitHub OAuth redirect URL')
param githubRedirectUrl string

@description('Custom domain name for the Static Web App (optional, e.g., app.example.com)')
param customDomainName string = ''

var tags = {
  environment: environment
  project: 'baluster'
}

var acrName = replace(resourceGroupName, '-', '')

// Resource names with environment suffix
var cosmosDbAccountNameFinal = !empty(cosmosDbAccountName) ? cosmosDbAccountName : 'baluster-cosmos-${environment}'
var cosmosDbDatabaseNameFinal = !empty(cosmosDbDatabaseName) ? cosmosDbDatabaseName : 'baluster-${environment}'
var containerAppEnvironmentNameFinal = !empty(containerAppEnvironmentName)
  ? containerAppEnvironmentName
  : 'baluster-${environment}-app'
var restAppNameFinal = !empty(restAppName) ? restAppName : 'baluster-rest-${environment}'
var grpcAppNameFinal = !empty(grpcAppName) ? grpcAppName : 'baluster-grpc-${environment}'
var staticWebAppNameFinal = !empty(staticWebAppName) ? staticWebAppName : 'baluster-web-${environment}'

// Cosmos DB Account
resource cosmosAccount 'Microsoft.DocumentDB/databaseAccounts@2023-09-15' = {
  name: cosmosDbAccountNameFinal
  location: location
  kind: 'GlobalDocumentDB'
  properties: {
    consistencyPolicy: {
      defaultConsistencyLevel: 'Session'
    }
    locations: [
      {
        locationName: location
        failoverPriority: 0
        isZoneRedundant: false
      }
    ]
    databaseAccountOfferType: 'Standard'
    capabilities: [
      {
        name: 'EnableServerless'
      }
    ]
    enableAutomaticFailover: false
    enableMultipleWriteLocations: false
  }
  tags: tags
}

// Cosmos DB Database
resource cosmosDatabase 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases@2023-09-15' = {
  parent: cosmosAccount
  name: cosmosDbDatabaseNameFinal
  properties: {
    resource: {
      id: cosmosDbDatabaseNameFinal
    }
  }
}

// Cosmos DB Container - Organizations (also contains organization_members)
resource organizationsContainer 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers@2023-09-15' = {
  parent: cosmosDatabase
  name: 'organizations'
  properties: {
    resource: {
      id: 'organizations'
      partitionKey: {
        paths: [
          '/organization_id'
        ]
        kind: 'Hash'
      }
      indexingPolicy: {
        indexingMode: 'consistent'
        automatic: true
        compositeIndexes: [
          [
            {
              path: '/entity_type'
              order: 'ascending'
            }
            {
              path: '/id'
              order: 'ascending'
            }
          ]
        ]
      }
    }
  }
}

// Cosmos DB Container - Applications
resource applicationsContainer 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers@2023-09-15' = {
  parent: cosmosDatabase
  name: 'applications'
  properties: {
    resource: {
      id: 'applications'
      partitionKey: {
        paths: [
          '/organization_id'
        ]
        kind: 'Hash'
      }
      indexingPolicy: {
        indexingMode: 'consistent'
        automatic: true
        compositeIndexes: [
          [
            {
              path: '/entity_type'
              order: 'ascending'
            }
            {
              path: '/entity_id'
              order: 'ascending'
            }
            {
              path: '/created_at'
              order: 'descending'
            }
          ]
        ]
      }
    }
  }
}

// Cosmos DB Container - Service Keys
resource serviceKeysContainer 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers@2023-09-15' = {
  parent: cosmosDatabase
  name: 'service_keys'
  properties: {
    resource: {
      id: 'service_keys'
      partitionKey: {
        paths: [
          '/organization_id'
        ]
        kind: 'Hash'
      }
      indexingPolicy: {
        indexingMode: 'consistent'
        automatic: true
        compositeIndexes: [
          [
            {
              path: '/entity_type'
              order: 'ascending'
            }
            {
              path: '/entity_id'
              order: 'ascending'
            }
            {
              path: '/created_at'
              order: 'descending'
            }
          ]
        ]
      }
    }
  }
}

// Cosmos DB Container - API Keys
resource apiKeysContainer 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers@2023-09-15' = {
  parent: cosmosDatabase
  name: 'api_keys'
  properties: {
    resource: {
      id: 'api_keys'
      partitionKey: {
        paths: [
          '/organization_id'
        ]
        kind: 'Hash'
      }
      indexingPolicy: {
        indexingMode: 'consistent'
        automatic: true
      }
    }
  }
}

// Cosmos DB Container - Users
resource usersContainer 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers@2023-09-15' = {
  parent: cosmosDatabase
  name: 'users'
  properties: {
    resource: {
      id: 'users'
      partitionKey: {
        paths: [
          '/github_id'
        ]
        kind: 'Hash'
      }
      indexingPolicy: {
        indexingMode: 'consistent'
        automatic: true
        compositeIndexes: [
          [
            {
              path: '/entity_type'
              order: 'ascending'
            }
            {
              path: '/entity_id'
              order: 'ascending'
            }
            {
              path: '/created_at'
              order: 'descending'
            }
          ]
        ]
      }
    }
  }
}

// Log Analytics Workspace
resource logAnalyticsWorkspace 'Microsoft.OperationalInsights/workspaces@2023-09-01' = {
  name: '${resourceGroupName}-logs-${environment}'
  location: location
  properties: {
    sku: {
      name: 'PerGB2018'
    }
    retentionInDays: 30
  }
  tags: tags
}

// Container Apps Environment
resource containerAppEnvironment 'Microsoft.App/managedEnvironments@2023-05-01' = {
  name: containerAppEnvironmentNameFinal
  location: location
  properties: {
    appLogsConfiguration: {
      destination: 'log-analytics'
      logAnalyticsConfiguration: {
        customerId: logAnalyticsWorkspace.properties.customerId
        sharedKey: logAnalyticsWorkspace.listKeys().primarySharedKey
      }
    }
  }
  tags: tags
}

// Container Registry
resource containerRegistry 'Microsoft.ContainerRegistry/registries@2023-07-01' = {
  name: '${acrName}registry${environment}'
  location: location
  sku: {
    name: 'Basic'
  }
  properties: {
    adminUserEnabled: true
  }
  tags: tags
}

// User Assigned Identity for REST API
resource restIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: '${restAppNameFinal}-identity'
  location: location
  tags: tags
}

// User Assigned Identity for gRPC API
resource grpcIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: '${grpcAppNameFinal}-identity'
  location: location
  tags: tags
}

// Role Assignment - REST ACR Pull
resource restAcrRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(containerRegistry.id, restIdentity.id, 'AcrPull')
  scope: containerRegistry
  properties: {
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions',
      '7f951dda-4ed3-4680-a7ca-43fe172d538d'
    ) // AcrPull
    principalId: restIdentity.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

// Role Assignment - gRPC ACR Pull
resource grpcAcrRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(containerRegistry.id, grpcIdentity.id, 'AcrPull')
  scope: containerRegistry
  properties: {
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions',
      '7f951dda-4ed3-4680-a7ca-43fe172d538d'
    ) // AcrPull
    principalId: grpcIdentity.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

// REST API Container App
resource restContainerApp 'Microsoft.App/containerApps@2023-05-01' = {
  name: restAppNameFinal
  location: location
  properties: {
    managedEnvironmentId: containerAppEnvironment.id
    configuration: {
      ingress: {
        external: true
        targetPort: 8080
        transport: 'http'
        allowInsecure: false
        corsPolicy: {
          allowCredentials: true
          allowedOrigins: ['*']
          allowedMethods: ['*']
          allowedHeaders: ['*']
          maxAge: 3600
        }
      }
      registries: [
        {
          server: containerRegistry.properties.loginServer
          identity: restIdentity.id
        }
      ]
      secrets: [
        {
          name: 'cosmos-key'
          value: cosmosAccount.listKeys().primaryMasterKey
        }
        {
          name: 'jwt-secret'
          value: jwtSecret
        }
        {
          name: 'github-client-id'
          value: githubClientId
        }
        {
          name: 'github-client-secret'
          value: githubClientSecret
        }
      ]
    }
    template: {
      containers: [
        {
          name: 'rest-api'
          image: '${containerRegistry.properties.loginServer}/baluster-rest:latest'
          resources: {
            cpu: json('0.5')
            memory: '1Gi'
          }
          env: [
            {
              name: 'PORT'
              value: '8080'
            }
            {
              name: 'COSMOS_ENDPOINT'
              value: cosmosAccount.properties.documentEndpoint
            }
            {
              name: 'COSMOS_KEY'
              secretRef: 'cosmos-key'
            }
            {
              name: 'COSMOS_DATABASE'
              value: cosmosDbDatabaseNameFinal
            }
            {
              name: 'JWT_SECRET'
              secretRef: 'jwt-secret'
            }
            {
              name: 'GITHUB_CLIENT_ID'
              secretRef: 'github-client-id'
            }
            {
              name: 'GITHUB_CLIENT_SECRET'
              secretRef: 'github-client-secret'
            }
            {
              name: 'GITHUB_REDIRECT_URL'
              value: githubRedirectUrl
            }
          ]
        }
      ]
      scale: {
        minReplicas: 1
        maxReplicas: 3
      }
    }
  }
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${restIdentity.id}': {}
    }
  }
  tags: tags
  dependsOn: [
    restAcrRoleAssignment
  ]
}

// gRPC API Container App
resource grpcContainerApp 'Microsoft.App/containerApps@2023-05-01' = {
  name: grpcAppNameFinal
  location: location
  properties: {
    managedEnvironmentId: containerAppEnvironment.id
    configuration: {
      ingress: {
        external: true
        targetPort: 5050
        transport: 'http'
        allowInsecure: false
        corsPolicy: {
          allowCredentials: true
          allowedOrigins: ['*']
          allowedMethods: ['*']
          allowedHeaders: ['*']
          maxAge: 3600
        }
      }
      registries: [
        {
          server: containerRegistry.properties.loginServer
          identity: grpcIdentity.id
        }
      ]
      secrets: [
        {
          name: 'cosmos-key'
          value: cosmosAccount.listKeys().primaryMasterKey
        }
        {
          name: 'jwt-secret'
          value: jwtSecret
        }
        {
          name: 'github-client-id'
          value: githubClientId
        }
        {
          name: 'github-client-secret'
          value: githubClientSecret
        }
      ]
    }
    template: {
      containers: [
        {
          name: 'grpc-api'
          image: '${containerRegistry.properties.loginServer}/baluster-grpc:latest'
          resources: {
            cpu: json('0.5')
            memory: '1Gi'
          }
          env: [
            {
              name: 'PORT'
              value: '5050'
            }
            {
              name: 'COSMOS_ENDPOINT'
              value: cosmosAccount.properties.documentEndpoint
            }
            {
              name: 'COSMOS_KEY'
              secretRef: 'cosmos-key'
            }
            {
              name: 'COSMOS_DATABASE'
              value: cosmosDbDatabaseNameFinal
            }
            {
              name: 'JWT_SECRET'
              secretRef: 'jwt-secret'
            }
            {
              name: 'GITHUB_CLIENT_ID'
              secretRef: 'github-client-id'
            }
            {
              name: 'GITHUB_CLIENT_SECRET'
              secretRef: 'github-client-secret'
            }
            {
              name: 'GITHUB_REDIRECT_URL'
              value: githubRedirectUrl
            }
          ]
        }
      ]
      scale: {
        minReplicas: 1
        maxReplicas: 3
      }
    }
  }
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${grpcIdentity.id}': {}
    }
  }
  tags: tags
  dependsOn: [
    grpcAcrRoleAssignment
  ]
}

// Static Web App
resource staticWebApp 'Microsoft.Web/staticSites@2023-01-01' = {
  name: staticWebAppNameFinal
  location: location
  sku: {
    name: 'Free'
    tier: 'Free'
  }
  properties: {
    buildProperties: {
      appLocation: '/'
      outputLocation: 'dist'
    }
  }
  tags: tags
}

// Determine if domain is apex (root) domain
// Apex domains have only one dot (e.g., "example.com"), subdomains have two or more (e.g., "app.example.com")
var isApexDomain = !empty(customDomainName) && length(split(customDomainName, '.')) == 2

// Custom Domain for Static Web App (optional)
@description('Custom domain resource for Static Web App')
resource staticWebAppCustomDomain 'Microsoft.Web/staticSites/customDomains@2023-01-01' = if (!empty(customDomainName)) {
  parent: staticWebApp
  name: customDomainName
  properties: {
    validationMethod: isApexDomain ? 'dns-txt-token' : 'cname-delegation'
  }
}