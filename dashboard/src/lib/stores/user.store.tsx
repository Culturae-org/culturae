"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { apiGet } from "../api-client";
import { ADMIN_PROFILE_ENDPOINTS } from "../api/endpoints";
import type { AdminUser } from "../types/user.types";
import { useAuth } from "./auth.store";

type UserContextValue = {
  user: AdminUser | null;
  userId: string | null;
  setUser: (u: AdminUser | null) => void;
  clearUser: () => void;
  fetchProfile: () => Promise<AdminUser | null>;
  refetchProfile: () => Promise<AdminUser | null>;
  isLoading: boolean;
};

const UserContext = createContext<UserContextValue | undefined>(undefined);

const STORAGE_KEY = "culturae_user";

function loadUserFromStorage(): {
  user: AdminUser | null;
  userId: string | null;
} {
  if (typeof window === "undefined") return { user: null, userId: null };

  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { user: null, userId: null };

    const parsed = JSON.parse(raw);
    const user = parsed.user ?? null;
    return {
      user,
      userId: parsed.userId ?? null,
    };
  } catch {
    return { user: null, userId: null };
  }
}

function saveUserToStorage(user: AdminUser | null, userId: string | null) {
  if (typeof window === "undefined") return;

  try {
    if (user) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ user, userId }));

      window.dispatchEvent(
        new CustomEvent("localStorageChange", {
          detail: { key: STORAGE_KEY, value: JSON.stringify({ user, userId }) },
        }),
      );
    } else {
      localStorage.removeItem(STORAGE_KEY);
      window.dispatchEvent(
        new CustomEvent("localStorageChange", {
          detail: { key: STORAGE_KEY, value: null },
        }),
      );
    }
  } catch (error) {
    console.error("Failed to save user to localStorage:", error);
  }
}

export function UserProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();

  const [user, setUserState] = useState<AdminUser | null>(
    () => loadUserFromStorage().user,
  );
  const [userId, setUserIdState] = useState<string | null>(
    () => loadUserFromStorage().userId,
  );
  const [isLoading, setIsLoading] = useState(false);

  const isUpdatingFromStorage = useRef(false);
  const isUpdatingToStorage = useRef(false);
  const hasFetchedProfile = useRef(false);

  useEffect(() => {
    if (typeof window === "undefined") return;

    const handleStorageChange = (e: StorageEvent) => {
      if (isUpdatingToStorage.current) return;

      if (e.key === STORAGE_KEY && e.newValue) {
        try {
          isUpdatingFromStorage.current = true;
          const parsed = JSON.parse(e.newValue);
          setUserState(parsed.user ?? null);
          setUserIdState(parsed.userId ?? null);
        } catch {
        } finally {
          isUpdatingFromStorage.current = false;
        }
      }
    };

    const handleCustomStorageChange = (e: CustomEvent) => {
      if (e.detail?.key === STORAGE_KEY) {
        handleStorageChange({
          key: STORAGE_KEY,
          newValue: e.detail.value,
        } as StorageEvent);
      }
    };

    const handleAuthChange = (e: CustomEvent) => {
      if (!e.detail?.isAuthenticated) {
        try {
          isUpdatingFromStorage.current = true;
          setUserState(null);
          setUserIdState(null);
          hasFetchedProfile.current = false;
        } finally {
          isUpdatingFromStorage.current = false;
        }
      }
    };

    window.addEventListener("storage", handleStorageChange);
    window.addEventListener(
      "localStorageChange",
      handleCustomStorageChange as EventListener,
    );
    window.addEventListener("authChange", handleAuthChange as EventListener);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
      window.removeEventListener(
        "localStorageChange",
        handleCustomStorageChange as EventListener,
      );
      window.removeEventListener("authChange", handleAuthChange as EventListener);
    };
  }, []);

  useEffect(() => {
    if (isUpdatingFromStorage.current) return;

    try {
      isUpdatingToStorage.current = true;
      saveUserToStorage(user, userId);
    } finally {
      isUpdatingToStorage.current = false;
    }
  }, [user, userId]);

  useEffect(() => {
    if (user?.id) {
      setUserIdState(user.id);
    }
  }, [user?.id]);

  useEffect(() => {
    if (!isAuthenticated && user) {
      setUserState(null);
      setUserIdState(null);
      hasFetchedProfile.current = false;
    }
  }, [isAuthenticated, user]);

  const setUser = useCallback((u: AdminUser | null) => {
    setUserState(u);
  }, []);

  const clearUser = useCallback(() => {
    setUserState(null);
    setUserIdState(null);
    hasFetchedProfile.current = false;
  }, []);

  const fetchProfile = useCallback(async (): Promise<AdminUser | null> => {
    setIsLoading(true);
    try {
      const res = await apiGet(ADMIN_PROFILE_ENDPOINTS.ME);

      if (!res.ok) {
        if (res.status === 401) {
          return null;
        }
        throw new Error("Failed to fetch profile");
      }

      const data = await res.json();
      const u = data?.data ?? data?.user ?? data;

      setUserState(u);
      hasFetchedProfile.current = true;

      return u;
    } catch (error) {
      console.error("Failed to fetch profile:", error);
      return null;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const refetchProfile = useCallback(async (): Promise<AdminUser | null> => {
    if (!isAuthenticated) {
      return null;
    }

    try {
      const res = await apiGet(ADMIN_PROFILE_ENDPOINTS.ME);

      if (!res.ok) {
        if (res.status === 401) {
          return null;
        }
        throw new Error("Failed to fetch profile");
      }

      const data = await res.json();
      const u = data?.data ?? data?.user ?? data;

      setUserState(u);

      return u;
    } catch (error) {
      console.error("Failed to refetch profile:", error);
      return null;
    }
  }, [isAuthenticated]);

  useEffect(() => {
    if (
      isAuthenticated &&
      (!user || !user.id) &&
      !hasFetchedProfile.current &&
      !isLoading
    ) {
      hasFetchedProfile.current = true;
      fetchProfile();
    }
  }, [isAuthenticated, user, isLoading, fetchProfile]);

  return (
    <UserContext.Provider
      value={{
        user,
        userId,
        setUser,
        clearUser,
        fetchProfile,
        refetchProfile,
        isLoading,
      }}
    >
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  const ctx = useContext(UserContext);
  if (!ctx) throw new Error("useUser must be used within UserProvider");
  return ctx;
}

export default UserProvider;
