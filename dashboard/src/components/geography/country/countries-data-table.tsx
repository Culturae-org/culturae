"use client";

import {
  CountryFlag,
  EmojiFlag,
} from "@/components/geography/country/country-flag";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTablePagination } from "@/components/ui/data-table-pagination";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useCountriesList } from "@/hooks/useGeography";
import { useIsMobile } from "@/hooks/useMobile";
import type { Country } from "@/lib/types/geography.types";
import {
  IconColumns,
  IconDotsVertical,
  IconEdit,
  IconEye,
  IconFilter,
  IconRefresh,
  IconSearch,
  IconX,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import * as React from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router";
import { CountryEditDialog } from "./country-edit-dialog";

const POPULATION_PRESETS = [
  { label: "All", min: "", max: "" },
  { label: "< 1M", min: "", max: "1000000" },
  { label: "1M - 10M", min: "1000000", max: "10000000" },
  { label: "10M - 50M", min: "10000000", max: "50000000" },
  { label: "50M - 100M", min: "50000000", max: "100000000" },
  { label: "> 100M", min: "100000000", max: "" },
];

const AREA_PRESETS = [
  { label: "All", min: "", max: "" },
  { label: "< 1,000", min: "", max: "1000" },
  { label: "1K - 100K", min: "1000", max: "100000" },
  { label: "100K - 500K", min: "100000", max: "500000" },
  { label: "500K - 1M", min: "500000", max: "1000000" },
  { label: "> 1M", min: "1000000", max: "" },
];

interface CountriesDataTableProps {
  datasetId: string;
  onTotalCountChange?: (count: number) => void;
}

function CountryIdCell({ id }: { id: string }) {
  const [isExpanded, setIsExpanded] = React.useState(false);

  return (
    <button
      type="button"
      className="font-mono text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors bg-transparent border-none p-0 text-left"
      onClick={() => setIsExpanded(!isExpanded)}
      onKeyUp={(e) => {
        if (e.key === "Enter" || e.key === " ") setIsExpanded(!isExpanded);
      }}
      title={isExpanded ? "Click to collapse" : "Click to expand"}
    >
      {isExpanded ? id : `${id.slice(0, 8)}...`}
    </button>
  );
}

export function CountriesDataTable({
  datasetId,
  onTotalCountChange,
}: CountriesDataTableProps) {
  const {
    countries: data,
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
  } = useCountriesList(datasetId);

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    [],
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      id: false,
      dataset_id: false,
      iso_alpha3: false,
      iso_numeric: false,
      official_name: false,
      latitude: false,
      longitude: false,
      area_km2: false,
      currency: false,
      languages: false,
      neighbors: false,
      tld: false,
      phone_code: false,
      driving_side: false,
      independent: false,
      updated_at: false,
    });
  const [viewCountry, setViewCountry] = React.useState<Country | null>(null);
  const [editCountry, setEditCountry] = React.useState<Country | null>(null);
  const [editOpen, setEditOpen] = React.useState(false);

  const isMobile = useIsMobile();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const pathname = useLocation().pathname;

  const viewId = searchParams.get("countryView");
  React.useEffect(() => {
    if (viewId && data.length > 0) {
      const found = data.find((c) => c.id === viewId);
      if (found) setViewCountry(found);
    }
  }, [viewId, data]);

  React.useEffect(() => {
    if (onTotalCountChange) onTotalCountChange(totalCount);
  }, [totalCount, onTotalCountChange]);

  const filteredRegions = React.useMemo(() => {
    if (!filters.continent) return regions;
    return regions.filter((r) => r.continent === filters.continent);
  }, [regions, filters.continent]);

  const hasActiveFilters = React.useMemo(() => {
    return Object.values(filters).some((v) => v !== "");
  }, [filters]);

  const getPopulationPresetLabel = () => {
    const preset = POPULATION_PRESETS.find(
      (p) => p.min === filters.populationMin && p.max === filters.populationMax,
    );
    return preset?.label || "Custom";
  };

  const getAreaPresetLabel = () => {
    const preset = AREA_PRESETS.find(
      (p) => p.min === filters.areaMin && p.max === filters.areaMax,
    );
    return preset?.label || "Custom";
  };

  const columns: ColumnDef<Country>[] = React.useMemo(
    () => [
      {
        accessorKey: "id",
        header: "ID",
        cell: ({ row }) => <CountryIdCell id={row.getValue("id")} />,
      },
      {
        id: "flag",
        header: "Flag",
        cell: ({ row }) => {
          const country = row.original;
          return (
            <CountryFlag
              countryCode={country.iso_alpha2}
              size="md"
              fallback={<EmojiFlag flag={country.flag} />}
            />
          );
        },
        enableSorting: false,
      },
      {
        accessorKey: "name",
        header: "Name",
        cell: ({ row }) => {
          const name = row.getValue("name") as Record<string, string>;
          return (
            <div className="font-medium">{name?.en || name?.fr || "-"}</div>
          );
        },
      },
      {
        accessorKey: "official_name",
        header: "Official Name",
        cell: ({ row }) => {
          const name = row.getValue("official_name") as Record<string, string>;
          return (
            <div className="text-muted-foreground text-sm">
              {name?.en || name?.fr || "-"}
            </div>
          );
        },
      },
      {
        accessorKey: "iso_alpha2",
        header: "ISO-2",
        cell: ({ row }) => (
          <Badge variant="outline" className="font-mono">
            {row.getValue("iso_alpha2")}
          </Badge>
        ),
      },
      {
        accessorKey: "iso_alpha3",
        header: "ISO-3",
        cell: ({ row }) => (
          <Badge variant="secondary" className="font-mono">
            {row.getValue("iso_alpha3")}
          </Badge>
        ),
      },
      {
        accessorKey: "iso_numeric",
        header: "ISO Num",
        cell: ({ row }) => (
          <span className="font-mono text-muted-foreground">
            {row.getValue("iso_numeric")}
          </span>
        ),
      },
      {
        accessorKey: "continent",
        header: "Continent",
        cell: ({ row }) => {
          const continent = row.getValue("continent") as string;
          return <Badge variant="secondary">{continent}</Badge>;
        },
      },
      {
        accessorKey: "region",
        header: "Region",
        cell: ({ row }) => {
          const region = row.getValue("region") as string;
          return region ? (
            <span className="text-sm">{region}</span>
          ) : (
            <span className="text-muted-foreground">-</span>
          );
        },
      },
      {
        accessorKey: "capital",
        header: "Capital",
        cell: ({ row }) => {
          const capital = row.original.capital as Record<string, string>;
          return (
            <span className="text-sm">{capital?.en || capital?.fr || "-"}</span>
          );
        },
      },
      {
        accessorKey: "population",
        header: "Population",
        cell: ({ row }) => {
          const pop = row.getValue("population") as number;
          return (
            <span className="text-sm tabular-nums">
              {pop ? pop.toLocaleString() : "-"}
            </span>
          );
        },
      },
      {
        accessorKey: "area_km2",
        header: "Area (km²)",
        cell: ({ row }) => {
          const area = row.getValue("area_km2") as number;
          return (
            <span className="text-sm tabular-nums">
              {area ? area.toLocaleString() : "-"}
            </span>
          );
        },
      },
      {
        id: "latitude",
        header: "Latitude",
        cell: ({ row }) => (
          <span className="font-mono text-xs">
            {row.original.coordinates?.lat?.toFixed(4) ?? "-"}
          </span>
        ),
      },
      {
        id: "longitude",
        header: "Longitude",
        cell: ({ row }) => (
          <span className="font-mono text-xs">
            {row.original.coordinates?.lng?.toFixed(4) ?? "-"}
          </span>
        ),
      },
      {
        accessorKey: "currency",
        header: "Currency",
        cell: ({ row }) => {
          const currency = row.original.currency;
          return currency?.code ? (
            <span className="text-sm">
              {currency.code} ({currency.symbol})
            </span>
          ) : (
            <span className="text-muted-foreground">-</span>
          );
        },
      },
      {
        accessorKey: "languages",
        header: "Languages",
        cell: ({ row }) => {
          const langs = row.original.languages;
          return langs?.length ? (
            <span className="text-sm">
              {langs.slice(0, 3).join(", ")}
              {langs.length > 3 ? "..." : ""}
            </span>
          ) : (
            <span className="text-muted-foreground">-</span>
          );
        },
      },
      {
        accessorKey: "neighbors",
        header: "Neighbors",
        cell: ({ row }) => {
          const neighbors = row.original.neighbors;
          return neighbors?.length ? (
            <span className="text-sm font-mono">
              {neighbors.slice(0, 3).join(", ")}
              {neighbors.length > 3 ? `+${neighbors.length - 3}` : ""}
            </span>
          ) : (
            <span className="text-muted-foreground">-</span>
          );
        },
      },
      {
        accessorKey: "tld",
        header: "TLD",
        cell: ({ row }) => (
          <span className="font-mono text-sm">
            {row.getValue("tld") || "-"}
          </span>
        ),
      },
      {
        accessorKey: "phone_code",
        header: "Phone",
        cell: ({ row }) => (
          <span className="font-mono text-sm">
            {row.getValue("phone_code") || "-"}
          </span>
        ),
      },
      {
        accessorKey: "driving_side",
        header: "Driving",
        cell: ({ row }) => (
          <span className="text-sm capitalize">
            {row.getValue("driving_side") || "-"}
          </span>
        ),
      },
      {
        accessorKey: "independent",
        header: "Independent",
        cell: ({ row }) => {
          const isIndependent = row.getValue("independent") as boolean;
          return (
            <Badge
              variant={isIndependent ? "default" : "outline"}
              className="text-xs"
            >
              {isIndependent ? "Yes" : "Territory"}
            </Badge>
          );
        },
      },
      {
        accessorKey: "created_at",
        header: "Created",
        cell: ({ row }) => {
          const date = new Date(row.getValue("created_at"));
          return (
            <div className="text-sm text-muted-foreground">
              {date.toLocaleDateString("en-US", {
                day: "2-digit",
                month: "2-digit",
                year: "numeric",
              })}
            </div>
          );
        },
      },
      {
        accessorKey: "updated_at",
        header: "Updated",
        cell: ({ row }) => {
          const date = new Date(row.getValue("updated_at"));
          return (
            <div className="text-sm text-muted-foreground">
              {date.toLocaleDateString("en-US", {
                day: "2-digit",
                month: "2-digit",
                year: "numeric",
              })}
            </div>
          );
        },
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => {
          const country = row.original;

          return (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="h-8 w-8 p-0">
                  <span className="sr-only">Open menu</span>
                  <IconDotsVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>Actions</DropdownMenuLabel>
                <DropdownMenuItem
                  onSelect={() => {
                    setViewCountry(country);
                    const params = new URLSearchParams(searchParams.toString());
                    params.set("countryView", country.id);
                    navigate(`${pathname}?${params.toString()}`);
                  }}
                >
                  <IconEye className="mr-2 h-4 w-4" />
                  View Country
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onSelect={() => {
                    setViewCountry(null);
                    setEditCountry(country);
                    setEditOpen(true);
                  }}
                >
                  <IconEdit className="mr-2 h-4 w-4" />
                  Edit Country
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          );
        },
      },
    ],
    [searchParams, pathname, navigate],
  );

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    manualPagination: true,
    pageCount: totalPages,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      pagination: {
        pageIndex: currentPage - 1,
        pageSize: currentLimit,
      },
    },
  });

  if (loading && !refreshing && data.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Countries</CardTitle>
          <CardDescription>Manage country data</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-32">
            <div className="text-muted-foreground">Loading countries...</div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Countries</CardTitle>
          <CardDescription>Manage country data</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-32">
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>Countries</CardTitle>
        <CardDescription>Manage country data</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="relative">
          <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
            <div className="relative flex-1 max-w-sm w-full lg:w-auto">
              <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search countries by name, code..."
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                className="pl-10 pr-10"
              />
              {search && (
                <IconX
                  className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground cursor-pointer"
                  onClick={() => setSearch("")}
                />
              )}
            </div>
            <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
              <Button
                variant="outline"
                size="sm"
                onClick={refresh}
                disabled={loading || refreshing}
              >
                <IconRefresh
                  className={`h-4 w-4 ${loading || refreshing ? "animate-spin" : ""}`}
                />
              </Button>

              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="outline"
                    size="sm"
                    className={hasActiveFilters ? "border-primary" : ""}
                  >
                    <IconFilter className="h-4 w-4" />
                    Filter
                    {hasActiveFilters && (
                      <Badge variant="secondary" className="ml-1 h-5 px-1">
                        {Object.values(filters).filter((v) => v !== "").length}
                      </Badge>
                    )}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-72">
                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Continent
                  </div>
                  <div className="px-2 pb-2">
                    <Select
                      value={filters.continent || "all"}
                      onValueChange={(value) =>
                        setFilter("continent", value === "all" ? "" : value)
                      }
                    >
                      <SelectTrigger className="h-8">
                        <SelectValue placeholder="All continents" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All continents</SelectItem>
                        {continents.map((c) => (
                          <SelectItem key={c.slug} value={c.slug}>
                            {(c.name as Record<string, string>)?.en || c.slug}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Region
                  </div>
                  <div className="px-2 pb-2">
                    <Select
                      value={filters.region || "all"}
                      onValueChange={(value) =>
                        setFilter("region", value === "all" ? "" : value)
                      }
                      disabled={!filters.continent}
                    >
                      <SelectTrigger className="h-8">
                        <SelectValue
                          placeholder={
                            filters.continent
                              ? "All regions"
                              : "Select continent first"
                          }
                        />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All regions</SelectItem>
                        {filteredRegions.map((r) => (
                          <SelectItem key={r.slug} value={r.slug}>
                            {(r.name as Record<string, string>)?.en || r.slug}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <DropdownMenuSeparator />

                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Population
                  </div>
                  <div className="px-2 pb-2 flex flex-wrap gap-1">
                    {POPULATION_PRESETS.map((preset) => (
                      <Button
                        key={preset.label}
                        variant={
                          getPopulationPresetLabel() === preset.label
                            ? "default"
                            : "outline"
                        }
                        size="sm"
                        className="h-7 text-xs"
                        onClick={() => {
                          setFilter("populationMin", preset.min);
                          setFilter("populationMax", preset.max);
                        }}
                      >
                        {preset.label}
                      </Button>
                    ))}
                  </div>

                  <DropdownMenuSeparator />

                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Area (km²)
                  </div>
                  <div className="px-2 pb-2 flex flex-wrap gap-1">
                    {AREA_PRESETS.map((preset) => (
                      <Button
                        key={preset.label}
                        variant={
                          getAreaPresetLabel() === preset.label
                            ? "default"
                            : "outline"
                        }
                        size="sm"
                        className="h-7 text-xs"
                        onClick={() => {
                          setFilter("areaMin", preset.min);
                          setFilter("areaMax", preset.max);
                        }}
                      >
                        {preset.label}
                      </Button>
                    ))}
                  </div>

                  <DropdownMenuSeparator />

                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Status
                  </div>
                  <div className="px-2 pb-2 flex flex-wrap gap-1">
                    {[
                      { label: "All", value: "" },
                      { label: "Independent", value: "true" },
                      { label: "Territories", value: "false" },
                    ].map((opt) => (
                      <Button
                        key={opt.label}
                        variant={
                          filters.independent === opt.value
                            ? "default"
                            : "outline"
                        }
                        size="sm"
                        className="h-7 text-xs"
                        onClick={() => setFilter("independent", opt.value)}
                      >
                        {opt.label}
                      </Button>
                    ))}
                  </div>

                  <DropdownMenuSeparator />

                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Driving side
                  </div>
                  <div className="px-2 pb-2 flex flex-wrap gap-1">
                    {[
                      { label: "All", value: "" },
                      { label: "Right", value: "right" },
                      { label: "Left", value: "left" },
                    ].map((opt) => (
                      <Button
                        key={opt.label}
                        variant={
                          filters.drivingSide === opt.value
                            ? "default"
                            : "outline"
                        }
                        size="sm"
                        className="h-7 text-xs"
                        onClick={() => setFilter("drivingSide", opt.value)}
                      >
                        {opt.label}
                      </Button>
                    ))}
                  </div>

                  {hasActiveFilters && (
                    <>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        onClick={clearFilters}
                        className="text-destructive focus:text-destructive"
                      >
                        <IconX className="mr-2 h-4 w-4" />
                        Clear all filters
                      </DropdownMenuItem>
                    </>
                  )}
                </DropdownMenuContent>
              </DropdownMenu>

              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="sm">
                    <IconColumns className="h-4 w-4" />
                    Columns
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  className="max-h-80 overflow-y-auto"
                >
                  {table
                    .getAllColumns()
                    .filter((column) => column.getCanHide())
                    .map((column) => {
                      return (
                        <DropdownMenuCheckboxItem
                          key={column.id}
                          className="capitalize"
                          checked={column.getIsVisible()}
                          onCheckedChange={(value) =>
                            column.toggleVisibility(!!value)
                          }
                        >
                          {column.id.replace(/_/g, " ")}
                        </DropdownMenuCheckboxItem>
                      );
                    })}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>

          {hasActiveFilters && (
            <div className="flex flex-wrap gap-2 pb-4">
              {filters.continent && (
                <Badge variant="secondary" className="gap-1">
                  Continent:{" "}
                  {(
                    continents.find((c) => c.slug === filters.continent)
                      ?.name as Record<string, string>
                  )?.en || filters.continent}
                  <IconX
                    className="h-3 w-3 cursor-pointer"
                    onClick={() => setFilter("continent", "")}
                  />
                </Badge>
              )}
              {filters.region && (
                <Badge variant="secondary" className="gap-1">
                  Region:{" "}
                  {(
                    regions.find((r) => r.slug === filters.region)
                      ?.name as Record<string, string>
                  )?.en || filters.region}
                  <IconX
                    className="h-3 w-3 cursor-pointer"
                    onClick={() => setFilter("region", "")}
                  />
                </Badge>
              )}
              {(filters.populationMin || filters.populationMax) && (
                <Badge variant="secondary" className="gap-1">
                  Population: {getPopulationPresetLabel()}
                  <IconX
                    className="h-3 w-3 cursor-pointer"
                    onClick={() => {
                      setFilter("populationMin", "");
                      setFilter("populationMax", "");
                    }}
                  />
                </Badge>
              )}
              {(filters.areaMin || filters.areaMax) && (
                <Badge variant="secondary" className="gap-1">
                  Area: {getAreaPresetLabel()}
                  <IconX
                    className="h-3 w-3 cursor-pointer"
                    onClick={() => {
                      setFilter("areaMin", "");
                      setFilter("areaMax", "");
                    }}
                  />
                </Badge>
              )}
              {filters.independent && (
                <Badge variant="secondary" className="gap-1">
                  {filters.independent === "true"
                    ? "Independent only"
                    : "Territories only"}
                  <IconX
                    className="h-3 w-3 cursor-pointer"
                    onClick={() => setFilter("independent", "")}
                  />
                </Badge>
              )}
              {filters.drivingSide && (
                <Badge variant="secondary" className="gap-1">
                  {filters.drivingSide === "right"
                    ? "Right-hand drive"
                    : "Left-hand drive"}
                  <IconX
                    className="h-3 w-3 cursor-pointer"
                    onClick={() => setFilter("drivingSide", "")}
                  />
                </Badge>
              )}
            </div>
          )}

          <div className="rounded-md border overflow-x-auto">
            <Table className="w-full">
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => {
                      return (
                        <TableHead key={header.id}>
                          {header.isPlaceholder
                            ? null
                            : flexRender(
                                header.column.columnDef.header,
                                header.getContext(),
                              )}
                        </TableHead>
                      );
                    })}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      className="cursor-pointer"
                      onClick={(e) => {
                        const target = e.target as HTMLElement;
                        if (
                          target.closest(
                            'button, [role="checkbox"], [role="menuitem"], [data-radix-collection-item]',
                          )
                        )
                          return;
                        const country = row.original;
                        setViewCountry(country);
                        const params = new URLSearchParams(
                          searchParams.toString(),
                        );
                        params.set("countryView", country.id);
                        navigate(`${pathname}?${params.toString()}`);
                      }}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id}>
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext(),
                          )}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell
                      colSpan={columns.length}
                      className="h-24 text-center"
                    >
                      No countries found.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>

          <DataTablePagination
            pageIndex={currentPage - 1}
            pageCount={totalPages}
            pageSize={currentLimit}
            totalCount={totalCount}
            onPageChange={(page) => goToPage(page + 1)}
            onPageSizeChange={setPageSize}
            loading={loading || refreshing}
          />

          <Drawer
            direction={isMobile ? "bottom" : "right"}
            open={!!viewCountry}
            onOpenChange={(open) => {
              if (!open) {
                setViewCountry(null);
                const params = new URLSearchParams(searchParams.toString());
                params.delete("countryView");
                const qs = params.toString();
                navigate(qs ? `${pathname}?${qs}` : pathname);
              }
            }}
          >
            <DrawerContent>
              {viewCountry && (
                <>
                  <DrawerHeader className="gap-1">
                    <DrawerTitle className="flex items-center gap-2">
                      <EmojiFlag flag={viewCountry.flag} />
                      {viewCountry.name?.en ||
                        viewCountry.name?.fr ||
                        viewCountry.slug}
                    </DrawerTitle>
                    <DrawerDescription>
                      Country details for{" "}
                      <b>{viewCountry.name?.en || viewCountry.slug}</b>
                    </DrawerDescription>
                  </DrawerHeader>

                  <div className="flex flex-col gap-6 overflow-y-auto px-4 pb-4">
                    <div className="flex flex-col items-center gap-3 p-4 bg-muted/50 rounded-lg">
                      <CountryFlag
                        countryCode={viewCountry.iso_alpha2}
                        size="lg"
                        fallback={<EmojiFlag flag={viewCountry.flag} />}
                      />
                      <div className="text-center">
                        <h3 className="font-semibold text-lg">
                          {viewCountry.name?.en || viewCountry.slug}
                        </h3>
                        {viewCountry.capital?.en && (
                          <p className="text-sm text-muted-foreground">
                            Capital: {viewCountry.capital.en}
                          </p>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant="outline" className="font-mono">
                          {viewCountry.iso_alpha2}
                        </Badge>
                        <Badge variant="secondary" className="font-mono">
                          {viewCountry.iso_alpha3}
                        </Badge>
                        {viewCountry.iso_numeric && (
                          <Badge
                            variant="outline"
                            className="font-mono text-xs"
                          >
                            {viewCountry.iso_numeric}
                          </Badge>
                        )}
                      </div>
                    </div>

                    <div className="space-y-4">
                      <h4 className="font-medium text-sm">Geography</h4>
                      <div className="space-y-3 pl-2">
                        <div className="flex flex-col gap-1">
                          <Label className="text-xs text-muted-foreground">
                            Continent
                          </Label>
                          <Badge variant="secondary" className="w-fit">
                            {viewCountry.continent}
                          </Badge>
                        </div>
                        {viewCountry.region && (
                          <div className="flex flex-col gap-1">
                            <Label className="text-xs text-muted-foreground">
                              Region
                            </Label>
                            <div className="text-sm">{viewCountry.region}</div>
                          </div>
                        )}
                        <div className="grid grid-cols-2 gap-3">
                          <div className="flex flex-col gap-1">
                            <Label className="text-xs text-muted-foreground">
                              Population
                            </Label>
                            <div className="text-sm font-medium">
                              {viewCountry.population?.toLocaleString() || "-"}
                            </div>
                          </div>
                          <div className="flex flex-col gap-1">
                            <Label className="text-xs text-muted-foreground">
                              Area
                            </Label>
                            <div className="text-sm font-medium">
                              {viewCountry.area_km2
                                ? `${viewCountry.area_km2.toLocaleString()} km²`
                                : "-"}
                            </div>
                          </div>
                        </div>
                        <div className="flex flex-col gap-1">
                          <Label className="text-xs text-muted-foreground">
                            Coordinates
                          </Label>
                          <div className="text-xs font-mono">
                            {viewCountry.coordinates
                              ? `${viewCountry.coordinates.lat.toFixed(4)}, ${viewCountry.coordinates.lng.toFixed(4)}`
                              : "-"}
                          </div>
                        </div>
                      </div>
                    </div>

                    <Separator />

                    <div className="space-y-4">
                      <h4 className="font-medium text-sm">Details</h4>
                      <div className="space-y-3 pl-2">
                        <div className="grid grid-cols-2 gap-3">
                          <div className="flex flex-col gap-1">
                            <Label className="text-xs text-muted-foreground">
                              Currency
                            </Label>
                            <div className="text-sm">
                              {viewCountry.currency?.name
                                ? `${viewCountry.currency.name} (${viewCountry.currency.code}) ${viewCountry.currency.symbol}`
                                : "-"}
                            </div>
                          </div>
                          <div className="flex flex-col gap-1">
                            <Label className="text-xs text-muted-foreground">
                              Phone Code
                            </Label>
                            <div className="text-sm font-mono">
                              {viewCountry.phone_code || "-"}
                            </div>
                          </div>
                          <div className="flex flex-col gap-1">
                            <Label className="text-xs text-muted-foreground">
                              TLD
                            </Label>
                            <div className="text-sm font-mono">
                              {viewCountry.tld || "-"}
                            </div>
                          </div>
                          <div className="flex flex-col gap-1">
                            <Label className="text-xs text-muted-foreground">
                              Driving Side
                            </Label>
                            <div className="text-sm capitalize">
                              {viewCountry.driving_side || "-"}
                            </div>
                          </div>
                        </div>
                        <div className="flex flex-col gap-1">
                          <Label className="text-xs text-muted-foreground">
                            Languages
                          </Label>
                          <div className="text-sm">
                            {viewCountry.languages?.join(", ") || "-"}
                          </div>
                        </div>
                        <div className="flex flex-col gap-1">
                          <Label className="text-xs text-muted-foreground">
                            Independent
                          </Label>
                          <Badge
                            variant={
                              viewCountry.independent ? "default" : "outline"
                            }
                            className="w-fit"
                          >
                            {viewCountry.independent ? "Yes" : "Territory"}
                          </Badge>
                        </div>
                      </div>
                    </div>

                    {viewCountry.neighbors?.length > 0 && (
                      <>
                        <Separator />
                        <div className="space-y-4">
                          <h4 className="font-medium text-sm">
                            Neighbors ({viewCountry.neighbors.length})
                          </h4>
                          <div className="flex flex-wrap gap-1 pl-2">
                            {viewCountry.neighbors.map((n) => (
                              <Badge key={n} variant="outline">
                                {n}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      </>
                    )}

                    <Separator />

                    <div className="space-y-4">
                      <h4 className="font-medium text-sm">Names (i18n)</h4>
                      <div className="grid grid-cols-2 gap-2 pl-2">
                        {Object.entries(viewCountry.name || {}).map(
                          ([lang, name]) => (
                            <div key={lang} className="flex items-center gap-2">
                              <Badge
                                variant="outline"
                                className="uppercase text-xs w-10 justify-center"
                              >
                                {lang}
                              </Badge>
                              <span className="text-sm">{name}</span>
                            </div>
                          ),
                        )}
                      </div>
                    </div>
                  </div>

                  <DrawerFooter className="flex flex-row gap-2">
                    <DrawerClose asChild>
                      <Button variant="outline" className="flex-1">
                        Close
                      </Button>
                    </DrawerClose>
                  </DrawerFooter>
                </>
              )}
            </DrawerContent>
          </Drawer>

          {editCountry && (
            <CountryEditDialog
              country={editCountry}
              datasetId={datasetId}
              open={editOpen}
              onOpenChange={(open) => {
                setEditOpen(open);
                if (!open) setEditCountry(null);
              }}
              onSave={refresh}
            />
          )}
        </div>
      </CardContent>
    </Card>
  );
}
