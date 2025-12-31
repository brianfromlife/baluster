import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { api } from "@/lib/api";
import { Onboarding } from "@/components/Onboarding";
import "./Dashboard.css";

interface Application {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  permissions?: string[];
  created_at: string;
  updated_at: string;
}

interface ServiceKey {
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

interface ApiKey {
  id: string;
  organization_id: string;
  application_id: string;
  name: string;
  expires_at?: string;
  created_at: string;
  updated_at: string;
}

export function DashboardPage() {
  const { user, loading: authLoading, refetchUser } = useAuth();

  // Get organization from user context (already fetched via /me endpoint)
  const organization = user?.organization;

  // Show onboarding if user has no organization (only when we've confirmed there is none)
  const hasOrganization = !!organization;
  const showOnboarding = !authLoading && !hasOrganization;

  // Fetch applications for the organization
  const { data: applicationsData, isLoading: appsLoading } = useQuery<
    { applications?: Application[] } | Application[]
  >({
    queryKey: ["applications", organization?.id],
    queryFn: async () => {
      if (!organization?.id) return [];

      const response = await api.applications.list(organization.id);
      return response?.applications || [];
    },
    enabled: hasOrganization && !authLoading,
  });

  // Ensure allApplications is always an array
  const allApplications = Array.isArray(applicationsData)
    ? applicationsData
    : Array.isArray(applicationsData?.applications)
    ? applicationsData.applications
    : [];

  // Fetch service keys for the organization
  const { data: serviceKeysData, isLoading: keysLoading } = useQuery<
    { service_keys?: ServiceKey[] } | ServiceKey[]
  >({
    queryKey: ["serviceKeys", organization?.id],
    queryFn: async () => {
      if (!organization?.id) return [];

      return api.serviceKeys.list(organization.id);
    },
    enabled: hasOrganization && !authLoading,
  });

  // Ensure serviceKeys is always an array
  const serviceKeys = Array.isArray(serviceKeysData)
    ? serviceKeysData
    : Array.isArray(serviceKeysData?.service_keys)
    ? serviceKeysData.service_keys
    : [];

  // Fetch API keys for the organization
  const { data: apiKeysData, isLoading: apiKeysLoading } = useQuery<
    { api_keys?: ApiKey[] } | ApiKey[]
  >({
    queryKey: ["apiKeys", organization?.id],
    queryFn: async () => {
      if (!organization?.id) return [];

      return api.apiKeys.list(organization.id);
    },
    enabled: hasOrganization && !authLoading,
  });

  // Ensure apiKeys is always an array
  const apiKeys = Array.isArray(apiKeysData)
    ? apiKeysData
    : Array.isArray(apiKeysData?.api_keys)
    ? apiKeysData.api_keys
    : [];

  const isLoading = authLoading || appsLoading || keysLoading || apiKeysLoading;

  const handleOnboardingComplete = async () => {
    await refetchUser();
  };

  return (
    <>
      {showOnboarding ? (
        <Onboarding onComplete={handleOnboardingComplete} />
      ) : (
        <section className="dashboard-section">
          {organization && (
            <div className="organization-header">
              <span className="organization-label">Organization:</span>
              <span className="organization-names">{organization.name}</span>
            </div>
          )}
          <div className="section-header">
            <h2 className="section-title">Applications</h2>
            <div className="section-header-right">
              {(!allApplications || allApplications.length > 0) && (
                <Link to="/applications" className="section-create-button">
                  + Create Application
                </Link>
              )}
              <span className="section-count">{allApplications?.length} total</span>
            </div>
          </div>

          {isLoading ? (
            <div className="loading-state">Loading applications...</div>
          ) : !allApplications || allApplications?.length === 0 ? (
            <div className="empty-state">
              <p>No applications found.</p>
              <p className="empty-state-subtitle">
                Get started by creating your first application.
              </p>
              <Link to="/applications" className="empty-state-link">
                Create your first application →
              </Link>
            </div>
          ) : (
            <div className="applications-grid">
              {allApplications?.map((app) => (
                <div key={app.id} className="application-card">
                  <div className="card-header">
                    <h3 className="card-title">{app.name}</h3>
                    <span className="card-id">{app.id.slice(0, 8)}...</span>
                  </div>
                  {app.description && <p className="card-description">{app.description}</p>}
                  {app.permissions && app.permissions.length > 0 && (
                    <div className="card-permissions">
                      <span className="permissions-label">
                        {app.permissions.length} permission{app.permissions.length !== 1 ? "s" : ""}
                      </span>
                    </div>
                  )}
                  <div className="card-footer">
                    <Link to={`/applications/${app.id}`} className="card-link">
                      View details →
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="section-header" style={{ marginTop: "3rem" }}>
            <h2 className="section-title">Service Keys</h2>
            <div className="section-header-right">
              {(!serviceKeys || serviceKeys.length > 0) &&
                allApplications &&
                allApplications.length > 0 && (
                  <Link to="/service-keys/create" className="section-create-button">
                    + Create Service Key
                  </Link>
                )}
              <span className="section-count">{serviceKeys?.length || 0} total</span>
            </div>
          </div>

          {keysLoading ? (
            <div className="loading-state">Loading service keys...</div>
          ) : !serviceKeys || serviceKeys?.length === 0 ? (
            <div className="empty-state">
              <p>No service keys found.</p>
              {allApplications && allApplications.length > 0 ? (
                <>
                  <p className="empty-state-subtitle">
                    Create a service key to allow secure communication between applications.
                  </p>
                  <Link to="/service-keys/create" className="empty-state-link">
                    Create your first service key →
                  </Link>
                </>
              ) : (
                <>
                  <p className="empty-state-subtitle">
                    You need to create at least one application before you can create a service key.
                    Service keys are used to grant access to your applications.
                  </p>
                  <Link to="/applications" className="empty-state-link">
                    Create your first application →
                  </Link>
                </>
              )}
            </div>
          ) : (
            <div className="service-keys-grid">
              {serviceKeys?.map((key) => (
                <div key={key.id} className="service-key-card">
                  <div className="card-header">
                    <h3 className="card-title">{key.name}</h3>
                    <span className="card-id">{key.id.slice(0, 8)}...</span>
                  </div>
                  {key.applications && key.applications.length > 0 && (
                    <div className="card-applications">
                      <span className="applications-label">
                        {key.applications.length} application
                        {key.applications.length !== 1 ? "s" : ""}
                      </span>
                    </div>
                  )}
                  <div className="card-footer">
                    <span className="card-meta">
                      Created {new Date(key.created_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="section-header" style={{ marginTop: "3rem" }}>
            <h2 className="section-title">API Keys</h2>
            <div className="section-header-right">
              {allApplications && allApplications.length > 0 && (
                <Link to="/api-keys/create" className="section-create-button">
                  + Create API Key
                </Link>
              )}
              <span className="section-count">{apiKeys?.length || 0} total</span>
            </div>
          </div>

          {apiKeysLoading ? (
            <div className="loading-state">Loading API keys...</div>
          ) : !apiKeys || apiKeys?.length === 0 ? (
            <div className="empty-state">
              <p>No API keys found.</p>
              {allApplications && allApplications.length > 0 ? (
                <>
                  <p className="empty-state-subtitle">
                    Create an API key to authenticate requests to Baluster APIs for a specific
                    application.
                  </p>
                  <Link to="/api-keys/create" className="empty-state-link">
                    Create your first API key →
                  </Link>
                </>
              ) : (
                <>
                  <p className="empty-state-subtitle">
                    You need to create at least one application before you can create an API key.
                    API keys are used to authenticate requests to Baluster APIs for a specific
                    application.
                  </p>
                  <Link to="/applications" className="empty-state-link">
                    Create your first application →
                  </Link>
                </>
              )}
            </div>
          ) : (
            <div className="service-keys-grid">
              {apiKeys?.map((key) => (
                <div key={key.id} className="service-key-card">
                  <div className="card-header">
                    <h3 className="card-title">{key.name}</h3>
                    <span className="card-id">{key.id.slice(0, 8)}...</span>
                  </div>
                  {key.expires_at && (
                    <div className="card-applications">
                      <span className="applications-label">
                        Expires {new Date(key.expires_at).toLocaleDateString()}
                      </span>
                    </div>
                  )}
                  <div className="card-footer">
                    <span className="card-meta">
                      Created {new Date(key.created_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}
    </>
  );
}
