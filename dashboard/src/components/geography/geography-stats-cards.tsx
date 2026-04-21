"use client";

import { Badge } from "@/components/ui/badge";
import { IconFlag, IconMap, IconWorld } from "@tabler/icons-react";

interface GeographyStats {
  country_count: number;
  continent_count: number;
  region_count: number;
  flag_count: number;
  flag_png512_count?: number;
  flag_png1024_count?: number;
}

interface GeographyStatsCardsProps {
  stats: GeographyStats | null;
}

export function GeographyStatsCards({ stats }: GeographyStatsCardsProps) {
  if (!stats) return null;

  return (
    <div className="flex flex-wrap gap-2">
      <Badge variant="secondary" className="gap-1.5">
        <IconFlag className="h-3.5 w-3.5" />
        <span>{stats.country_count} pays</span>
      </Badge>
      <Badge variant="secondary" className="gap-1.5">
        <IconWorld className="h-3.5 w-3.5" />
        <span>{stats.continent_count} continents</span>
      </Badge>
      <Badge variant="secondary" className="gap-1.5">
        <IconMap className="h-3.5 w-3.5" />
        <span>{stats.region_count} régions</span>
      </Badge>
      <Badge variant="secondary" className="gap-1.5">
        <IconFlag className="h-3.5 w-3.5" />
        <span>{stats.flag_count} drapeaux</span>
      </Badge>
    </div>
  );
}
