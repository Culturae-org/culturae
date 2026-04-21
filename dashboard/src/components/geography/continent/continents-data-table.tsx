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
import { DataTablePagination } from "@/components/ui/data-table-pagination";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
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
import { useContinents } from "@/hooks/useGeography";
import { useIsMobile } from "@/hooks/useMobile";
import type { Continent } from "@/lib/types/geography.types";
import {
  IconAlertTriangle,
  IconColumns,
  IconRefresh,
  IconSearch,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type ColumnFiltersState,
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

interface ContinentsDataTableProps {
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

export function ContinentsDataTable({ datasetId }: ContinentsDataTableProps) {
  const {
    continents: data,
    loading,
    refreshing,
    error,
    refresh,
  } = useContinents(datasetId);
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
  const [pagination, setPagination] = React.useState({
    pageIndex: 0,
    pageSize: 10,
  });

  const formatNumber = React.useMemo(
    () => (num: number) => new Intl.NumberFormat("fr-FR").format(num),
    [],
  );

  const columns: ColumnDef<Continent>[] = React.useMemo(
    () => [
      {
        accessorKey: "id",
        header: "ID",
        size: 90,
        maxSize: 90,
        cell: ({ row }) => <ExpandableIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "dataset_id",
        header: "Dataset ID",
        size: 90,
        maxSize: 90,
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
        accessorKey: "area_km2",
        header: "Area (km²)",
        cell: ({ row }) => (
          <span className="tabular-nums">
            {formatNumber(row.getValue("area_km2"))}
          </span>
        ),
      },
      {
        accessorKey: "population",
        header: "Population",
        cell: ({ row }) => (
          <span className="tabular-nums">
            {formatNumber(row.getValue("population"))}
          </span>
        ),
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
    [formatNumber],
  );

  const isMobile = useIsMobile();

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onPaginationChange: setPagination,
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
          <CardTitle>Continents</CardTitle>
          <CardDescription>Manage continents data</CardDescription>
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
                  <Skeleton key={`cont-skel-${i}`} className="h-10 w-full" />
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
          <CardTitle>Continents</CardTitle>
          <CardDescription>Manage continents data</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <IconAlertTriangle className="h-8 w-8 text-red-500 mx-auto" />
              <p className="mt-2 text-sm text-destructive">{error}</p>
              <Button
                variant="outline"
                size="sm"
                onClick={() => refresh()}
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
        <CardTitle>Continents</CardTitle>
        <CardDescription>Manage continents data</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="relative flex-1 max-w-sm w-full lg:w-auto">
            <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search continents..."
              value={globalFilter}
              onChange={(e) => setGlobalFilter(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
            <Button
              variant="outline"
              size="sm"
              onClick={() => refresh()}
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

        {isMobile ? (
          <div className="space-y-3">
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => {
                const name = row.original.name;
                const displayName =
                  name?.en || name?.fr || Object.values(name || {})[0] || "-";
                return (
                  <div
                    key={row.id}
                    className="rounded-lg border bg-card p-4 space-y-3"
                  >
                    <div className="font-medium">{displayName}</div>
                    <div className="flex flex-wrap gap-2">
                      <Badge variant="secondary">
                        {(row.original.countries || []).length} countries
                      </Badge>
                    </div>
                    <div className="text-xs text-muted-foreground">
                      <div>Area: {formatNumber(row.original.area_km2)} km²</div>
                      <div>
                        Population: {formatNumber(row.original.population)}
                      </div>
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                No continents found.
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
                        <TableCell
                          key={cell.id}
                          style={{
                            width: cell.column.columnDef.size
                              ? `${cell.column.columnDef.size}px`
                              : undefined,
                          }}
                        >
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
                      No continents found.
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
          totalCount={data.length}
          onPageChange={(page) => table.setPageIndex(page)}
          onPageSizeChange={(size) => table.setPageSize(size)}
        />
      </CardContent>
    </Card>
  );
}
