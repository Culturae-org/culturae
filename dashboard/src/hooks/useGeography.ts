"use client";

import geographyService from "@/lib/services/geography.service";
import type {
  Continent,
  Country,
  GeographyDatasetStats,
  Region,
} from "@/lib/types/geography.types";
import { useCallback, useEffect, useRef, useState } from "react";

export function useGeographyData(datasetId: string | undefined, pageSize = 20) {
  const [countries, setCountries] = useState<Country[]>([]);
  const [continents, setContinents] = useState<Continent[]>([]);
  const [regions, setRegions] = useState<Region[]>([]);
  const [stats, setStats] = useState<GeographyDatasetStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totalCountries, setTotalCountries] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [limit, setLimit] = useState(pageSize);

  useEffect(() => {
    setLimit(pageSize);
  }, [pageSize]);

  const fetchCountries = useCallback(
    async (page = 1, searchQuery?: string) => {
      if (!datasetId) return;

      setLoading(true);
      setError(null);
      try {
        const response = searchQuery
          ? await geographyService.searchCountries(datasetId, searchQuery, {
              page,
              limit,
            })
          : await geographyService.getCountries(datasetId, {
              page,
              limit,
            });
        setCountries(response.data);
        setTotalCountries(response.total);
        setCurrentPage(page);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to fetch countries",
        );
      } finally {
        setLoading(false);
      }
    },
    [datasetId, limit],
  );

  const fetchContinents = useCallback(async () => {
    if (!datasetId) return;

    try {
      const data = await geographyService.getContinents(datasetId);
      setContinents(data);
    } catch (err) {
      console.error("Failed to fetch continents:", err);
    }
  }, [datasetId]);

  const fetchRegions = useCallback(async () => {
    if (!datasetId) return;

    try {
      const data = await geographyService.getRegions(datasetId);
      setRegions(data);
    } catch (err) {
      console.error("Failed to fetch regions:", err);
    }
  }, [datasetId]);

  const fetchStats = useCallback(async () => {
    if (!datasetId) return;

    try {
      const data = await geographyService.getDatasetStats(datasetId);
      setStats(data);
    } catch (err) {
      console.error("Failed to fetch stats:", err);
    }
  }, [datasetId]);

  const fetchCountriesByContinent = useCallback(
    async (continent: string) => {
      if (!datasetId) return [];

      try {
        return await geographyService.getCountriesByContinent(
          datasetId,
          continent,
        );
      } catch (err) {
        console.error("Failed to fetch countries by continent:", err);
        return [];
      }
    },
    [datasetId],
  );

  const fetchCountriesByRegion = useCallback(
    async (region: string) => {
      if (!datasetId) return [];

      try {
        return await geographyService.getCountriesByRegion(datasetId, region);
      } catch (err) {
        console.error("Failed to fetch countries by region:", err);
        return [];
      }
    },
    [datasetId],
  );

  useEffect(() => {
    if (datasetId) {
      fetchCountries();
      fetchContinents();
      fetchRegions();
      fetchStats();
    }
  }, [datasetId, fetchCountries, fetchContinents, fetchRegions, fetchStats]);

  return {
    countries,
    continents,
    regions,
    stats,
    loading,
    error,
    totalCountries,
    currentPage,
    totalPages: Math.ceil(totalCountries / limit),
    fetchCountries,
    fetchContinents,
    fetchRegions,
    fetchStats,
    fetchCountriesByContinent,
    fetchCountriesByRegion,
  };
}

export function useContinents(datasetId: string) {
  const [continents, setContinents] = useState<Continent[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(
    async (isRefresh = false) => {
      if (!datasetId) return;
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);
        const data = await geographyService.getContinents(datasetId);
        setContinents(data);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load continents",
        );
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [datasetId],
  );

  useEffect(() => {
    fetch();
  }, [fetch]);

  const refresh = useCallback(() => fetch(true), [fetch]);

  return { continents, loading, refreshing, error, refresh };
}

export function useRegions(datasetId: string) {
  const [regions, setRegions] = useState<Region[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(
    async (isRefresh = false) => {
      if (!datasetId) return;
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);
        const data = await geographyService.getRegions(datasetId);
        setRegions(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load regions");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [datasetId],
  );

  useEffect(() => {
    fetch();
  }, [fetch]);

  const refresh = useCallback(() => fetch(true), [fetch]);

  return { regions, loading, refreshing, error, refresh };
}

export interface CountryFilters {
  continent: string;
  region: string;
  populationMin: string;
  populationMax: string;
  areaMin: string;
  areaMax: string;
  independent: string;
  drivingSide: string;
}

const EMPTY_COUNTRY_FILTERS: CountryFilters = {
  continent: "",
  region: "",
  populationMin: "",
  populationMax: "",
  areaMin: "",
  areaMax: "",
  independent: "",
  drivingSide: "",
};

export function useCountriesList(datasetId: string) {
  const [countries, setCountries] = useState<Country[]>([]);
  const [continents, setContinents] = useState<Continent[]>([]);
  const [regions, setRegions] = useState<Region[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [currentLimit, setCurrentLimit] = useState(30);
  const [search, setSearchState] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [filters, setFiltersState] = useState<CountryFilters>(
    EMPTY_COUNTRY_FILTERS,
  );

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
    const t = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(t);
  }, [search]);

  const fetchCountries = useCallback(
    async (
      page: number,
      limit: number,
      q: string,
      f: CountryFilters,
      isRefresh = false,
    ) => {
      if (!datasetId) return;
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);
        const res = await geographyService.listCountries(datasetId, {
          page,
          limit,
          q: q.trim() || undefined,
          continent: f.continent || undefined,
          region: f.region || undefined,
          population_min: f.populationMin || undefined,
          population_max: f.populationMax || undefined,
          area_min: f.areaMin || undefined,
          area_max: f.areaMax || undefined,
          independent: f.independent || undefined,
          driving_side: f.drivingSide || undefined,
        });
        setCountries(res.data);
        setCurrentPage(res.page);
        setTotalPages(res.total_pages);
        setTotalCount(res.total);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load countries",
        );
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [datasetId],
  );

  const fetchFilterOptions = useCallback(async () => {
    if (!datasetId) return;
    try {
      const [c, r] = await Promise.all([
        geographyService.getContinents(datasetId),
        geographyService.getRegions(datasetId),
      ]);
      setContinents(c);
      setRegions(r);
    } catch (err) {
      console.error("Failed to fetch filter options:", err);
    }
  }, [datasetId]);

  useEffect(() => {
    fetchCountries(1, stateRef.current.currentLimit, debouncedSearch, filters);
    setCurrentPage(1);
  }, [fetchCountries, debouncedSearch, filters]);

  useEffect(() => {
    fetchFilterOptions();
  }, [fetchFilterOptions]);

  const goToPage = useCallback(
    (page: number) => {
      setCurrentPage(page);
      fetchCountries(
        page,
        stateRef.current.currentLimit,
        stateRef.current.debouncedSearch,
        stateRef.current.filters,
      );
    },
    [fetchCountries],
  );

  const setPageSize = useCallback(
    (limit: number) => {
      setCurrentLimit(limit);
      setCurrentPage(1);
      fetchCountries(
        1,
        limit,
        stateRef.current.debouncedSearch,
        stateRef.current.filters,
      );
    },
    [fetchCountries],
  );

  const refresh = useCallback(() => {
    const {
      currentPage: p,
      currentLimit: l,
      debouncedSearch: q,
      filters: f,
    } = stateRef.current;
    fetchCountries(p, l, q, f, true);
  }, [fetchCountries]);

  const setFilter = useCallback((key: keyof CountryFilters, value: string) => {
    setFiltersState((prev) => {
      const next = { ...prev, [key]: value };
      if (key === "continent" && value !== prev.continent) next.region = "";
      return next;
    });
  }, []);

  const setSearch = useCallback((q: string) => setSearchState(q), []);

  const clearFilters = useCallback(
    () => setFiltersState(EMPTY_COUNTRY_FILTERS),
    [],
  );

  return {
    countries,
    continents,
    regions,
    loading,
    refreshing,
    error,
    currentPage,
    totalPages,
    totalCount,
    currentLimit,
    filters,
    search,
    goToPage,
    setPageSize,
    refresh,
    setFilter,
    setSearch,
    clearFilters,
  };
}
