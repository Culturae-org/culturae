"use client";

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import { apiGet } from "@/lib/api-client";
import { USERS_ENDPOINTS } from "@/lib/api/endpoints";
import type { UserSearchResult } from "@/lib/types/user.types";
import {
  IconApi,
  IconArrowsShuffle,
  IconChartBar,
  IconDashboard,
  IconDatabase,
  IconDatabaseExport,
  IconDeviceGamepad2,
  IconQuestionMark,
  IconReport,
  IconSettings,
  IconUsers,
  IconWorld,
} from "@tabler/icons-react";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router";

const PAGES = [
  { title: "Dashboard", url: "/", icon: IconDashboard },
  { title: "Users", url: "/users", icon: IconUsers },
  { title: "Datasets", url: "/datasets", icon: IconDatabase },
  { title: "Questions", url: "/questions", icon: IconQuestionMark },
  { title: "Geography", url: "/geography", icon: IconWorld },
  { title: "Games", url: "/games", icon: IconDeviceGamepad2 },
  { title: "Matchmaking", url: "/games/matchmaking", icon: IconArrowsShuffle },
  { title: "Analytics", url: "/analytics", icon: IconChartBar },
  { title: "Logs", url: "/logs", icon: IconReport },
  { title: "Reports", url: "/reports", icon: IconReport },
  { title: "API Explorer", url: "/api-explorer", icon: IconApi },
  { title: "Backup", url: "/backup", icon: IconDatabaseExport },
  { title: "Settings", url: "/settings", icon: IconSettings },
];

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [users, setUsers] = useState<UserSearchResult[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((o) => !o);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  useEffect(() => {
    if (!query || query.length < 2) {
      setUsers([]);
      return;
    }
    const controller = new AbortController();
    const timeout = setTimeout(async () => {
      try {
        const res = await apiGet(
          `${USERS_ENDPOINTS.SEARCH}?q=${encodeURIComponent(query)}&limit=5`,
          {
            signal: controller.signal,
          },
        );
        if (res.ok) {
          const data = await res.json();
          setUsers(
            Array.isArray(data.data)
              ? data.data
              : Array.isArray(data)
                ? data
                : [],
          );
        }
      } catch {}
    }, 300);
    return () => {
      clearTimeout(timeout);
      controller.abort();
    };
  }, [query]);

  const handleNavigate = useCallback(
    (url: string) => {
      setOpen(false);
      setQuery("");
      navigate(url);
    },
    [navigate],
  );

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput
        placeholder="Search pages, users..."
        value={query}
        onValueChange={setQuery}
      />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup heading="Pages">
          {PAGES.map((page) => (
            <CommandItem
              key={page.url}
              onSelect={() => handleNavigate(page.url)}
            >
              <page.icon className="mr-2 h-4 w-4" />
              {page.title}
            </CommandItem>
          ))}
        </CommandGroup>
        {users.length > 0 && (
          <>
            <CommandSeparator />
            <CommandGroup heading="Users">
              {users.map((u) => (
                <CommandItem
                  key={u.id}
                  onSelect={() => handleNavigate(`/users/${u.id}`)}
                >
                  <IconUsers className="mr-2 h-4 w-4" />
                  <span>{u.username}</span>
                  <span className="text-muted-foreground ml-2 text-xs">
                    {u.email}
                  </span>
                </CommandItem>
              ))}
            </CommandGroup>
          </>
        )}
      </CommandList>
    </CommandDialog>
  );
}
