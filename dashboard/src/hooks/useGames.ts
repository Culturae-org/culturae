"use client";

import { gamesService } from "@/lib/services/games.service";
import type {
  AdminGame,
  Game,
} from "@/lib/types/games.types";
import { useCallback, useEffect, useRef, useState } from "react";

export type EnrichedGame = Game & { _user_score: number; _is_winner: boolean };

interface GamesFilters {
  mode: string;
  status: string;
  search: string;
  archived: string;
}

export function useGames() {
  const [games, setGames] = useState<AdminGame[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [currentLimit, setCurrentLimit] = useState(10);
  const [filters, setFiltersState] = useState<GamesFilters>({
    mode: "",
    status: "",
    search: "",
    archived: "",
  });
  const [debouncedSearch, setDebouncedSearch] = useState("");

  const stateRef = useRef({
    currentPage,
    currentLimit,
    filters,
    debouncedSearch,
  });
  useEffect(() => {
    stateRef.current = { currentPage, currentLimit, filters, debouncedSearch };
  });

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(filters.search), 300);
    return () => clearTimeout(timer);
  }, [filters.search]);

  const fetchGames = useCallback(
    async (
      page: number,
      limit: number,
      f: GamesFilters,
      ds: string,
      isRefresh = false,
    ) => {
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);
        const res = await gamesService.list({
          page,
          limit,
          mode: f.mode || undefined,
          status: f.status || undefined,
          query: ds.trim() || undefined,
          archived: f.archived || undefined,
        });
        setGames(res.data);
        setCurrentPage(res.page);
        setTotalPages(res.total_pages);
        setTotalCount(res.total);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load games");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [],
  );

  useEffect(() => {
    fetchGames(
      1,
      stateRef.current.currentLimit,
      stateRef.current.filters,
      stateRef.current.debouncedSearch,
    );
    setCurrentPage(1);
  }, [fetchGames]);

  useEffect(() => {
    fetchGames(1, stateRef.current.currentLimit, filters, debouncedSearch);
    setCurrentPage(1);
  }, [filters.mode, filters.status, filters.archived, debouncedSearch]);

  const goToPage = useCallback(
    (page: number) => {
      setCurrentPage(page);
      fetchGames(
        page,
        stateRef.current.currentLimit,
        stateRef.current.filters,
        stateRef.current.debouncedSearch,
      );
    },
    [fetchGames],
  );

  const setPageSize = useCallback(
    (limit: number) => {
      setCurrentLimit(limit);
      setCurrentPage(1);
      fetchGames(
        1,
        limit,
        stateRef.current.filters,
        stateRef.current.debouncedSearch,
      );
    },
    [fetchGames],
  );

  const refresh = useCallback(() => {
    const { currentPage, currentLimit, filters, debouncedSearch } =
      stateRef.current;
    fetchGames(currentPage, currentLimit, filters, debouncedSearch, true);
  }, [fetchGames]);

  const setFilter = useCallback((key: keyof GamesFilters, value: string) => {
    setFiltersState((prev) => ({ ...prev, [key]: value }));
  }, []);

  const clearFilters = useCallback(() => {
    setFiltersState({ mode: "", status: "", search: "", archived: "" });
  }, []);

  const updateGameInList = useCallback((updatedGame: AdminGame) => {
    setGames((prev) =>
      prev.map((g) => (g.id === updatedGame.id ? updatedGame : g)),
    );
  }, []);

  const removeGame = useCallback((gameId: string) => {
    setGames((prev) => prev.filter((g) => g.id !== gameId));
    setTotalCount((c) => Math.max(0, c - 1));
  }, []);

  return {
    games,
    setGames,
    loading,
    refreshing,
    error,
    currentPage,
    totalPages,
    totalCount,
    currentLimit,
    filters,
    goToPage,
    setPageSize,
    refresh,
    setFilter,
    clearFilters,
    updateGameInList,
    removeGame,
  };
}

interface UserGamesFilters {
  status: string;
  mode: string;
}

export function useUserGames(userId: string) {
  const [games, setGames] = useState<EnrichedGame[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFiltersState] = useState<UserGamesFilters>({
    status: "",
    mode: "",
  });

  const fetchGames = useCallback(
    async (isRefresh = false) => {
      if (!userId) return;
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);
        const raw = await gamesService.getUserGameHistory(
          userId,
          filters.status || undefined,
          filters.mode || undefined,
        );
        const enriched: EnrichedGame[] = raw.map((item) => ({
          ...item.game,
          players: item.players ?? item.game?.players ?? [],
          _user_score: item.user_score ?? 0,
          _is_winner: item.is_winner ?? false,
        }));
        setGames(
          enriched.sort(
            (a, b) =>
              new Date(b.created_at).getTime() -
              new Date(a.created_at).getTime(),
          ),
        );
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load games");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [userId, filters.status, filters.mode],
  );

  useEffect(() => {
    if (userId) {
      fetchGames();
    }
  }, [userId, fetchGames]);

  const refresh = useCallback(() => {
    if (userId) {
      fetchGames(true);
    }
  }, [userId, fetchGames]);

  const setFilter = useCallback((key: keyof UserGamesFilters, value: string) => {
    setFiltersState((prev) => ({ ...prev, [key]: value }));
  }, []);

  const clearFilters = useCallback(() => {
    setFiltersState({ status: "", mode: "" });
  }, []);

  return {
    games,
    loading,
    refreshing,
    error,
    filters,
    refresh,
    setFilter,
    clearFilters,
  };
}
