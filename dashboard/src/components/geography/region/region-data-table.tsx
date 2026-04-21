"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { DataTablePagination } from "@/components/ui/data-table-pagination";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useContinents, useRegions } from "@/hooks/useGeography";
import { useIsMobile } from "@/hooks/useMobile";
import type { Region } from "@/lib/types/geography.types";
import {
  IconAlertTriangle,
  IconColumns,
  IconFilter,
  IconRefresh,
  IconSearch,
  IconX,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type ColumnFiltersState,
  type PaginationState,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import * as React from "react";

interface RegionsDataTableProps {
  datasetId: string;
}

function ExpandableIdCell({ id }: { id: string }) {
  const [isExpanded, setIsExpanded] = React.useState(false);

  return (
    <button
      type="button"
      className="font-mono text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors text-left"
      onClick={() => setIsExpanded(!isExpanded)}
      title={isExpanded ? "Click to collapse" : "Click to expand"}
    >
      {isExpanded ? id : `${id.slice(0, 8)}...`}
    </button>
  );
}

export function RegionsDataTable({ datasetId }: RegionsDataTableProps) {
  const {
    regions: allRegions,
    loading,
    refreshing,
    error,
    refresh,
  } = useRegions(datasetId);
  const { continents } = useContinents(datasetId);

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    [],
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      id: false,
      dataset_id: false,
      slug: false,
      created_at: false,
      updated_at: false,
    });
  const [globalFilter, setGlobalFilter] = React.useState("");
  const [pagination, setPagination] = React.useState<PaginationState>({
    pageIndex: 0,
    pageSize: 10,
  });

  const [selectedContinents, setSelectedContinents] = React.useState<string[]>(
    [],
  );

  const isMobile = useIsMobile();

  const filteredRegions = React.useMemo(() => {
    if (selectedContinents.length === 0) return allRegions;
    return allRegions.filter((r: Region) =>
      selectedContinents.includes(r.continent),
    );
  }, [allRegions, selectedContinents]);

  const handleContinentChange = (continentSlug: string, checked: boolean) => {
    if (checked) {
      setSelectedContinents((prev) => [...prev, continentSlug]);
    } else {
      setSelectedContinents((prev) => prev.filter((c) => c !== continentSlug));
    }
    setPagination((prev) => ({ ...prev, pageIndex: 0 }));
  };

  const clearContinentFilter = (continentSlug: string) => {
    setSelectedContinents((prev) => prev.filter((c) => c !== continentSlug));
    setPagination((prev) => ({ ...prev, pageIndex: 0 }));
  };

  const _clearAllContinentFilters = () => {
    setSelectedContinents([]);
    setPagination((prev) => ({ ...prev, pageIndex: 0 }));
  };

  const columns: ColumnDef<Region>[] = React.useMemo(
    () => [
      {
        accessorKey: "id",
        header: "ID",
        cell: ({ row }) => <ExpandableIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "dataset_id",
        header: "Dataset ID",
        cell: ({ row }) => <ExpandableIdCell id={row.getValue("dataset_id")} />,
      },
      {
        accessorKey: "slug",
        header: "Slug",
        cell: ({ row }) => (
          <code className="text-xs bg-muted px-1.5 py-0.5 rounded">
            {row.getValue("slug")}
          </code>
        ),
      },
      {
        accessorKey: "name",
        header: "Name",
        cell: ({ row }) => {
          const name = row.getValue<Record<string, string>>("name");
          return (
            <div className="font-medium">
              {name?.en || name?.fr || Object.values(name || {})[0] || "-"}
            </div>
          );
        },
        filterFn: (row, id, value) => {
          const name = row.getValue<Record<string, string>>(id);
          const searchValue = value.toLowerCase();
          return Object.values(name || {}).some((n) =>
            n.toLowerCase().includes(searchValue),
          );
        },
      },
      {
        accessorKey: "continent",
        header: "Continent",
        cell: ({ row }) => (
          <Badge variant="outline">{row.getValue("continent")}</Badge>
        ),
      },
      {
        accessorKey: "countries",
        header: "Countries",
        cell: ({ row }) => {
          const countries = row.getValue<string[]>("countries") || [];
          return (
            <Badge variant="secondary">{countries.length} countries</Badge>
          );
        },
      },
      {
        accessorKey: "created_at",
        header: "Created",
        cell: ({ row }) => (
          <span className="text-xs text-muted-foreground">
            {new Date(row.getValue("created_at")).toLocaleDateString("fr-FR")}
          </span>
        ),
      },
      {
        accessorKey: "updated_at",
        header: "Updated",
        cell: ({ row }) => (
          <span className="text-xs text-muted-foreground">
            {new Date(row.getValue("updated_at")).toLocaleDateString("fr-FR")}
          </span>
        ),
      },
    ],
    [],
  );

  const table = useReactTable({
    data: filteredRegions,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onGlobalFilterChange: setGlobalFilter,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      globalFilter,
      pagination,
    },
  });

  if (loading && !refreshing) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Regions</CardTitle>
          <CardDescription>Manage regions data</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4 py-4">
            <div className="flex items-center justify-between">
              <Skeleton className="h-9 w-64" />
              <Skeleton className="h-9 w-32" />
            </div>
            <div className="rounded-md border">
              <div className="space-y-3 p-4">
                {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
                  <Skeleton key={`region-skel-${i}`} className="h-10 w-full" />
                ))}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Regions</CardTitle>
          <CardDescription>Manage regions data</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <IconAlertTriangle className="h-8 w-8 text-red-500 mx-auto" />
              <p className="mt-2 text-sm text-red-600">{error}</p>
              <Button
                variant="outline"
                size="sm"
                onClick={refresh}
                className="mt-2"
              >
                <IconRefresh className="mr-2 h-4 w-4 text-current" />
                Retry
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>Regions</CardTitle>
        <CardDescription>Manage regions data</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="relative flex-1 max-w-sm w-full lg:w-auto">
            <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search regions..."
              value={globalFilter}
              onChange={(e) => setGlobalFilter(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconFilter className="h-4 w-4" />
                  Filter
                  {selectedContinents.length > 0 && (
                    <Badge
                      variant="secondary"
                      className="ml-2 h-5 px-1.5 text-xs"
                    >
                      {selectedContinents.length}
                    </Badge>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Continent
                </div>
                {continents.map((c) => (
                  <DropdownMenuItem
                    key={c.slug}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={selectedContinents.includes(c.slug)}
                      onCheckedChange={(checked) =>
                        handleContinentChange(c.slug, !!checked)
                      }
                      className="mr-2"
                    />
                    {(c.name as Record<string, string>)?.en || c.slug}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>

            <Button
              variant="outline"
              size="sm"
              onClick={refresh}
              disabled={refreshing}
            >
              <IconRefresh
                className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
              />
            </Button>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconColumns className="h-4 w-4 mr-2" />
                  Columns
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
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

        {selectedContinents.length > 0 && (
          <div className="flex flex-wrap gap-2 pb-4">
            {selectedContinents.map((continentSlug) => {
              const continent = continents.find(
                (c) => c.slug === continentSlug,
              );
              return (
                <Badge
                  key={continentSlug}
                  variant="secondary"
                  className="gap-1"
                >
                  {(continent?.name as Record<string, string>)?.en ||
                    continentSlug}
                  <IconX
                    className="h-3 w-3 cursor-pointer"
                    onClick={() => clearContinentFilter(continentSlug)}
                  />
                </Badge>
              );
            })}
          </div>
        )}

        {isMobile ? (
          <div className="space-y-3">
            {filteredRegions.length ? (
              filteredRegions
                .slice(
                  pagination.pageIndex * pagination.pageSize,
                  (pagination.pageIndex + 1) * pagination.pageSize,
                )
                .map((row) => (
                  <div
                    key={row.id}
                    className="rounded-lg border bg-card p-4 space-y-3"
                  >
                    <div className="font-medium">
                      {(row.name as Record<string, string>)?.en ||
                        (row.name as Record<string, string>)?.fr ||
                        "-"}
                    </div>
                    <div className="flex flex-wrap gap-2">
                      <Badge variant="outline">{row.continent}</Badge>
                      <Badge variant="secondary">
                        {(row.countries || []).length} countries
                      </Badge>
                    </div>
                  </div>
                ))
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                No regions found.
              </div>
            )}
          </div>
        ) : (
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id}>
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                              header.column.columnDef.header,
                              header.getContext(),
                            )}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow key={row.id}>
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
                      No regions found.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}

        <DataTablePagination
          pageIndex={pagination.pageIndex}
          pageCount={table.getPageCount() || 1}
          pageSize={pagination.pageSize}
          totalCount={filteredRegions.length}
          onPageChange={(page) => table.setPageIndex(page)}
          onPageSizeChange={(size) => table.setPageSize(size)}
        />
      </CardContent>
    </Card>
  );
}
