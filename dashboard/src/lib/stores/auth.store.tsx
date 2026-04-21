"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { configureApiClient } from "../api-client";
import { authService } from "../services/auth.service";
import type { LoginCredentials } from "../types/auth.types";
import type { AdminUser } from "../types/user.types";

type AuthContextValue = {
  login: (identifier: string, password: string) => Promise<AdminUser>;
  logout: () => Promise<boolean>;
  isAuthenticated: boolean;
  isLoading: boolean;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

const STORAGE_KEY = "culturae_auth";
const USER_STORAGE_KEY = "culturae_user";

function loadAuthFromStorage(): boolean {
  if (typeof window === "undefined") return false;

  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return false;

    const parsed = JSON.parse(raw);
    return parsed.isAuthenticated ?? false;
  } catch {
    return false;
  }
}

function saveAuthToStorage(isAuthenticated: boolean) {
  if (typeof window === "undefined") return;

  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ isAuthenticated }));

    // Notify other tabs of auth change
    window.dispatchEvent(
      new CustomEvent("authChange", {
        detail: { isAuthenticated },
      })
    );
  } catch (error) {
    console.error("Failed to save auth to localStorage:", error);
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(() =>
    loadAuthFromStorage()
  );
  const [isLoading, setIsLoading] = useState(true);

  const checkAuth = useCallback(async () => {
    try {
      const response = await fetch("/api/v1/admin/me", {
        credentials: "include",
      });
      setIsAuthenticated(response.ok);
    } catch {
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!isLoading) {
      saveAuthToStorage(isAuthenticated);
    }
  }, [isAuthenticated, isLoading]);

  useEffect(() => {
    if (typeof window === "undefined") return;

    const handleAuthChange = (e: CustomEvent) => {
      setIsAuthenticated(e.detail?.isAuthenticated ?? false);
    };

    window.addEventListener(
      "authChange",
      handleAuthChange as EventListener
    );

    return () => {
      window.removeEventListener(
        "authChange",
        handleAuthChange as EventListener
      );
    };
  }, []);

  useEffect(() => {
    configureApiClient(
      async () => {
        try {
          const response = await fetch("/api/v1/auth/refresh", {
            method: "POST",
            credentials: "include",
          });
          return response.ok;
        } catch {
          return false;
        }
      },
      () => {
        setIsAuthenticated(false);
      },
    );
    checkAuth();
  }, [checkAuth]);

  const login = useCallback(async (identifier: string, password: string) => {
    const credentials: LoginCredentials = { identifier, password };
    const response = await authService.login(credentials);
    setIsAuthenticated(true);
    return response.user;
  }, []);

  const logout = useCallback(async () => {
    try {
      await fetch("/api/v1/auth/logout", {
        method: "POST",
        credentials: "include",
      });
    } catch {
      // Ignore logout errors
    }
    setIsAuthenticated(false);

    try {
      localStorage.removeItem(USER_STORAGE_KEY);
    } catch {
      // Ignore
    }

    saveAuthToStorage(false);
    return true;
  }, []);

  return (
    <AuthContext.Provider value={{ login, logout, isAuthenticated, isLoading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
