import { apiGet, apiPut } from "@/lib/api-client";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

export interface UseSettingsOptions<T> {
  endpoint: string;
  defaults: T;
  validate?: (value: T) => boolean | string;
}

export function useSettings<T extends Record<string, unknown>>({
  endpoint,
  defaults,
  validate,
}: UseSettingsOptions<T>) {
  const [config, setConfig] = useState<T>(defaults);
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);

  useEffect(() => {
    const loadConfig = async () => {
      try {
        const res = await apiGet(endpoint);
        if (res.ok) {
          const json = await res.json();
          const data: T = (json.data ?? json) as T;
          setConfig(data);
        }
      } catch (error) {
        console.error("Failed to load settings:", error);
      } finally {
        setInitialLoading(false);
      }
    };

    loadConfig();
  }, [endpoint]);

  const updateProperty = useCallback(
    <K extends keyof T>(key: K, value: T[K]) => {
      setConfig((prev) => ({ ...prev, [key]: value }));
    },
    [],
  );

  const save = useCallback(
    async (data?: T) => {
      const toSave = data ?? config;

      if (validate) {
        const validation = validate(toSave);
        if (validation !== true) {
          toast.error(
            typeof validation === "string"
              ? validation
              : "Invalid configuration",
          );
          return false;
        }
      }

      setLoading(true);
      try {
        const res = await apiPut(endpoint, toSave);
        if (!res.ok) {
          throw new Error("Failed to save settings");
        }
        toast.success("Configuration updated successfully");
        return true;
      } catch (error) {
        toast.error(
          error instanceof Error
            ? error.message
            : "Failed to save configuration",
        );
        return false;
      } finally {
        setLoading(false);
      }
    },
    [config, endpoint, validate],
  );

  const reset = useCallback(() => {
    setConfig(defaults);
  }, [defaults]);

  return {
    config,
    setConfig,
    updateProperty,
    save,
    reset,
    loading,
    initialLoading,
  };
}
