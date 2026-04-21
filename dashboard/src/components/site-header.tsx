"use client";

import { CommandPalette } from "@/components/command-palette";
import { ThemeToggle } from "@/components/theme-toggle";
import { Badge } from "@/components/ui/badge";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { ArrowUpCircle } from "lucide-react";
import { Link, useLocation } from "react-router";

const ROUTE_LABELS: Record<string, string> = {
  "": "Dashboard",
  users: "Users",
  datasets: "Datasets",
  questions: "Questions",
  geography: "Geography",
  games: "Games",
  matchmaking: "Matchmaking",
  analytics: "Analytics",
  logs: "Logs",
  reports: "Reports",
  "api-explorer": "API Explorer",
  backup: "Backup",
  settings: "Settings",
};

const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function segmentLabel(seg: string, prevSeg?: string): string {
  if (ROUTE_LABELS[seg]) return ROUTE_LABELS[seg];
  if (UUID_RE.test(seg)) {
    const parent = prevSeg ? ROUTE_LABELS[prevSeg] : undefined;
    const prefix = parent ? parent.replace(/s$/, "") : "Detail";
    return `${prefix} ${seg.slice(0, 8)}...`;
  }
  return decodeURIComponent(seg);
}

interface SiteHeaderProps {
  version?: string | null;
  environment?: string | null;
  updateAvailable?: boolean;
  latestVersion?: string | null;
}

export function SiteHeader({ version, environment, updateAvailable, latestVersion }: SiteHeaderProps) {
  const pathname = useLocation().pathname;
  const segments = pathname.split("/").filter(Boolean);

  const breadcrumbs = segments.map((seg, i) => ({
    label: segmentLabel(seg, segments[i - 1]),
    href: `/${segments.slice(0, i + 1).join("/")}`,
  }));

  return (
    <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-(--header-height)">
      <div className="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator
          orientation="vertical"
          className="mx-2 data-[orientation=vertical]:h-4"
        />
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              {segments.length === 0 ? (
                <BreadcrumbPage>Dashboard</BreadcrumbPage>
              ) : (
                <BreadcrumbLink asChild>
                  <Link to="/">Dashboard</Link>
                </BreadcrumbLink>
              )}
            </BreadcrumbItem>
            {breadcrumbs.map((crumb, i) => (
              <span key={crumb.href} className="contents">
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  {i === breadcrumbs.length - 1 ? (
                    <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                  ) : (
                    <BreadcrumbLink asChild>
                      <Link to={crumb.href}>{crumb.label}</Link>
                    </BreadcrumbLink>
                  )}
                </BreadcrumbItem>
              </span>
            ))}
          </BreadcrumbList>
        </Breadcrumb>
        <div className="ml-auto flex items-center gap-2">
          {updateAvailable && (
            <Tooltip>
              <TooltipTrigger asChild>
                <a
                  href="https://github.com/Culturae-org/culturae/releases/latest"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-yellow-500 hover:text-yellow-400 transition-colors"
                >
                  <ArrowUpCircle className="h-5 w-5" />
                </a>
              </TooltipTrigger>
              <TooltipContent>
                Update available: {latestVersion ?? "new version"}
              </TooltipContent>
            </Tooltip>
          )}
          {environment && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Badge
                  variant={
                    environment === "development" || environment === "dev"
                      ? "secondary"
                      : "default"
                  }
                  className="hidden text-[10px] sm:inline-flex"
                >
                  {environment === "development" || environment === "dev"
                    ? "dev"
                    : "prod"}
                </Badge>
              </TooltipTrigger>
              <TooltipContent>Version: {version ?? "unknown"}</TooltipContent>
            </Tooltip>
          )}
          <Button variant="ghost" asChild size="sm" className="hidden sm:flex">
            <a
              href="https://culturae.me"
              rel="noopener noreferrer"
              target="_blank"
              className="dark:text-foreground"
            >
              Culturae
            </a>
          </Button>
          <ThemeToggle />
        </div>
      </div>
      <CommandPalette />
    </header>
  );
}
