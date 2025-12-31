import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, useNavigate } from "react-router-dom";
import { api } from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import "./ApplicationDetail.css";

const PERMISSION_EXAMPLES = ["admin", "basic", "user.update.address", "user.delete"];

interface Application {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  permissions?: string[];
  created_at: string;
  updated_at: string;
}

interface AuditHistory {
  id: string;
  entity_id: string;
  organization_id: string;
  action: "created" | "updated" | "deleted";
  created_by_user_id: string;
  created_by_github_id: string;
  created_by_username: string;
  created_at: string;
}

export function ApplicationDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useAuth();
  const [permissionInput, setPermissionInput] = useState("");
  const [error, setError] = useState<string | null>(null);

  const organization = user?.organization;

  const { data: application, isLoading: applicationLoading } = useQuery<Application>({
    queryKey: ["application", id],
    queryFn: async () => {
      if (!id) throw new Error("Application ID is required");
      if (!organization?.id) throw new Error("Organization ID is required");
      return api.applications.get(id);
    },
    enabled: !!id && !!organization?.id,
  });

  const { data: historyData, isLoading: historyLoading } = useQuery<{ history: AuditHistory[] }>({
    queryKey: ["application", id, "history"],
    queryFn: async () => {
      if (!id) throw new Error("Application ID is required");
      if (!organization?.id) throw new Error("Organization ID is required");
      return api.applications.getHistory(id);
    },
    enabled: !!id && !!organization?.id,
  });

  const history = historyData?.history || [];
  const isLoading = applicationLoading || historyLoading;

  const updateMutation = useMutation({
    mutationFn: (data: { name: string; description?: string; permissions?: string[] }) => {
      if (!id) throw new Error("Application ID is required");
      if (!organization?.id) throw new Error("Organization ID is required");
      return api.applications.update(id, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["application", id] });
      queryClient.invalidateQueries({ queryKey: ["application", id, "history"] });
      queryClient.invalidateQueries({ queryKey: ["applications"] });
      setError(null);
    },
    onError: (err: unknown) => {
      if (err instanceof Error) {
        setError(err.message || "Failed to update application");
      } else {
        setError("Failed to update application");
      }
    },
  });

  const addPermission = () => {
    if (!application) return;

    const trimmed = permissionInput.trim();
    if (trimmed && !application.permissions?.includes(trimmed)) {
      const updatedPermissions = [...(application.permissions || []), trimmed];
      updateMutation.mutate({
        name: application.name,
        description: application.description,
        permissions: updatedPermissions,
      });
      setPermissionInput("");
    }
  };

  const removePermission = (permission: string) => {
    if (!application) return;

    const updatedPermissions = (application.permissions || []).filter((p) => p !== permission);
    updateMutation.mutate({
      name: application.name,
      description: application.description,
      permissions: updatedPermissions,
    });
  };

  const addExamplePermission = (example: string) => {
    if (!application || application.permissions?.includes(example)) return;

    const updatedPermissions = [...(application.permissions || []), example];
    updateMutation.mutate({
      name: application.name,
      description: application.description,
      permissions: updatedPermissions,
    });
  };

  const handlePermissionInputKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      addPermission();
    }
  };

  if (isLoading) {
    return <div className="application-detail-loading">Loading application...</div>;
  }

  if (!application) {
    return (
      <div className="application-detail-error">
        <p>Application not found</p>
        <button onClick={() => navigate("/dashboard")} className="back-button">
          Back to Dashboard
        </button>
      </div>
    );
  }

  return (
      <section className="application-detail-section">
        <div className="application-detail-header">
          <button onClick={() => navigate("/dashboard")} className="back-link">
            ← Back to Dashboard
          </button>
          <h1 className="application-detail-title">{application.name}</h1>
          {application.description && <p className="application-detail-description">{application.description}</p>}
        </div>

        <div className="application-detail-content">
          <div className="permissions-section">
            <h2 className="permissions-section-title">Permissions</h2>
            <p className="permissions-section-description">
              Add or remove permissions for this application. Permissions define what actions can be performed.
            </p>

            <div className="permissions-input-group">
              <input
                type="text"
                value={permissionInput}
                onChange={(e) => setPermissionInput(e.target.value)}
                onKeyDown={handlePermissionInputKeyDown}
                placeholder="Enter permission name"
                className="form-input"
                disabled={updateMutation.isPending}
              />
              <button
                type="button"
                onClick={addPermission}
                className="permission-add-button"
                disabled={updateMutation.isPending || !permissionInput.trim()}
              >
                Add
              </button>
            </div>

            <div className="permission-examples">
              <span className="permission-examples-label">Examples:</span>
              {PERMISSION_EXAMPLES.map((example) => (
                <button
                  key={example}
                  type="button"
                  onClick={() => addExamplePermission(example)}
                  className="permission-example-button"
                  disabled={updateMutation.isPending || application.permissions?.includes(example)}
                >
                  {example}
                </button>
              ))}
            </div>

            {updateMutation.isPending && (
              <div className="update-status">Updating permissions...</div>
            )}

            {error && <div className="form-error">{error}</div>}

            {application.permissions && application.permissions.length > 0 ? (
              <div className="permissions-list">
                {application.permissions.map((permission) => (
                  <div key={permission} className="permission-tag">
                    <span>{permission}</span>
                    <button
                      type="button"
                      onClick={() => removePermission(permission)}
                      className="permission-remove"
                      disabled={updateMutation.isPending}
                      aria-label={`Remove ${permission}`}
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div className="permissions-empty">
                <p>No permissions yet. Add your first permission above.</p>
              </div>
            )}
          </div>

          <div className="audit-history-section">
            <h2 className="audit-history-section-title">Audit History</h2>
            <p className="audit-history-section-description">
              Track who made changes and when to this application.
            </p>

            {history.length > 0 ? (
              <div className="audit-history-table">
                <table>
                  <thead>
                    <tr>
                      <th>Action</th>
                      <th>User</th>
                      <th>Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    {history.map((record) => (
                      <tr key={record.id}>
                        <td>
                          <span className={`audit-action audit-action-${record.action}`}>
                            {record.action}
                          </span>
                        </td>
                        <td>{record.created_by_username || record.created_by_user_id}</td>
                        <td>{new Date(record.created_at).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="audit-history-empty">
                <p>No audit history available yet.</p>
              </div>
            )}
          </div>
        </div>
      </section>
  );
}
