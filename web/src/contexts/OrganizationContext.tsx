import { createContext } from "react";

export interface OrganizationContextType {
  organizationId: string | null;
  setOrganizationId: (id: string | null) => void;
}

export const OrganizationContext = createContext<OrganizationContextType | undefined>(undefined);

