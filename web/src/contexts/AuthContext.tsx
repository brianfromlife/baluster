import { createContext } from "react";

export interface Organization {
  id: string;
  name: string;
}

export interface User {
  id: string;
  githubId: string;
  username: string;
  email: string;
  avatarUrl: string;
  organization?: Organization;
}

export interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (token: string) => void;
  logout: () => void;
  refetchUser: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);
