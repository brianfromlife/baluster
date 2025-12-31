import { useState } from "react";
import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { api } from "@/lib/api";
import { z } from "zod";
import styles from "./CreateApiKeyForm.module.css";

interface Application {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  permissions?: string[];
  created_at: string;
  updated_at: string;
}

export function CreateApiKeyForm() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const organization = user?.organization;

  // Default expiry to 1 year from now
  const getDefaultExpiry = () => {
    const date = new Date();
    date.setFullYear(date.getFullYear() + 1);
    return date.toISOString().split("T")[0]; // Format as YYYY-MM-DD
  };

  const [name, setName] = useState("");
  const [selectedApplicationId, setSelectedApplicationId] = useState("");
  const [expiresAt, setExpiresAt] = useState(getDefaultExpiry());
  const [error, setError] = useState<string | null>(null);
  const [createdToken, setCreatedToken] = useState<string | null>(null);

  const { data: applicationsResponse } = useQuery<{ applications: Application[] }>({
    queryKey: ["applications", organization?.id],
    queryFn: async () => {
      if (!organization?.id) return { applications: [] };
      return api.applications.list(organization.id);
    },
    enabled: !!organization?.id,
  });

  const applications = applicationsResponse?.applications || [];
  const hasApplications = applications.length > 0;

  const createMutation = useMutation({
    mutationFn: (data: {
      organization_id: string;
      application_id: string;
      name: string;
      expires_at?: string;
    }) => api.apiKeys.create(data),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ["apiKeys", organization?.id] });
      setCreatedToken(response.token_value ?? null);
    },
    onError: (err: unknown) => {
      if (err instanceof Error) {
        setError(err.message || "Failed to create API key");
      } else {
        setError("Failed to create API key");
      }
    },
  });

  const apiKeySchema = z.object({
    name: z.string().min(1, "API key name is required").trim(),
    application_id: z.string().min(1, "Application is required"),
    organization_id: z.string().min(1, "Organization is required"),
    expires_at: z.string().optional(),
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setCreatedToken(null);

    if (!organization?.id) {
      setError("Organization is required");
      return;
    }

    const result = apiKeySchema.safeParse({
      name,
      application_id: selectedApplicationId,
      organization_id: organization.id,
      expires_at: expiresAt ? new Date(expiresAt).toISOString() : undefined,
    });

    if (!result.success) {
      setError(result.error.errors[0].message);
      return;
    }

    createMutation.mutate(result.data);
  };

  const handleCopyToken = () => {
    if (createdToken) {
      navigator.clipboard.writeText(createdToken);
      alert("Token copied to clipboard!");
    }
  };

  const handleContinue = () => {
    navigate("/dashboard");
  };

  if (!hasApplications) {
    return (
      <div className={styles.createApiKey}>
        <div className={styles.createApiKeyContent}>
          <h1 className={styles.createApiKeyTitle}>Create API Key</h1>
          <div className={styles.emptyState} style={{ marginTop: "2rem" }}>
            <p>No applications available.</p>
            <p className={styles.emptyStateSubtitle}>
              You need to create at least one application before you can create an API key. API keys
              are used to authenticate requests to Baluster APIs for a specific application.
            </p>
            <Link to="/applications" className={styles.emptyStateLink}>
              Create your first application →
            </Link>
          </div>
        </div>
      </div>
    );
  }

  if (createdToken) {
    return (
      <div className={styles.createApiKey}>
        <div className={styles.createApiKeyContent}>
          <h1 className={styles.createApiKeyTitle}>API Key Created Successfully</h1>
          <p className={styles.createApiKeyDescription}>
            Save this token now - you won't be able to see it again!
          </p>

          <div className={styles.tokenDisplay}>
            <label className={styles.formLabel}>API Key Token</label>
            <div className={styles.tokenInputGroup}>
              <input type="text" readOnly value={createdToken} className={styles.tokenInput} />
              <button type="button" onClick={handleCopyToken} className={styles.copyButton}>
                Copy
              </button>
            </div>
            <p className={styles.tokenWarning}>
              ⚠️ This token will not be shown again. Make sure to save it securely.
            </p>
          </div>

          <button type="button" onClick={handleContinue} className={styles.formSubmit}>
            Continue
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.createApiKey}>
      <div className={styles.createApiKeyContent}>
        <h1 className={styles.createApiKeyTitle}>Create API Key</h1>
        <p className={styles.createApiKeyDescription}>
          API keys are used to authenticate requests to Baluster APIs for a specific application.
          Set an expiry date (default: 1 year).
        </p>

        <form onSubmit={handleSubmit} className={styles.createApiKeyForm}>
          <div className={styles.formGroup}>
            <label htmlFor="api-key-name" className={styles.formLabel}>
              API Key Name*
            </label>
            <input
              id="api-key-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Example: user_service_api_key"
              className={styles.formInput}
              disabled={createMutation.isPending}
              autoFocus
              autoComplete="off"
            />
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="api-key-application" className={styles.formLabel}>
              Application*
            </label>
            <select
              id="api-key-application"
              value={selectedApplicationId}
              onChange={(e) => setSelectedApplicationId(e.target.value)}
              className={styles.formInput}
              disabled={createMutation.isPending}
            >
              <option value="">Select an application</option>
              {applications.map((app) => (
                <option key={app.id} value={app.id}>
                  {app.name}
                </option>
              ))}
            </select>
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="api-key-expiry" className={styles.formLabel}>
              Expires At
            </label>
            <input
              id="api-key-expiry"
              type="date"
              value={expiresAt}
              onChange={(e) => setExpiresAt(e.target.value)}
              className={styles.formInput}
              disabled={createMutation.isPending}
              min={new Date().toISOString().split("T")[0]}
            />
            <p className={styles.formHint}>
              Leave empty for no expiration. Default: 1 year from today.
            </p>
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
              disabled={createMutation.isPending || !name.trim() || !selectedApplicationId}
            >
              {createMutation.isPending ? "Creating..." : "Create API Key"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
