"use client";

import {
  type AnyApiReferenceConfiguration,
  ApiReferenceReact,
} from "@scalar/api-reference-react";
import "@scalar/api-reference-react/style.css";
import { OPENAPI_ENDPOINTS } from "@/lib/api/endpoints";
import { useTheme } from "next-themes";

export default function ApiExplorerPage() {
  const { theme } = useTheme();

  const configuration: Partial<AnyApiReferenceConfiguration> = {
    url: OPENAPI_ENDPOINTS.SPEC,
    theme: "deepSpace",
    darkMode: theme !== "light",
    hideModels: true,
    hideDarkModeToggle: true,
    hideClientButton: true,
    fetch: (input: string | URL | Request, init?: RequestInit) => {
      return fetch(input, {
        ...init,
        credentials: "include",
      });
    },
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-4 p-4 border-b">
        <h1 className="text-xl font-semibold">API Explorer</h1>
      </div>
      <div className="flex-1 overflow-auto">
        <ApiReferenceReact key={theme} configuration={configuration} />
      </div>
    </div>
  );
}
