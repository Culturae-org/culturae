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

const HEARTBEAT_TIMEOUT_MS = 90_000;

const getRealtimeUrl = () => {
  if (typeof window === "undefined") return "";

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const host = window.location.host;

  const isAdmin = window.location.pathname.startsWith("/console");
  const endpoint = isAdmin ? "/api/v1/admin/realtime" : "/api/v1/realtime";

  if (process.env.NODE_ENV === "production") {
    return `${protocol}//${host}${endpoint}`;
  }
  return `ws://localhost:8080${endpoint}`;
};

export function RealtimeProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  const { userId } = useUser();

  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<MessageEvent | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const heartbeatTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const sendQueueRef = useRef<string[]>([]);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 10;

  const resetHeartbeatTimeout = useCallback(() => {
    if (heartbeatTimeoutRef.current) clearTimeout(heartbeatTimeoutRef.current);
    heartbeatTimeoutRef.current = setTimeout(() => {
      wsRef.current?.close(4000, "heartbeat timeout");
    }, HEARTBEAT_TIMEOUT_MS);
  }, []);

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
      const ws = new WebSocket(getRealtimeUrl());

      ws.onopen = () => {
        setIsConnected(true);
        reconnectAttemptsRef.current = 0;
        resetHeartbeatTimeout();

        const queued = sendQueueRef.current.splice(0);
        for (const msg of queued) {
          try { ws.send(msg); } catch { /* ignore */ }
        }

        window.dispatchEvent(
          new CustomEvent("websocket-reconnected", { detail: { timestamp: Date.now() } })
        );
      };

      ws.onclose = (event) => {
        setIsConnected(false);
        wsRef.current = null;
        if (heartbeatTimeoutRef.current) {
          clearTimeout(heartbeatTimeoutRef.current);
          heartbeatTimeoutRef.current = null;
        }

        if (
          event.code !== 1000 &&
          event.code !== 1001 &&
          reconnectAttemptsRef.current < maxReconnectAttempts
        ) {
          const baseDelay = 100;
          const delay = Math.min(
            baseDelay * (2 ** reconnectAttemptsRef.current - 1) + baseDelay,
            3000,
          );

          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttemptsRef.current++;
            connect();
          }, delay);
        } else if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
          console.error("Max reconnection attempts reached");
          window.dispatchEvent(
            new CustomEvent("websocket-reconnect-failed", { detail: { attempts: maxReconnectAttempts } })
          );
        }
      };

      ws.onerror = (error) => {
        console.error("Realtime connection error:", error);
      };

      ws.onmessage = (event) => {
        setLastMessage(event);
        resetHeartbeatTimeout();

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
    if (heartbeatTimeoutRef.current) {
      clearTimeout(heartbeatTimeoutRef.current);
      heartbeatTimeoutRef.current = null;
    }

    sendQueueRef.current = [];

    if (wsRef.current) {
      wsRef.current.close(1000, "User disconnected");
      wsRef.current = null;
    }

    setIsConnected(false);
    reconnectAttemptsRef.current = 0;
  }, []);

  const send = useCallback((data: unknown) => {
    const message = typeof data === "string" ? data : JSON.stringify(data);
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      try {
        wsRef.current.send(message);
      } catch (error) {
        console.error("Failed to send realtime message:", error);
      }
    } else {
      sendQueueRef.current.push(message);
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

    const handleOnline = () => {
      reconnectAttemptsRef.current = 0;
      connect();
    };

    window.addEventListener("beforeunload", handleBeforeUnload);
    window.addEventListener("online", handleOnline);

    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
      window.removeEventListener("online", handleOnline);
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (heartbeatTimeoutRef.current) {
        clearTimeout(heartbeatTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close(1000, "Component unmounted");
      }
    };
  }, [connect]);

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
