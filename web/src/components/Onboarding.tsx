import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import styles from "./Onboarding.module.css";

interface OnboardingProps {
  onComplete: () => Promise<void>;
}

export function Onboarding({ onComplete }: OnboardingProps) {
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const createMutation = useMutation({
    mutationFn: (data: { name: string }) => api.organizations.create(data),
    onSuccess: async (data) => {
      // Set organization ID in localStorage immediately so subsequent API calls include the header
      if (data?.id) {
        localStorage.setItem("baluster_org_id", data.id);
      }
      // Invalidate queries to refresh data
      await queryClient.invalidateQueries({ queryKey: ["organizations"] });
      await new Promise((resolve) => setTimeout(resolve, 300));
      await onComplete();
    },
    onError: (err: unknown) => {
      if (err instanceof Error) {
        setError(err.message || "Failed to create organization");
      } else {
        setError("Failed to create organization");
      }
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const trimmedName = name.trim();
    if (!trimmedName) {
      setError("Organization name is required");
      return;
    }

    if (trimmedName.length < 3) {
      setError("Organization name must be at least 3 characters");
      return;
    }

    createMutation.mutate({
      name: trimmedName,
    });
  };

  return (
    <div className={styles.onboarding}>
      <div className={styles.onboardingContent}>
        <div className={styles.onboardingIcon}>
          <svg
            width="64"
            height="64"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
          </svg>
        </div>
        <h1 className={styles.onboardingTitle}>Welcome to Baluster!</h1>
        <p className={styles.onboardingDescription}>
          Get started by creating your first organization. Organizations help you organize and
          manage your applications and services.
        </p>

        <form onSubmit={handleSubmit} className={styles.onboardingForm}>
          <div className={styles.formGroup}>
            <label htmlFor="org-name" className={styles.formLabel}>
              Organization Name *
            </label>
            <input
              id="org-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Organization"
              className={styles.formInput}
              disabled={createMutation.isPending}
              autoFocus
              autoComplete="off"
            />
          </div>

          {error && <div className={styles.formError}>{error}</div>}

          <button
            type="submit"
            className={styles.formSubmit}
            disabled={createMutation.isPending || !name.trim() || name.trim().length < 3}
          >
            {createMutation.isPending ? "Creating..." : "Create Organization"}
          </button>
        </form>
      </div>
    </div>
  );
}
