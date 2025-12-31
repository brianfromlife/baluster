import axios from "axios";
import { getToken, setToken } from "./auth";

const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";

const apiClient = axios.create({
  baseURL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Add auth token and organization ID to requests
apiClient.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  const url = config.url || "";
  const isAdminRoute = url.startsWith("/admin/v1") || url.startsWith("admin/v1");
  const shouldSkipOrgHeader =
    url.includes("/me") ||
    (url.includes("/organizations") && config.method?.toLowerCase() === "post");

  if (isAdminRoute && !shouldSkipOrgHeader) {
    const orgId = localStorage.getItem("baluster_org_id");
    if (orgId) {
      config.headers["x-org-id"] = orgId;
    }
  }

  return config;
});

apiClient.interceptors.response.use(
  (response) => {
    const refreshedToken = response.headers["x-refreshed-token"];
    if (refreshedToken) {
      setToken(refreshedToken);
    }
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid - clear it
      localStorage.removeItem("baluster_token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);

// API Types

// Request Types
export interface CreateOrganizationRequest {
  name: string;
}

export interface CreateApplicationRequest {
  name: string;
  description?: string;
  permissions?: string[];
}

export interface UpdateApplicationRequest {
  name: string;
  description?: string;
  permissions?: string[];
}

export interface CreateServiceKeyRequest {
  name: string;
  applications: Array<{
    application_id: string;
    application_name: string;
    permissions: string[];
  }>;
  expires_at?: string;
}

export interface CreateApiKeyRequest {
  application_id: string;
  name: string;
  expires_at?: string;
}

export interface CreateRoleRequest {
  application_id: string;
  name: string;
  description?: string;
}

export interface GithubOAuthParams {
  redirect_uri: string;
}

export interface GithubOAuthCallbackParams {
  code: string;
  state: string;
}

// Response Types
export interface Organization {
  id: string;
  name: string;
  created_at?: string;
  updated_at?: string;
}

export interface UserResponse {
  id: string;
  github_id: string;
  username: string;
  email: string;
  avatar_url: string;
  organization?: Organization;
}

export interface OrganizationListResponse {
  organizations: Organization[];
}

export interface Application {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  permissions?: string[];
  created_at: string;
  updated_at: string;
}

export interface ApplicationListResponse {
  applications: Application[];
}

export interface ServiceKey {
  id: string;
  organization_id: string;
  name: string;
  applications: Array<{
    application_id: string;
    application_name: string;
    permissions: string[];
  }>;
  expires_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ServiceKeyResponse extends ServiceKey {
  token_value?: string;
}

export interface ServiceKeyListResponse {
  service_keys?: ServiceKey[];
}

export interface ApiKey {
  id: string;
  organization_id: string;
  application_id: string;
  name: string;
  expires_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ApiKeyResponse extends ApiKey {
  token_value?: string;
}

export interface ApiKeyListResponse {
  api_keys?: ApiKey[];
}

export interface Role {
  id: string;
  application_id: string;
  name: string;
  description?: string;
  created_at?: string;
  updated_at?: string;
}

export interface RoleListResponse {
  roles: Role[];
}

export interface OAuthResponse {
  auth_url?: string;
  url?: string;
  token?: string;
  [key: string]: unknown;
}

export interface AuditHistory {
  id: string;
  entity_id: string;
  organization_id: string;
  action: "created" | "updated" | "deleted";
  created_by_user_id: string;
  created_by_github_id: string;
  created_by_username: string;
  created_at: string;
}

export interface AuditHistoryResponse {
  history: AuditHistory[];
}

// Consolidated API Object
export const api = {
  auth: {
    githubOAuth: async (redirectUri: string): Promise<OAuthResponse> => {
      const response = await apiClient.get<OAuthResponse>("/api/auth/github", {
        params: { redirect_uri: redirectUri },
      });
      return response.data;
    },

    githubOAuthCallback: async (code: string, state: string): Promise<OAuthResponse> => {
      const response = await apiClient.get<OAuthResponse>("/api/auth/github/callback", {
        params: { code, state },
      });
      return response.data;
    },

    getCurrentUser: async (): Promise<UserResponse> => {
      const response = await apiClient.get<UserResponse>("admin/v1/me");
      return response.data;
    },

    logout: async (): Promise<void> => {
      await apiClient.post("/api/auth/logout");
    },
  },

  organizations: {
    list: async (): Promise<OrganizationListResponse> => {
      const response = await apiClient.get<OrganizationListResponse>("/admin/v1/organizations");
      return response.data;
    },

    listMy: async (): Promise<OrganizationListResponse> => {
      const response = await apiClient.get<OrganizationListResponse>("/admin/v1/me/organizations");
      return response.data;
    },

    create: async (data: CreateOrganizationRequest): Promise<Organization> => {
      const response = await apiClient.post<Organization>("/admin/v1/organizations", data);
      return response.data;
    },
  },

  applications: {
    list: async (organizationId: string): Promise<ApplicationListResponse> => {
      const response = await apiClient.get<ApplicationListResponse>(
        `/admin/v1/organizations/${organizationId}/applications`,
        {
          headers: { "x-org-id": organizationId },
        }
      );
      return response.data;
    },

    create: async (data: CreateApplicationRequest): Promise<Application> => {
      const response = await apiClient.post<Application>("/admin/v1/applications", data);
      return response.data;
    },

    get: async (applicationId: string): Promise<Application> => {
      const response = await apiClient.get<Application>(`/admin/v1/applications/${applicationId}`);
      return response.data;
    },

    getHistory: async (applicationId: string): Promise<AuditHistoryResponse> => {
      const response = await apiClient.get<AuditHistoryResponse>(
        `/admin/v1/applications/${applicationId}/history`
      );
      return response.data;
    },

    update: async (applicationId: string, data: UpdateApplicationRequest): Promise<Application> => {
      const response = await apiClient.put<Application>(
        `/admin/v1/applications/${applicationId}`,
        data
      );
      return response.data;
    },
  },

  roles: {
    list: async (applicationId: string): Promise<RoleListResponse> => {
      const response = await apiClient.get<RoleListResponse>(
        `/api/v1/applications/${applicationId}/roles`
      );
      return response.data;
    },

    create: async (data: CreateRoleRequest): Promise<Role> => {
      const response = await apiClient.post<Role>("/api/v1/roles", data);
      return response.data;
    },
  },

  serviceKeys: {
    list: async (organizationId: string): Promise<ServiceKey[]> => {
      const response = await apiClient.get<ServiceKey[]>(
        `/admin/v1/organizations/${organizationId}/service-keys`,
        {
          headers: { "x-org-id": organizationId },
        }
      );
      return response.data;
    },

    get: async (serviceKeyId: string): Promise<ServiceKeyResponse> => {
      const response = await apiClient.get<ServiceKeyResponse>(
        `/admin/v1/service-keys/${serviceKeyId}`
      );
      return response.data;
    },

    getHistory: async (serviceKeyId: string): Promise<AuditHistoryResponse> => {
      const response = await apiClient.get<AuditHistoryResponse>(
        `/admin/v1/service-keys/${serviceKeyId}/history`
      );
      return response.data;
    },

    create: async (data: CreateServiceKeyRequest): Promise<ServiceKeyResponse> => {
      const response = await apiClient.post<ServiceKeyResponse>("/admin/v1/service-keys", data);
      return response.data;
    },
  },

  apiKeys: {
    list: async (organizationId: string): Promise<ApiKey[]> => {
      const response = await apiClient.get<ApiKey[]>(
        `/admin/v1/organizations/${organizationId}/api-keys`,
        {
          headers: { "x-org-id": organizationId },
        }
      );
      return response.data;
    },

    get: async (tokenId: string): Promise<ApiKeyResponse> => {
      const response = await apiClient.get<ApiKeyResponse>(`/admin/v1/api-keys/${tokenId}`);
      return response.data;
    },

    getHistory: async (tokenId: string): Promise<AuditHistoryResponse> => {
      const response = await apiClient.get<AuditHistoryResponse>(
        `/admin/v1/api-keys/${tokenId}/history`
      );
      return response.data;
    },

    create: async (data: CreateApiKeyRequest): Promise<ApiKeyResponse> => {
      const response = await apiClient.post<ApiKeyResponse>("/admin/v1/api-keys", data);
      return response.data;
    },
  },
};
