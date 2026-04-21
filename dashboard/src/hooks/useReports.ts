"use client";

import { reportsService } from "@/lib/services/reports.service";
import type { Report } from "@/lib/types/reports.types";
import { useCallback, useEffect, useRef, useState } from "react";

interface ReportsFilters {
  status: string;
}

export function useReports() {
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [currentLimit, setCurrentLimit] = useState(20);
  const [filters, setFiltersState] = useState<ReportsFilters>({ status: "" });

  const stateRef = useRef({ currentPage, currentLimit, filters });
  useEffect(() => {
    stateRef.current = { currentPage, currentLimit, filters };
  });

  const fetchReports = useCallback(
    async (
      page: number,
      limit: number,
      f: ReportsFilters,
      isRefresh = false,
    ) => {
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);
        const res = await reportsService.list({
          page,
          limit,
          status: f.status || undefined,
        });
        setReports(res.data);
        setCurrentPage(res.page);
        setTotalPages(res.total_pages);
        setTotalCount(res.total);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load reports");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [],
  );

  useEffect(() => {
    fetchReports(1, stateRef.current.currentLimit, filters);
    setCurrentPage(1);
  }, [fetchReports, filters]);

  const goToPage = useCallback(
    (page: number) => {
      setCurrentPage(page);
      fetchReports(
        page,
        stateRef.current.currentLimit,
        stateRef.current.filters,
      );
    },
    [fetchReports],
  );

  const setPageSize = useCallback(
    (limit: number) => {
      setCurrentLimit(limit);
      setCurrentPage(1);
      fetchReports(1, limit, stateRef.current.filters);
    },
    [fetchReports],
  );

  const refresh = useCallback(() => {
    fetchReports(
      stateRef.current.currentPage,
      stateRef.current.currentLimit,
      stateRef.current.filters,
      true,
    );
  }, [fetchReports]);

  const setFilter = useCallback((key: keyof ReportsFilters, value: string) => {
    setFiltersState((prev) => ({ ...prev, [key]: value }));
  }, []);

  const updateReportInList = useCallback((updatedReport: Report) => {
    setReports((prev) =>
      prev.map((r) => (r.id === updatedReport.id ? updatedReport : r)),
    );
  }, []);

  return {
    reports,
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
    updateReportInList,
  };
}
