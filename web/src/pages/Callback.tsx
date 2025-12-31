import { useEffect, useRef } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { api } from "@/lib/api";

export function CallbackPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { login } = useAuth();
  const hasProcessed = useRef(false);

  useEffect(() => {
    // Prevent multiple executions
    if (hasProcessed.current) {
      return;
    }

    const code = searchParams.get("code");
    const state = searchParams.get("state");

    if (!code || !state) {
      console.error("Missing code or state in callback URL");
      navigate("/login");
      return;
    }

    // Mark as processed immediately to prevent duplicate calls
    hasProcessed.current = true;

    const handleCallback = async () => {
      try {
        const response = await api.auth.githubOAuthCallback(code, state);

        if (response.token) {
          login(response.token);
          navigate("/dashboard");
        } else {
          console.error("No token received from callback");
          navigate("/login");
        }
      } catch (error) {
        console.error("Failed to complete OAuth callback:", error);
        navigate("/login");
      }
    };

    handleCallback();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty deps - only run once on mount

  return (
    <div style={{ padding: "2rem", textAlign: "center" }}>
      <p>Completing authentication...</p>
    </div>
  );
}
