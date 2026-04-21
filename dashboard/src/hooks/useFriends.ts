"use client";

import { friendsService } from "@/lib/services/friends.service";
import type { FriendRequest, Friendship } from "@/lib/types/friends.types";
import { useCallback, useEffect, useState } from "react";

export function useUserFriends(userId: string, page: number, limit: number) {
  const [friends, setFriends] = useState<Friendship[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchFriends = useCallback(
    async (isRefresh = false) => {
      if (isRefresh) setRefreshing(true);
      else setLoading(true);
      setError(null);
      try {
        const res = await friendsService.getFriendsForUser(userId, page, limit);
        setFriends(res.data);
        setTotal(res.total);
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Failed to load friends";
        setError(msg);
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [userId, page, limit],
  );

  useEffect(() => {
    fetchFriends();
  }, [fetchFriends]);

  const refresh = useCallback(() => fetchFriends(true), [fetchFriends]);

  return { friends, total, loading, refreshing, error, refresh };
}

export function useUserFriendRequests(
  userId: string,
  direction: string,
  status: string,
  page: number,
  limit: number,
) {
  const [requests, setRequests] = useState<FriendRequest[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchRequests = useCallback(
    async (isRefresh = false) => {
      if (isRefresh) setRefreshing(true);
      else setLoading(true);
      setError(null);
      try {
        const res = await friendsService.getFriendRequestsForUser(userId, {
          direction: direction || undefined,
          status: status || undefined,
          page,
          limit,
        });
        setRequests(res.data);
        setTotal(res.total);
      } catch (e: unknown) {
        const msg =
          e instanceof Error ? e.message : "Failed to load friend requests";
        setError(msg);
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [userId, direction, status, page, limit],
  );

  useEffect(() => {
    fetchRequests();
  }, [fetchRequests]);

  const refresh = useCallback(() => fetchRequests(true), [fetchRequests]);

  return { requests, total, loading, refreshing, error, refresh };
}
