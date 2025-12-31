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
  const organization = user?.organization;

  // Show onboarding if user has no organization (only when we've confirmed there is none)
  const hasOrganization = !!organization;
  const showOnboarding = !authLoading && !hasOrganization;

  const { data: applicationsResponse, isLoading: appsLoading } = useQuery<{
    applications: Application[];
  }>({
    queryKey: ["applications", organization?.id],
    queryFn: async () => {
      if (!organization?.id) return { applications: [] };
      return api.applications.list(organization.id);
    },
    enabled: hasOrganization && !authLoading,
  });

  const allApplications = applicationsResponse?.applications || [];

  const { data: serviceKeysData, isLoading: keysLoading } = useQuery<ServiceKey[]>({
    queryKey: ["serviceKeys", organization?.id],
    queryFn: async () => {
      if (!organization?.id) return [];
      const result = await api.serviceKeys.list(organization.id);
      return Array.isArray(result) ? result : [];
    },
    enabled: hasOrganization && !authLoading,
  });

  const serviceKeys = Array.isArray(serviceKeysData) ? serviceKeysData : [];

  const { data: apiKeysData, isLoading: apiKeysLoading } = useQuery<ApiKey[]>({
    queryKey: ["apiKeys", organization?.id],
    queryFn: async () => {
      if (!organization?.id) return [];
      const result = await api.apiKeys.list(organization.id);
      return Array.isArray(result) ? result : [];
    },
    enabled: hasOrganization && !authLoading,
  });

  const apiKeys = Array.isArray(apiKeysData) ? apiKeysData : [];

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
            <div className="table-container">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>ID</th>
                    <th>Description</th>
                    <th>Permissions</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {allApplications?.map((app) => (
                    <tr key={app.id}>
                      <td className="table-cell-name">{app.name}</td>
                      <td className="table-cell-id">{app.id.slice(0, 8)}...</td>
                      <td className="table-cell-description">
                        {app.description || <span className="text-muted">—</span>}
                      </td>
                      <td>
                        {app.permissions && app.permissions.length > 0 ? (
                          <span className="badge">
                            {app.permissions.length} permission
                            {app.permissions.length !== 1 ? "s" : ""}
                          </span>
                        ) : (
                          <span className="text-muted">—</span>
                        )}
                      </td>
                      <td>
                        <Link to={`/applications/${app.id}`} className="table-link">
                          View details →
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
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
            <div className="table-container">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>ID</th>
                    <th>Applications</th>
                    <th>Expires</th>
                    <th>Created</th>
                  </tr>
                </thead>
                <tbody>
                  {serviceKeys?.map((key) => (
                    <tr key={key.id}>
                      <td className="table-cell-name">{key.name}</td>
                      <td className="table-cell-id">{key.id.slice(0, 8)}...</td>
                      <td>
                        {key.applications && key.applications.length > 0 ? (
                          <span className="badge">
                            {key.applications.length} application
                            {key.applications.length !== 1 ? "s" : ""}
                          </span>
                        ) : (
                          <span className="text-muted">—</span>
                        )}
                      </td>
                      <td>
                        {key.expires_at ? (
                          new Date(key.expires_at).toLocaleDateString()
                        ) : (
                          <span className="text-muted">Never</span>
                        )}
                      </td>
                      <td className="table-cell-date">
                        {new Date(key.created_at).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
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
            <div className="table-container">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>ID</th>
                    <th>Application ID</th>
                    <th>Expires</th>
                    <th>Created</th>
                  </tr>
                </thead>
                <tbody>
                  {apiKeys?.map((key) => (
                    <tr key={key.id}>
                      <td className="table-cell-name">{key.name}</td>
                      <td className="table-cell-id">{key.id.slice(0, 8)}...</td>
                      <td className="table-cell-id">{key.application_id.slice(0, 8)}...</td>
                      <td>
                        {key.expires_at ? (
                          new Date(key.expires_at).toLocaleDateString()
                        ) : (
                          <span className="text-muted">Never</span>
                        )}
                      </td>
                      <td className="table-cell-date">
                        {new Date(key.created_at).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}
    </>
  );
}
