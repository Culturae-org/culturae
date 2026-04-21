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
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useImports } from "@/hooks/useImports";
import { useIsMobile } from "@/hooks/useMobile";
import type { ImportJob } from "@/lib/types/datasets.types";
import {
  IconAlertTriangle,
  IconCalendar,
  IconCircleCheck,
  IconCircleX,
  IconClock,
  IconColumns,
  IconDotsVertical,
  IconDownload,
  IconRefresh,
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
import { format, formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import * as React from "react";
import { useNavigate } from "react-router";

function ImportIdCell({ id }: { id: string }) {
  const [isExpanded, setIsExpanded] = React.useState(false);

  return (
    <button
      type="button"
      className="font-mono text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors text-left"
      onClick={(e) => {
        e.stopPropagation();
        setIsExpanded(!isExpanded);
      }}
      title={isExpanded ? "Click to collapse" : "Click to expand"}
    >
      {isExpanded ? id : `${id.slice(0, 8)}...`}
    </button>
  );
}

function ImportStatusCell({ job }: { job: ImportJob }) {
  if (!job.finished_at) {
    return (
      <div className="flex items-center gap-2 text-blue-600">
        <IconClock className="h-4 w-4 animate-spin" />
        <span className="text-sm">In progress</span>
      </div>
    );
  }

  if (job.success) {
    return (
      <div className="flex items-center gap-2 text-green-600">
        <IconCircleCheck className="h-4 w-4" />
        <span className="text-sm">Success</span>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2 text-destructive">
      <IconCircleX className="h-4 w-4" />
      <span className="text-sm">Failure</span>
    </div>
  );
}

function ImportResultsCell({ job }: { job: ImportJob }) {
  return (
    <div className="flex gap-1 flex-wrap">
      <Badge variant="outline" className="text-green-600 text-xs">
        +{job.added}
      </Badge>
      {job.updated > 0 && (
        <Badge variant="outline" className="text-blue-600 text-xs">
          ~{job.updated}
        </Badge>
      )}
      {job.skipped > 0 && (
        <Badge variant="outline" className="text-muted-foreground text-xs">
          ={job.skipped}
        </Badge>
      )}
      {job.errors > 0 && (
        <Badge variant="outline" className="text-destructive text-xs">
          !{job.errors}
        </Badge>
      )}
    </div>
  );
}

function ImportActionsCell({ job }: { job: ImportJob }) {
  const navigate = useNavigate();

  const handleViewDetails = () => {
    navigate(`?view=${job.id}`);
  };

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
        <DropdownMenuItem onSelect={handleViewDetails}>
          View Details
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => navigator.clipboard.writeText(job.id)}>
          Copy Import ID
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function getDuration(job: ImportJob): string {
  if (!job.finished_at) return "In progress...";
  const start = new Date(job.started_at);
  const end = new Date(job.finished_at);
  const diff = end.getTime() - start.getTime();
  const seconds = Math.floor(diff / 1000);

  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}m ${remainingSeconds}s`;
}

export function ImportDatasetsDataTable() {
  const navigate = useNavigate();
  const { imports, loading, error, currentPage, totalPages, fetchImports } =
    useImports();

  const [data, setData] = React.useState<ImportJob[]>([]);

  React.useEffect(() => {
    setData(imports || []);
  }, [imports]);

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    [],
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      manifest_url: false,
      message: false,
    });
  const [currentLimit, setCurrentLimit] = React.useState(20);
  const [refreshing, setRefreshing] = React.useState(false);

  const currentPageRef = React.useRef(currentPage);
  const currentLimitRef = React.useRef(currentLimit);

  const fetchImportsData = React.useCallback(
    async (page: number, limit: number, isRefresh = false) => {
      try {
        if (isRefresh) {
          setRefreshing(true);
        }

        const offset = (page - 1) * limit;
        await fetchImports({ limit, offset });

        setCurrentLimit(limit);
      } catch (err) {
        console.error("Error fetching imports:", err);
      } finally {
        setRefreshing(false);
      }
    },
    [fetchImports],
  );

  const columns: ColumnDef<ImportJob>[] = React.useMemo(
    () => [
      {
        accessorKey: "id",
        header: "ID",
        cell: ({ row }) => <ImportIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "dataset",
        header: "Dataset",
        cell: ({ row }) => (
          <div className="font-medium">{row.getValue("dataset")}</div>
        ),
      },
      {
        accessorKey: "version",
        header: "Version",
        cell: ({ row }) => (
          <Badge variant="secondary">{row.getValue("version")}</Badge>
        ),
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => <ImportStatusCell job={row.original} />,
      },
      {
        accessorKey: "started_at",
        header: "Date",
        cell: ({ row }) => {
          const date = new Date(row.getValue("started_at"));
          return (
            <div className="flex items-center gap-1 text-sm">
              <IconCalendar className="h-4 w-4 text-muted-foreground" />
              <span title={format(date, "PPpp", { locale: enUS })}>
                {formatDistanceToNow(date, { addSuffix: true, locale: enUS })}
              </span>
            </div>
          );
        },
      },
      {
        id: "duration",
        header: "Duration",
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground">
            {getDuration(row.original)}
          </span>
        ),
      },
      {
        id: "results",
        header: "Results",
        cell: ({ row }) => <ImportResultsCell job={row.original} />,
      },
      {
        accessorKey: "manifest_url",
        header: "Manifest URL",
        cell: ({ row }) => (
          <div className="text-xs text-muted-foreground truncate max-w-[200px]">
            {row.getValue("manifest_url")}
          </div>
        ),
      },
      {
        accessorKey: "message",
        header: "Message",
        cell: ({ row }) => (
          <div className="text-xs text-muted-foreground truncate max-w-[200px]">
            {row.getValue("message") || "-"}
          </div>
        ),
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => <ImportActionsCell job={row.original} />,
      },
    ],
    [],
  );

  React.useEffect(() => {
    fetchImportsData(currentPage, currentLimit);
  }, [fetchImportsData, currentLimit, currentPage]);

  React.useEffect(() => {
    currentPageRef.current = currentPage;
  }, [currentPage]);

  React.useEffect(() => {
    currentLimitRef.current = currentLimit;
  }, [currentLimit]);

  const isMobile = useIsMobile();

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
    onPaginationChange: (updater) => {
      const newPagination =
        typeof updater === "function"
          ? updater(table.getState().pagination)
          : updater;
      const newPage = newPagination.pageIndex + 1;
      const newLimit = newPagination.pageSize;

      if (newPage !== currentPage || newLimit !== currentLimit) {
        fetchImportsData(newPage, newLimit);
      }
    },
  });

  if (loading && !refreshing && data.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>History</CardTitle>
              <CardDescription>
                View and manage dataset import jobs
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <div className="space-y-3 p-4">
              {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
                <Skeleton key={`hist-skel-${i}`} className="h-10 w-full" />
              ))}
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
          <CardTitle>History</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <IconAlertTriangle className="h-8 w-8 text-red-500 mx-auto" />
              <p className="mt-2 text-sm text-red-600">{error}</p>
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  fetchImportsData(
                    currentPageRef.current,
                    currentLimitRef.current,
                    true,
                  )
                }
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
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <CardTitle>History</CardTitle>
            <CardDescription>
              View and manage dataset import jobs
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
            <Button
              variant="outline"
              size="sm"
              onClick={() =>
                fetchImportsData(
                  currentPageRef.current,
                  currentLimitRef.current,
                  true,
                )
              }
              disabled={refreshing}
            >
              <IconRefresh
                className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
              />
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconColumns className="mr-2 h-4 w-4" />
                  Columns
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
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
                        {column.id}
                      </DropdownMenuCheckboxItem>
                    );
                  })}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {isMobile ? (
          <div className="space-y-3">
            {(data || []).length ? (
              data.map((job) => (
                <button
                  type="button"
                  key={job.id}
                  className="w-full text-left rounded-lg border bg-card p-4 space-y-3 cursor-pointer hover:bg-muted/50"
                  onClick={() => navigate(`?view=${job.id}`)}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-medium text-sm">{job.dataset}</span>
                    <Badge variant="secondary">{job.version}</Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    <ImportStatusCell job={job} />
                    <span className="text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(job.started_at), {
                        addSuffix: true,
                        locale: enUS,
                      })}
                    </span>
                  </div>
                  <ImportResultsCell job={job} />
                </button>
              ))
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                <IconDownload className="mx-auto h-12 w-12 mb-4 opacity-50" />
                <p className="font-medium">No imports found</p>
                <p className="text-sm mt-1">Import a dataset to get started</p>
              </div>
            )}
          </div>
        ) : (
          <div className="rounded-md border">
            <Table>
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
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={() => navigate(`?view=${row.original.id}`)}
                    >
                      {row.getVisibleCells().map((cell) => {
                        const isActionCell = cell.column.id === "actions";
                        return (
                          <TableCell
                            key={cell.id}
                            onClick={
                              isActionCell
                                ? (e) => e.stopPropagation()
                                : undefined
                            }
                          >
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext(),
                            )}
                          </TableCell>
                        );
                      })}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell
                      colSpan={columns.length}
                      className="h-24 text-center"
                    >
                      No imports found.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}

        <DataTablePagination
          pageIndex={currentPage - 1}
          pageCount={totalPages || 1}
          pageSize={currentLimit}
          onPageChange={(pageIndex) =>
            fetchImportsData(pageIndex + 1, currentLimitRef.current)
          }
          onPageSizeChange={(limit) => fetchImportsData(1, limit)}
          loading={loading}
        />
      </CardContent>
    </Card>
  );
}
