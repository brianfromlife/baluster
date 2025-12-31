import { useState, useEffect, useCallback, useMemo, useRef, type ReactNode } from "react";
import { getToken, setToken, removeToken } from "@/lib/auth";
import { api } from "@/lib/api";
import { AuthContext, type User } from "@/contexts/AuthContext";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const isFetchingRef = useRef(false);

  const fetchUser = useCallback(async () => {
    // Prevent multiple simultaneous fetches
    if (isFetchingRef.current) {
      return;
    }
    isFetchingRef.current = true;

    try {
      const response = await api.auth.getCurrentUser();
      if (response) {
        setUser({
          id: response.id,
          githubId: response.github_id,
          username: response.username,
          email: response.email,
          avatarUrl: response.avatar_url,
          organization: response.organization || undefined,
        });
        if (response.organization?.id) {
          localStorage.setItem("baluster_org_id", response.organization.id);
        } else {
          localStorage.removeItem("baluster_org_id");
        }
      }
    } catch (error) {
      console.error("Failed to fetch user:", error);
      removeToken();
    } finally {
      setLoading(false);
      isFetchingRef.current = false;
    }
  }, []);

  useEffect(() => {
    const token = getToken();
    if (token) {
      fetchUser();
    } else {
      setLoading(false);
    }
    // Only run once on mount
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const login = useCallback(
    (token: string) => {
      setToken(token);
      fetchUser();
    },
    [fetchUser]
  );

  const logout = useCallback(async () => {
    try {
      await api.auth.logout();
    } catch (error) {
      console.error("Failed to logout:", error);
    } finally {
      removeToken();
      setUser(null);
    }
  }, []);

  const refetchUser = useCallback(async () => {
    // Reset the guard to allow a fresh fetch
    isFetchingRef.current = false;
    await fetchUser();
  }, [fetchUser]);

  const value = useMemo(
    () => ({ user, loading, login, logout, refetchUser }),
    [user, loading, login, logout, refetchUser]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
