import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { api } from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import { z } from "zod";
import styles from "./CreateApplicationForm.module.css";

const PERMISSION_EXAMPLES = ["admin", "basic", "user.update.address", "user.delete"];

export function CreateApplicationForm() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [permissions, setPermissions] = useState<string[]>([]);
  const [permissionInput, setPermissionInput] = useState("");
  const [error, setError] = useState<string | null>(null);

  const organization = user?.organization;

  const createMutation = useMutation({
    mutationFn: (data: {
      organization_id: string;
      name: string;
      description?: string;
      permissions?: string[];
    }) => api.applications.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["applications", organization?.id] });
      navigate("/dashboard");
    },
    onError: (err: unknown) => {
      if (err instanceof Error) {
        setError(err.message || "Failed to create application");
      } else {
        setError("Failed to create application");
      }
    },
  });

  const applicationSchema = z.object({
    name: z.string().min(1, "Application name is required").trim(),
    organization_id: z.string().min(1, "Organization is required"),
    description: z.string().trim().optional(),
    permissions: z.array(z.string()).optional(),
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!organization?.id) {
      setError("Organization is required");
      return;
    }

    const result = applicationSchema.safeParse({
      name,
      organization_id: organization.id,
      description: description.trim() || undefined,
      permissions: permissions.length > 0 ? permissions : undefined,
    });

    if (!result.success) {
      setError(result.error.errors[0].message);
      return;
    }

    createMutation.mutate(result.data);
  };

  const addPermission = () => {
    const trimmed = permissionInput.trim();
    if (trimmed && !permissions.includes(trimmed)) {
      setPermissions([...permissions, trimmed]);
      setPermissionInput("");
    }
  };

  const removePermission = (permission: string) => {
    setPermissions(permissions.filter((p) => p !== permission));
  };

  const addExamplePermission = (example: string) => {
    if (!permissions.includes(example)) {
      setPermissions([...permissions, example]);
    }
  };

  const handlePermissionInputKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      addPermission();
    }
  };

  return (
    <div className={styles.createApplication}>
      <div className={styles.createApplicationContent}>
        <h1 className={styles.createApplicationTitle}>Create Application</h1>
        <p className={styles.createApplicationDescription}>
          Create a new application to manage permissions and access control for your services.
        </p>

        <form onSubmit={handleSubmit} className={styles.createApplicationForm}>
          <div className={styles.formGroup}>
            <label htmlFor="app-name" className={styles.formLabel}>
              Application Name *
            </label>
            <input
              id="app-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="user_service, mail_service"
              className={styles.formInput}
              disabled={createMutation.isPending}
              autoFocus
              autoComplete="off"
            />
            <span className={styles.formHint}>
              Use underscores instead of spaces (e.g., user_service)
            </span>
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="app-description" className={styles.formLabel}>
              Description
            </label>
            <textarea
              id="app-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe what this application does..."
              className={styles.formTextarea}
              disabled={createMutation.isPending}
              rows={3}
            />
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="permissions" className={styles.formLabel}>
              Permissions
            </label>
            <div className={styles.permissionsInputGroup}>
              <input
                id="permissions"
                type="text"
                value={permissionInput}
                onChange={(e) => setPermissionInput(e.target.value)}
                onKeyDown={handlePermissionInputKeyDown}
                placeholder="Enter permission name"
                className={styles.formInput}
                disabled={createMutation.isPending}
              />
              <button
                type="button"
                onClick={addPermission}
                className={styles.permissionAddButton}
                disabled={createMutation.isPending || !permissionInput.trim()}
              >
                Add
              </button>
            </div>
            <div className={styles.permissionExamples}>
              <span className={styles.permissionExamplesLabel}>Examples:</span>
              {PERMISSION_EXAMPLES.map((example) => (
                <button
                  key={example}
                  type="button"
                  onClick={() => addExamplePermission(example)}
                  className={styles.permissionExampleButton}
                  disabled={createMutation.isPending || permissions.includes(example)}
                >
                  {example}
                </button>
              ))}
            </div>
            {permissions.length > 0 && (
              <div className={styles.permissionsList}>
                {permissions.map((permission) => (
                  <div key={permission} className={styles.permissionTag}>
                    <span>{permission}</span>
                    <button
                      type="button"
                      onClick={() => removePermission(permission)}
                      className={styles.permissionRemove}
                      disabled={createMutation.isPending}
                      aria-label={`Remove ${permission}`}
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

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
              disabled={createMutation.isPending || !name.trim()}
            >
              {createMutation.isPending ? "Creating..." : "Create Application"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
