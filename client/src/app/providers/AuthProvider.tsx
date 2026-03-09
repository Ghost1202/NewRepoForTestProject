import React, { createContext, useContext, useEffect, useMemo, useState } from "react";
import { AuthResponse, User } from "@/entities/user";
import { ssoApi } from "@/shared/api";
import { authEvents, authStore } from "@/shared/auth/token";

type AuthContextValue = {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (payload: AuthResponse) => void;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

const extractToken = (payload: AuthResponse) => payload.token ?? payload.access_token ?? null;

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [token, setToken] = useState<string | null>(authStore.getToken());
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  useEffect(() => {
    const unsubscribe = authEvents.onLogout(() => {
      setToken(null);
      setUser(null);
    });
    return unsubscribe;
  }, []);

  useEffect(() => {
    const initialize = async () => {
      const storedToken = authStore.getToken();

      try {
        const response = await ssoApi.get<User>("/users/login", { authMode: "optional" });
        setUser(response.data);
        setToken(storedToken);
      } catch {
        if (storedToken) {
          authStore.clearToken();
          setToken(null);
        }
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    void initialize();
  }, []);

  const login = (payload: AuthResponse) => {
    const nextToken = extractToken(payload);
    if (nextToken) {
      authStore.setToken(nextToken);
      setToken(nextToken);
    }
    if (payload.user) {
      setUser(payload.user);
    }
  };

  const logout = () => {
    authStore.clearToken();
    setToken(null);
    setUser(null);
  };

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      isAuthenticated: Boolean(token || user),
      isLoading,
      login,
      logout,
    }),
    [user, token, isLoading],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
};
