"use client";

import { useAuth as useAuthStore } from "@/lib/stores/auth.store";
import { useUser } from "@/lib/stores/user.store";
import { useNavigate } from "react-router";
import { toast } from "sonner";

export function useAuth() {
  const auth = useAuthStore();
  const { clearUser } = useUser();
  const navigate = useNavigate();

  const loginAndRedirect = async (
    identifier: string,
    password: string,
    redirectTo = "/",
  ) => {
    try {
      await auth.login(identifier, password);
      toast.success("Logged in successfully");
      navigate(redirectTo);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Login failed";
      toast.error(message);
      throw error;
    }
  };

  const logoutAndRedirect = async (redirectTo = "/login") => {
    try {
      await auth.logout();
      clearUser();
      toast.success("Logged out successfully");
      navigate(redirectTo);
    } catch (error) {
      console.error("Logout error:", error);
      clearUser();
      navigate(redirectTo);
    }
  };

  const requireAuth = (redirectTo = "/login") => {
    if (!auth.isAuthenticated) {
      toast.error("Please login to continue");
      navigate(redirectTo);
      return false;
    }
    return true;
  };

  return {
    ...auth,
    loginAndRedirect,
    logoutAndRedirect,
    requireAuth,
  };
}
