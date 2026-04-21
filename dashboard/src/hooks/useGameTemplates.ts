"use client";

import gameTemplatesService from "@/lib/services/game-templates.service";
import type {
  CreateGameTemplateRequest,
  GameTemplate,
  GameTemplatesQueryParams,
  UpdateGameTemplateRequest,
} from "@/lib/types/game-template.types";
import { useCallback, useState } from "react";
import { toast } from "sonner";

export function useGameTemplates() {
  const [templates, setTemplates] = useState<GameTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [currentLimit, setCurrentLimit] = useState(10);
  const fetchTemplates = useCallback(
    async (params: GameTemplatesQueryParams = {}) => {
      setLoading(true);
      setError(null);
      try {
        const res = await gameTemplatesService.getTemplates(params);
        setTemplates(res.data);
        setTotal(res.total);
        setCurrentPage(res.page);
        setTotalPages(res.total_pages);
        setCurrentLimit(res.limit);
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to load game templates";
        setError(msg);
        toast.error(msg);
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  const createTemplate = useCallback(
    async (data: CreateGameTemplateRequest): Promise<GameTemplate | null> => {
      try {
        const created = await gameTemplatesService.createTemplate(data);
        setTemplates((prev) => [created, ...prev]);
        setTotal((prev) => prev + 1);
        toast.success(`Template "${created.name}" created`);
        return created;
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to create template";
        toast.error(msg);
        return null;
      }
    },
    [],
  );

  const updateTemplate = useCallback(
    async (
      id: string,
      data: UpdateGameTemplateRequest,
    ): Promise<GameTemplate | null> => {
      try {
        const updated = await gameTemplatesService.updateTemplate(id, data);
        setTemplates((prev) => prev.map((t) => (t.id === id ? updated : t)));
        toast.success(`Template "${updated.name}" updated`);
        return updated;
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to update template";
        toast.error(msg);
        return null;
      }
    },
    [],
  );

  const deleteTemplate = useCallback(
    async (id: string, name: string): Promise<boolean> => {
      try {
        await gameTemplatesService.deleteTemplate(id);
        setTemplates((prev) => prev.filter((t) => t.id !== id));
        setTotal((prev) => prev - 1);
        toast.success(`Template "${name}" deleted`);
        return true;
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to delete template";
        toast.error(msg);
        return false;
      }
    },
    [],
  );

  const seedDefaultTemplates = useCallback(
    async (onDone?: () => void): Promise<void> => {
      try {
        const res = await gameTemplatesService.seedDefaultTemplates();
        if (res.created === 0) {
          toast.info("All default modes already exist");
        } else {
          toast.success(
            `${res.created} default mode${res.created !== 1 ? "s" : ""} added`,
          );
        }
        onDone?.();
      } catch (err) {
        const msg =
          err instanceof Error
            ? err.message
            : "Failed to seed default templates";
        toast.error(msg);
      }
    },
    [],
  );

  return {
    templates,
    loading,
    error,
    total,
    currentPage,
    totalPages,
    currentLimit,
    fetchTemplates,
    createTemplate,
    updateTemplate,
    deleteTemplate,
    seedDefaultTemplates,
  };
}
