"use client";

import { AuthProvider, useAuth } from "./auth.store";
import { RealtimeProvider, useRealtime } from "./realtime.store";
import { UserProvider, useUser } from "./user.store";

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <UserProvider>
        <RealtimeProvider>{children}</RealtimeProvider>
      </UserProvider>
    </AuthProvider>
  );
}

export { AuthProvider, UserProvider, RealtimeProvider };

export { useAuth, useUser, useRealtime };

export function useApp() {
  const auth = useAuth();
  const user = useUser();
  const realtime = useRealtime();

  return {
    ...auth,

    user: user.user,
    setUser: user.setUser,
    clearUser: user.clearUser,
    fetchProfile: user.fetchProfile,
    isLoadingUser: user.isLoading,

    isRealtimeConnected: realtime.isConnected,
    sendRealtime: realtime.send,
    lastRealtimeMessage: realtime.lastMessage,
  };
}
