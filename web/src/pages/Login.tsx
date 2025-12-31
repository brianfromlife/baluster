import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { api } from "@/lib/api";
import "./Login.css";

export function LoginPage() {
  const navigate = useNavigate();
  const { user } = useAuth();

  useEffect(() => {
    if (user) {
      navigate("/dashboard");
    }
  }, [user, navigate]);

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

  return (
    <div className="login-page">
      <div className="login-content">
        <div className="login-icon">
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
            <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"></path>
            <polyline points="10 17 15 12 10 7"></polyline>
            <line x1="15" y1="12" x2="3" y2="12"></line>
          </svg>
        </div>
        <h1 className="login-title">Welcome to Baluster</h1>
        <p className="login-description">Access Token Generator and Role Management</p>
        <button onClick={handleGitHubLogin} className="login-button">
          Login with GitHub
        </button>
      </div>
    </div>
  );
}
