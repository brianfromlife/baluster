import { useState, useMemo } from "react";
import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query";
import { useNavigate, Link } from "react-router-dom";
import { api } from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import { ServiceKeySuccessModal } from "@/components/ServiceKeySuccessModal";
import { z } from "zod";
import styles from "./CreateServiceKeyForm.module.css";

interface Application {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  permissions?: string[];
  created_at: string;
  updated_at: string;
}

interface SelectedApplication {
  application: Application;
  selectedPermissions: string[];
}

export function CreateServiceKeyForm() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedApplications, setSelectedApplications] = useState<SelectedApplication[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [showSuccessModal, setShowSuccessModal] = useState(false);
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [createdOrganizationId, setCreatedOrganizationId] = useState<string | null>(null);

  const organization = user?.organization;

  const { data: applicationsResponse } = useQuery<{ applications: Application[] }>({
    queryKey: ["applications", organization?.id],
    queryFn: async () => {
      if (!organization?.id) return { applications: [] };
      return api.applications.list(organization.id);
    },
    enabled: !!organization?.id,
  });

  const applications = useMemo(() => {
    return applicationsResponse?.applications || [];
  }, [applicationsResponse?.applications]);

  const hasApplications = applications.length > 0;

  const filteredApplications = useMemo(() => {
    if (!searchQuery.trim()) return [];
    const selectedIds = new Set(selectedApplications.map((sa) => sa.application.id));
    return applications.filter(
      (app) =>
        !selectedIds.has(app.id) && app.name.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [searchQuery, applications, selectedApplications]);

  const createMutation = useMutation({
    mutationFn: (data: {
      organization_id: string;
      name: string;
      applications: Array<{
        application_id: string;
        application_name: string;
        permissions: string[];
      }>;
    }) => api.serviceKeys.create(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["serviceKeys", organization?.id] });
      if (data.token_value) {
        setCreatedToken(data.token_value);
        setCreatedOrganizationId(data.organization_id || null);
        setShowSuccessModal(true);
      } else {
        navigate("/dashboard");
      }
    },
    onError: (err: unknown) => {
      if (err instanceof Error) {
        setError(err.message || "Failed to create service key");
      } else {
        setError("Failed to create service key");
      }
    },
  });

  const serviceKeySchema = z.object({
    name: z.string().min(3, "Service key name must be at least 3 characters").trim(),
    organization_id: z.string().min(1, "Organization is required"),
    applications: z
      .array(
        z.object({
          application_id: z.string(),
          application_name: z.string(),
          permissions: z.array(z.string()),
        })
      )
      .min(1, "At least one application must be selected"),
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!organization?.id) {
      setError("Organization is required");
      return;
    }

    const result = serviceKeySchema.safeParse({
      name,
      organization_id: organization.id,
      applications: selectedApplications.map((sa) => ({
        application_id: sa.application.id,
        application_name: sa.application.name,
        permissions: sa.selectedPermissions,
      })),
    });

    if (!result.success) {
      setError(result.error.errors[0].message);
      return;
    }

    createMutation.mutate(result.data);
  };

  const handleApplicationSelect = (application: Application) => {
    // Add application with all permissions selected by default
    setSelectedApplications([
      ...selectedApplications,
      {
        application,
        selectedPermissions: [...(application.permissions || [])],
      },
    ]);
    setSearchQuery("");
  };

  const handleRemoveApplication = (applicationId: string) => {
    setSelectedApplications(
      selectedApplications.filter((sa) => sa.application.id !== applicationId)
    );
  };

  const handlePermissionToggle = (applicationId: string, permission: string) => {
    setSelectedApplications(
      selectedApplications.map((sa) => {
        if (sa.application.id !== applicationId) return sa;
        const isSelected = sa.selectedPermissions.includes(permission);
        return {
          ...sa,
          selectedPermissions: isSelected
            ? sa.selectedPermissions.filter((p) => p !== permission)
            : [...sa.selectedPermissions, permission],
        };
      })
    );
  };

  const handleModalClose = () => {
    setShowSuccessModal(false);
    setCreatedToken(null);
    setCreatedOrganizationId(null);
    navigate("/dashboard");
  };

  if (!hasApplications) {
    return (
      <div className={styles.createServiceKey}>
        <div className={styles.createServiceKeyContent}>
          <h1 className={styles.createServiceKeyTitle}>Create Service Key</h1>
          <div className={styles.emptyState} style={{ marginTop: "2rem" }}>
            <p>No applications available.</p>
            <p className={styles.emptyStateSubtitle}>
              You need to create at least one application before you can create a service key.
              Service keys are used to grant access to your applications.
            </p>
            <Link to="/applications" className={styles.emptyStateLink}>
              Create your first application →
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.createServiceKey}>
      <div className={styles.createServiceKeyContent}>
        <h1 className={styles.createServiceKeyTitle}>Create Service Key</h1>
        <p className={styles.createServiceKeyDescription}>
          Create a service key to allow secure communication between applications. Add specific
          permissions for fine-grained access control.
        </p>

        <form onSubmit={handleSubmit} className={styles.createServiceKeyForm}>
          <div className={styles.formGroup}>
            <label htmlFor="key-name" className={styles.formLabel}>
              Service Key Name*
            </label>
            <input
              id="key-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="production-key, staging-key"
              className={styles.formInput}
              disabled={createMutation.isPending}
              autoFocus
              autoComplete="off"
            />
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="application-search" className={styles.formLabel}>
              Applications*
            </label>
            <div className={styles.typeaheadContainer}>
              <input
                id="application-search"
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onBlur={() => {
                  // Delay to allow click on dropdown item
                  setTimeout(() => setSearchQuery(""), 200);
                }}
                placeholder="Search and select applications..."
                className={styles.formInput}
                disabled={createMutation.isPending}
              />
              {searchQuery && filteredApplications.length > 0 && (
                <div className={styles.typeaheadDropdown}>
                  {filteredApplications.map((app) => (
                    <div
                      key={app.id}
                      className={styles.typeaheadItem}
                      onMouseDown={(e) => {
                        e.preventDefault();
                        handleApplicationSelect(app);
                      }}
                    >
                      <div className={styles.typeaheadItemName}>{app.name}</div>
                      {app.description && (
                        <div className={styles.typeaheadItemDesc}>{app.description}</div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {selectedApplications.length > 0 && (
            <div className={styles.formGroup}>
              <label className={styles.formLabel}>Selected Applications & Permissions</label>
              <div className={styles.selectedApplications}>
                {selectedApplications.map((sa) => (
                  <div key={sa.application.id} className={styles.selectedApplication}>
                    <div className={styles.selectedAppHeader}>
                      <h4 className={styles.selectedAppName}>{sa.application.name}</h4>
                      <button
                        type="button"
                        onClick={() => handleRemoveApplication(sa.application.id)}
                        className={styles.removeAppButton}
                        disabled={createMutation.isPending}
                        aria-label={`Remove ${sa.application.name}`}
                      >
                        <svg
                          width="14"
                          height="14"
                          viewBox="0 0 14 14"
                          fill="none"
                          xmlns="http://www.w3.org/2000/svg"
                        >
                          <path
                            d="M10.5 3.5L3.5 10.5M3.5 3.5L10.5 10.5"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          />
                        </svg>
                      </button>
                    </div>
                    {sa.application.description && (
                      <p className={styles.selectedAppDescription}>{sa.application.description}</p>
                    )}
                    {sa.application.permissions && sa.application.permissions.length > 0 ? (
                      <div className={styles.permissionsCheckboxes}>
                        {sa.application.permissions.map((permission) => (
                          <label key={permission} className={styles.permissionCheckbox}>
                            <input
                              type="checkbox"
                              checked={sa.selectedPermissions.includes(permission)}
                              onChange={() => handlePermissionToggle(sa.application.id, permission)}
                              disabled={createMutation.isPending}
                            />
                            <span>{permission}</span>
                          </label>
                        ))}
                      </div>
                    ) : (
                      <p className={styles.noPermissions}>
                        No permissions available for this application
                      </p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {error && <div className={styles.formError}>{error}</div>}

          <div className={styles.formActions}>
            <button
              type="button"
              onClick={() => navigate("/dashboard")}
              className={styles.formCancel}
              disabled={createMutation.isPending}
            >
              Cancel
            </button>
            <button
              type="submit"
              className={styles.formSubmit}
              disabled={
                createMutation.isPending ||
                !name.trim() ||
                name.trim().length < 3 ||
                selectedApplications.length === 0
              }
            >
              {createMutation.isPending ? "Creating..." : "Create Service Key"}
            </button>
          </div>
        </form>
      </div>
      {showSuccessModal && createdToken && (
        <ServiceKeySuccessModal
          token={createdToken}
          organizationId={createdOrganizationId || undefined}
          onClose={handleModalClose}
        />
      )}
    </div>
  );
}
