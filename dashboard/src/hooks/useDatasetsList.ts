"use client";

import { datasetsService } from "@/lib/services/datasets.service";
import { geographyService } from "@/lib/services/geography.service";
import type { DatasetUpdateInfo } from "@/lib/types/datasets.types";
import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

export interface UnifiedDataset {
  id: string;
  type: "questions" | "geography";
  name: string;
  slug: string;
  description?: string;
  version: string;
  source: string;
  manifest_url?: string;
  is_active: boolean;
  is_default: boolean;
  imported_at: string;
  created_at: string;
  updated_at: string;
  import_job_id?: string;
  question_count?: number;
  theme_count?: number;
  country_count?: number;
  continent_count?: number;
  region_count?: number;
  flag_count?: number;
  flag_png512_count?: number;
  flag_png1024_count?: number;
  latest_available_version?: string;
}

interface DatasetsListFilters {
  type: "all" | "questions" | "geography";
  source: string;
  status: "all" | "default";
}

const _CULTPEDIA_QUESTIONS_MANIFEST_URL =
  "https://raw.githubusercontent.com/Culturae-org/cultpedia/refs/heads/main/datasets/general-knowledge/manifest.json";
const _CULTPEDIA_GEOGRAPHY_MANIFEST_URL =
  "https://raw.githubusercontent.com/Culturae-org/cultpedia/refs/heads/main/datasets/geography/manifest.json";

export function useDatasetsList() {
  const [datasets, setDatasets] = useState<UnifiedDataset[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [filters, setFilters] = useState<DatasetsListFilters>({
    type: "all",
    source: "",
    status: "all",
  });
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");

  const [checkingUpdate, setCheckingUpdate] = useState(false);

  const [currentPage, setCurrentPage] = useState(1);
  const [currentLimit, setCurrentLimit] = useState(10);
  const [totalCount, setTotalCount] = useState(0);
  const [totalPages, setTotalPages] = useState(1);

  const stateRef = useRef({
    filters,
    debouncedSearch,
    currentPage,
    currentLimit,
  });
  useEffect(() => {
    stateRef.current = { filters, debouncedSearch, currentPage, currentLimit };
  });

  const fetchDatasets = useCallback(
    async (
      page: number,
      limit: number,
      f: DatasetsListFilters,
      isRefresh = false,
    ) => {
      try {
        isRefresh ? setRefreshing(true) : setLoading(true);
        setError(null);

        const defaultOnly = f.status === "default";
        const offset = (page - 1) * limit;
        const datasetType = f.type === "all" ? undefined : f.type;
        const response = await datasetsService.getDatasetsPaginated({
          limit,
          offset,
          default_only: defaultOnly,
          dataset_type: datasetType,
        });

        setDatasets(response.data || []);
        setTotalCount(response.total || 0);
        setTotalPages(response.total_pages || 1);
      } catch (err) {
        setError(err instanceof Error ? err.message : "An error occurred");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [],
  );

  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(t);
  }, [search]);

  useEffect(() => {
    fetchDatasets(1, stateRef.current.currentLimit, filters);
    setCurrentPage(1);
  }, [fetchDatasets, filters]);

  const refresh = useCallback(() => {
    const { currentPage, currentLimit, filters } = stateRef.current;
    fetchDatasets(currentPage, currentLimit, filters, true);
  }, [fetchDatasets]);

  const goToPage = useCallback(
    (page: number) => {
      const clamped = Math.max(1, page);
      setCurrentPage(clamped);
      fetchDatasets(
        clamped,
        stateRef.current.currentLimit,
        stateRef.current.filters,
      );
    },
    [fetchDatasets],
  );

  const setPageSize = useCallback(
    (size: number) => {
      setCurrentLimit(size);
      setCurrentPage(1);
      fetchDatasets(1, size, stateRef.current.filters);
    },
    [fetchDatasets],
  );

  const setFilter = useCallback(
    <K extends keyof DatasetsListFilters>(
      key: K,
      value: DatasetsListFilters[K],
    ) => {
      setFilters((prev) => ({ ...prev, [key]: value }));
    },
    [],
  );

  const clearFilters = useCallback(() => {
    setFilters({ type: "all", source: "", status: "all" });
    setSearch("");
  }, []);

  const createDataset = useCallback(
    async (data: {
      name: string;
      slug: string;
      description?: string;
      version: string;
      source?: string;
    }) => {
      try {
        const newDataset = await datasetsService.createDataset({
          name: data.name,
          slug: data.slug,
          description: data.description,
          version: data.version,
          source: (data.source || "custom") as
            | "cultpedia"
            | "custom"
            | "imported",
        });
        toast.success(`Dataset "${data.name}" created successfully`);
        refresh();
        return newDataset;
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to create dataset";
        toast.error(msg);
        throw err;
      }
    },
    [refresh],
  );

  const updateDataset = useCallback(
    async (id: string, updates: { is_active?: boolean }) => {
      try {
        const dataset = datasets.find((d) => d.id === id);
        if (!dataset) return;

        if (dataset.type === "questions") {
          await datasetsService.updateDataset(id, updates);
        } else {
          await geographyService.getDataset(id);
        }
        refresh();
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to update dataset";
        toast.error(msg);
        throw err;
      }
    },
    [datasets, refresh],
  );

  const toggleActive = useCallback(
    async (id: string) => {
      const dataset = datasets.find((d) => d.id === id);
      if (!dataset) return;

      const newActive = !dataset.is_active;

      try {
        if (dataset.type === "questions") {
          await datasetsService.updateDataset(id, { is_active: newActive });
        } else {
          await geographyService.getDataset(id);
        }

        setDatasets((prev) =>
          prev.map((d) => (d.id === id ? { ...d, is_active: newActive } : d)),
        );
        toast.success(
          `Dataset ${newActive ? "activated" : "deactivated"} successfully`,
        );
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to toggle dataset";
        toast.error(
          `Failed to ${newActive ? "activate" : "deactivate"} dataset. ${msg}`,
        );
        throw err;
      }
    },
    [datasets],
  );

  const deleteDataset = useCallback(
    async (id: string, force = false) => {
      try {
        const dataset = datasets.find((d) => d.id === id);
        if (!dataset) return;

        if (dataset.type === "questions") {
          await datasetsService.deleteDataset(id, force);
        } else {
          await geographyService.deleteDataset(id, force);
        }

        toast.success(`Dataset "${dataset.name}" deleted successfully`);
        refresh();
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to delete dataset";

        let userMessage = message;
        if (message.includes("cannot delete the only dataset of this type")) {
          userMessage = `You must have at least 2 datasets of the same type to delete one. This is the only ${datasets.find((d) => d.id === id)?.type} dataset.`;
        } else if (message.includes("cannot delete the default dataset")) {
          userMessage =
            "This dataset is currently set as default. Please set another dataset as default first.";
        }

        toast.error(userMessage);
        throw err;
      }
    },
    [datasets, refresh],
  );

  const setDefault = useCallback(
    async (id: string) => {
      try {
        const dataset = datasets.find((d) => d.id === id);
        if (!dataset) return;

        if (dataset.type === "questions") {
          await datasetsService.setDefaultDataset(id);
        } else {
          await geographyService.setDefaultDataset(id);
        }

        setDatasets((prev) =>
          prev.map((d) => ({
            ...d,
            is_default: d.type === dataset.type && d.id === id,
          })),
        );
        toast.success(
          `${dataset.type === "questions" ? "Questions" : "Geography"} dataset set as default successfully`,
        );
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to set default";
        toast.error(msg);
        throw err;
      }
    },
    [datasets],
  );

  const checkAllUpdates = useCallback(async () => {
    setCheckingUpdate(true);
    try {
      const updates = await datasetsService.checkAllUpdates();

      const datasetsWithUpdates = updates.filter((d) => d.has_update);

      if (datasetsWithUpdates.length > 0) {
        const msg = "New versions available: ";
        const updatesList = datasetsWithUpdates
          .map((d: DatasetUpdateInfo) => `${d.name} v${d.latest_version}`)
          .join(", ");
        toast.info(msg + updatesList);
      } else {
        toast.success("All datasets are up to date");
      }

      refresh();
    } catch (_err) {
      toast.error("Failed to check for updates");
    } finally {
      setCheckingUpdate(false);
    }
  }, [refresh]);

  return {
    datasets,
    setDatasets,
    loading,
    refreshing,
    error,
    filters,
    search,
    setSearch,
    setFilter,
    clearFilters,
    refresh,
    createDataset,
    updateDataset,
    toggleActive,
    deleteDataset,
    setDefault,
    checkAllUpdates,
    checkingUpdate,
    currentPage,
    totalPages,
    totalCount,
    currentLimit,
    goToPage,
    setPageSize,
  };
}
