"use client";

import { useAuth } from "@/lib/stores/auth.store";
import { useUser } from "@/lib/stores/user.store";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";

interface AuthGuardProps {
  children: React.ReactNode;
  requireAuth?: boolean;
}

export function AuthGuard({ children, requireAuth = true }: AuthGuardProps) {
  const { isAuthenticated, isLoading: isAuthLoading } = useAuth();
  const { user, isLoading: isUserLoading } = useUser();
  const navigate = useNavigate();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!mounted || isAuthLoading) return;

    if (requireAuth && !isAuthenticated) {
      navigate("/login", { replace: true });
    } else if (!requireAuth && isAuthenticated) {
      navigate("/", { replace: true });
    }
  }, [mounted, isAuthLoading, isAuthenticated, requireAuth, navigate]);

  if (!mounted || isAuthLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto mb-4" />
          <p className="text-sm text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (requireAuth && !isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-sm text-muted-foreground">
            Redirecting to login...
          </p>
        </div>
      </div>
    );
  }

  if (requireAuth && isUserLoading && !user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto mb-4" />
          <p className="text-sm text-muted-foreground">Loading user data...</p>
        </div>
      </div>
    );
  }

  return children;
}
