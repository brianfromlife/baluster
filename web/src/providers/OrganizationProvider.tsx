import { useState, useCallback, useMemo, type ReactNode } from "react";
import { OrganizationContext } from "@/contexts/OrganizationContext";

const ORG_ID_KEY = "baluster_org_id";

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const [organizationId, setOrganizationIdState] = useState<string | null>(() => {
    if (typeof window !== "undefined") {
      return localStorage.getItem(ORG_ID_KEY);
    }
    return null;
  });

  const setOrganizationId = useCallback((id: string | null) => {
    setOrganizationIdState(id);
    if (id) {
      localStorage.setItem(ORG_ID_KEY, id);
    } else {
      localStorage.removeItem(ORG_ID_KEY);
    }
  }, []);

  const value = useMemo(
    () => ({ organizationId, setOrganizationId }),
    [organizationId, setOrganizationId]
  );

  return <OrganizationContext.Provider value={value}>{children}</OrganizationContext.Provider>;
}
