"use client";

import { importsService } from "@/lib/services/imports.service";
import type { ImportJob } from "@/lib/types/datasets.types";
import { useCallback, useState } from "react";
import { toast } from "sonner";

interface ImportsQueryParams {
  page?: number;
  limit?: number;
  offset?: number;
}

export function useImports() {
  const [imports, setImports] = useState<ImportJob[]>([]);
  const [currentImport, setCurrentImport] = useState<ImportJob | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);

  const fetchImports = useCallback(async (params?: ImportsQueryParams) => {
    setLoading(true);
    setError(null);

    try {
      const result = await importsService.getImports(params);
      setImports(result.data);
      setCurrentPage(result.page);
      setTotalPages(result.total_pages);
      setTotalCount(result.total);
      return result;
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch imports";
      setError(message);
      toast.error(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchImportById = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);

    try {
      const importJob = await importsService.getImportById(id);
      setCurrentImport(importJob);
      return importJob;
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch import";
      setError(message);
      toast.error(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    imports,
    currentImport,
    loading,
    error,
    currentPage,
    totalPages,
    totalCount,

    fetchImports,
    fetchImportById,

    setImports,
    setCurrentImport,
    clearError: () => setError(null),
  };
}
