import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { api } from "@/lib/api";
import { Loading } from "@/components/Loading";
import "./Landing.css";

export function LandingPage() {
  const navigate = useNavigate();
  const { user, loading } = useAuth();

  useEffect(() => {
    if (!loading && user) {
      navigate("/dashboard");
    }
  }, [user, loading, navigate]);

  const handleGitHubLogin = async () => {
    try {
      const response = await api.auth.githubOAuth(window.location.origin + "/auth/callback");

      // Redirect to GitHub OAuth
      if (response.auth_url) {
        window.location.href = response.auth_url;
      } else {
        console.error("No auth_url in OAuth response");
      }
    } catch (error) {
      console.error("Failed to initiate OAuth:", error);
    }
  };

  if (loading) {
    return (
      <div className="landing-page">
        <Loading />
      </div>
    );
  }

  return (
    <div className="landing-page">
      <header className="landing-header">
        <div className="landing-header-content">
          <h1 className="landing-logo">Baluster</h1>
          <button onClick={handleGitHubLogin} className="landing-sign-in-button">
            Sign In
          </button>
        </div>
      </header>

      <main className="landing-main">
        <div className="landing-hero">
          <h2 className="landing-hero-title">Access Token Generator and Role Management</h2>
          <p className="landing-hero-description">
            Securely manage access tokens, applications, and permissions for your organization.
          </p>
          <button onClick={handleGitHubLogin} className="landing-cta-button">
            Get Started with GitHub
          </button>
        </div>

        <div className="landing-features">
          <div className="landing-feature">
            <div className="feature-icon">
              <svg
                width="48"
                height="48"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
            </div>
            <h3 className="feature-title">Secure Access</h3>
            <p className="feature-description">
              Manage API keys and service keys with fine-grained permissions and expiration
              controls.
            </p>
          </div>

          <div className="landing-feature">
            <div className="feature-icon">
              <svg
                width="48"
                height="48"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
              </svg>
            </div>
            <h3 className="feature-title">Application Management</h3>
            <p className="feature-description">
              Organize and manage your applications with custom permissions and access controls.
            </p>
          </div>

          <div className="landing-feature">
            <div className="feature-icon">
              <svg
                width="48"
                height="48"
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
            <h3 className="feature-title">Organization Control</h3>
            <p className="feature-description">
              Centralized management for teams with organization-level access and permissions.
            </p>
          </div>
        </div>
      </main>
    </div>
  );
}
