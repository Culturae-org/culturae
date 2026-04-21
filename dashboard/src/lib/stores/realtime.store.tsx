"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { useAuth } from "./auth.store";
import { useUser } from "./user.store";

type RealtimeContextValue = {
  isConnected: boolean;
  lastMessage: MessageEvent | null;
  send: (data: unknown) => void;
  connect: () => void;
  disconnect: () => void;
};

const RealtimeContext = createContext<RealtimeContextValue | undefined>(
  undefined,
);

const REALTIME_URL =
  typeof window !== "undefined"
    ? process.env.NODE_ENV === "production"
      ? `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/api/v1/realtime`
      : "ws://localhost:8080/api/v1/realtime"
    : "";

export function RealtimeProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  const { userId } = useUser();

  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<MessageEvent | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;

  const connect = useCallback(() => {
    if (!isAuthenticated || !userId) {
      return;
    }

    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    if (wsRef.current?.readyState === WebSocket.CONNECTING) {
      return;
    }

    try {
      const url = `${REALTIME_URL}`;
      const ws = new WebSocket(url);

      ws.onopen = () => {
        setIsConnected(true);
        reconnectAttemptsRef.current = 0;
      };

      ws.onclose = (event) => {
        setIsConnected(false);
        wsRef.current = null;

        if (
          event.code !== 1000 &&
          reconnectAttemptsRef.current < maxReconnectAttempts
        ) {
          const delay = Math.min(
            1000 * 2 ** reconnectAttemptsRef.current,
            30000,
          );

          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttemptsRef.current++;
            connect();
          }, delay);
        }
      };

      ws.onerror = (error) => {
        console.error("Realtime connection error:", error);
      };

      ws.onmessage = (event) => {
        setLastMessage(event);

        try {
          const data = JSON.parse(event.data);

          window.dispatchEvent(
            new CustomEvent("realtime-message", { detail: data }),
          );

          if (data && data.type === "presence" && data.user_id) {
            window.dispatchEvent(
              new CustomEvent("user-presence-change", {
                detail: { user_id: data.user_id, is_online: !!data.is_online },
              }),
            );
          }
        } catch (error) {
          console.error("Failed to parse realtime message:", error);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.error("Failed to create realtime connection:", error);
    }
  }, [isAuthenticated, userId]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close(1000, "User disconnected");
      wsRef.current = null;
    }

    setIsConnected(false);
    reconnectAttemptsRef.current = 0;
  }, []);

  const send = useCallback((data: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      try {
        const message = typeof data === "string" ? data : JSON.stringify(data);
        wsRef.current.send(message);
      } catch (error) {
        console.error("Failed to send realtime message:", error);
      }
    } else {
      console.warn("Cannot send message: realtime connection not established");
    }
  }, []);

  useEffect(() => {
    if (isAuthenticated && userId) {
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [isAuthenticated, userId, connect, disconnect]);

  useEffect(() => {
    const handleBeforeUnload = () => {
      if (wsRef.current) {
        wsRef.current.close(1000, "Page unloading");
      }
    };

    window.addEventListener("beforeunload", handleBeforeUnload);

    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close(1000, "Component unmounted");
      }
    };
  }, []);

  return (
    <RealtimeContext.Provider
      value={{
        isConnected,
        lastMessage,
        send,
        connect,
        disconnect,
      }}
    >
      {children}
    </RealtimeContext.Provider>
  );
}

export function useRealtime() {
  const ctx = useContext(RealtimeContext);
  if (!ctx) throw new Error("useRealtime must be used within RealtimeProvider");
  return ctx;
}

export default RealtimeProvider;
