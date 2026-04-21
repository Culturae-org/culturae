"use client";

import { CreateDatasetDrawer } from "@/components/datasets/datasets-create-drawer";
import { ImportDatasetDialog } from "@/components/datasets/datasets-import-dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
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
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { DataTablePagination } from "@/components/ui/data-table-pagination";
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
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { type UnifiedDataset, useDatasetsList } from "@/hooks/useDatasetsList";
import { useIsMobile } from "@/hooks/useMobile";
import {
  IconCalendar,
  IconCheck,
  IconColumns,
  IconDatabase,
  IconDotsVertical,
  IconDownload,
  IconFileText,
  IconFilter,
  IconFlag,
  IconMapPin,
  IconPlus,
  IconRefresh,
  IconSearch,
  IconStar,
  IconWorld,
  IconX,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import * as React from "react";
import { useNavigate } from "react-router";

interface DatasetActionsCellProps {
  dataset: UnifiedDataset;
  onSetDefault: (id: string) => void;
  onToggleActive: (id: string) => void;
  onDeleteClick: (dataset: UnifiedDataset) => void;
  setDatasets: React.Dispatch<React.SetStateAction<UnifiedDataset[]>>;
}

function DatasetActionsCell({
  dataset,
  onSetDefault,
  onToggleActive,
  onDeleteClick,
}: DatasetActionsCellProps) {
  const navigate = useNavigate();

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
          onClick={() => navigate(`?datasetView=${dataset.slug}`)}
        >
          View Details
        </DropdownMenuItem>
        {!dataset.is_default && (
          <DropdownMenuItem onClick={() => onSetDefault(dataset.id)}>
            Set as default
          </DropdownMenuItem>
        )}
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className="text-destructive"
          onClick={() => onDeleteClick(dataset)}
        >
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

interface MobileDatasetCardProps {
  dataset: UnifiedDataset;
  onSetDefault: (id: string) => void;
  onToggleActive: (id: string) => void;
  onDeleteClick: (dataset: UnifiedDataset) => void;
  onViewDetails: (id: string) => void;
  setDatasets: React.Dispatch<React.SetStateAction<UnifiedDataset[]>>;
}

function MobileDatasetCard({
  dataset,
  onSetDefault,
  onToggleActive,
  onDeleteClick,
  onViewDetails,
  setDatasets,
}: MobileDatasetCardProps) {
  return (
    <div className="rounded-lg border bg-card p-4 space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Badge
            variant={dataset.type === "questions" ? "default" : "secondary"}
          >
            {dataset.type === "questions" ? "Questions" : "Geography"}
          </Badge>
          {dataset.is_default && <IconStar className="h-4 w-4" />}
        </div>
        <DatasetActionsCell
          dataset={dataset}
          onSetDefault={onSetDefault}
          onToggleActive={onToggleActive}
          onDeleteClick={onDeleteClick}
          setDatasets={setDatasets}
        />
      </div>
      <div>
        <div className="font-medium">{dataset.name}</div>
        {dataset.description && (
          <div className="text-sm text-muted-foreground line-clamp-1">
            {dataset.description}
          </div>
        )}
      </div>
      <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground pt-2 border-t">
        <span>v{dataset.version}</span>
        <span>{dataset.source}</span>
        <span>
          {dataset.type === "questions"
            ? `${dataset.question_count || 0} questions`
            : `${dataset.country_count || 0} countries`}
        </span>
        <span>{dataset.is_active ? "Active" : "Inactive"}</span>
      </div>
      <div className="pt-2 border-t">
        <Button
          variant="outline"
          size="sm"
          className="w-full"
          onClick={() => onViewDetails(dataset.id)}
        >
          View Details
        </Button>
      </div>
    </div>
  );
}

export function DatasetsDataTable({ onRefresh }: { onRefresh?: () => void }) {
  const navigate = useNavigate();
  const {
    datasets,
    setDatasets,
    loading,
    refreshing,
    error,
    filters,
    search,
    setSearch,
    setFilter,
    clearFilters,
    refresh,
    createDataset,
    toggleActive,
    deleteDataset,
    setDefault,
    checkAllUpdates,
    checkingUpdate,
    // Pagination
    currentPage,
    totalPages,
    totalCount,
    currentLimit,
    goToPage,
    setPageSize,
  } = useDatasetsList();

  const isMobile = useIsMobile();

  const [createDrawerOpen, setCreateDrawerOpen] = React.useState(false);
  const [creating, _setCreating] = React.useState(false);

  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false);
  const [datasetToDelete, setDatasetToDelete] =
    React.useState<UnifiedDataset | null>(null);

  const [importDialogOpen, setImportDialogOpen] = React.useState(false);
  const [importType, setImportType] = React.useState<"questions" | "geography">(
    "questions",
  );

  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      source: false,
      description: false,
      id: false,
      slug: false,
      manifest_url: false,
      update_available: false,
    });

  const handleSetDefault = React.useCallback(
    async (id: string) => {
      await setDefault(id);
    },
    [setDefault],
  );

  const handleToggleActive = React.useCallback(
    async (id: string) => {
      await toggleActive(id);
    },
    [toggleActive],
  );

  const handleDeleteClick = React.useCallback((dataset: UnifiedDataset) => {
    setDatasetToDelete(dataset);
    setDeleteDialogOpen(true);
  }, []);

  const _getDeleteErrorMessage = (error: unknown): string => {
    const message = error instanceof Error ? error.message : String(error);

    if (message.includes("cannot delete the only dataset of this type")) {
      return `You must have at least 2 datasets of the same type to delete one. This is the only ${datasetToDelete?.type} dataset.`;
    }
    if (message.includes("cannot delete the default dataset")) {
      return `This dataset is currently set as default. Please set another ${datasetToDelete?.type} dataset as default first.`;
    }
    return message || "Failed to delete dataset";
  };

  const handleConfirmDelete = async () => {
    if (!datasetToDelete) return;

    try {
      await deleteDataset(datasetToDelete.id, false);
      onRefresh?.();
    } catch (_error) {
    } finally {
      setDeleteDialogOpen(false);
      setDatasetToDelete(null);
    }
  };

  const getDeleteDescription = () => {
    if (!datasetToDelete) return "";
    if (datasetToDelete.is_default) {
      return `This dataset is the default. You must set another ${datasetToDelete.type} dataset as default first before deleting it.`;
    }
    const sameTypeCount = datasets.filter(
      (d) => d.type === datasetToDelete.type && d.id !== datasetToDelete.id,
    ).length;
    if (sameTypeCount === 0) {
      return `This is the last ${datasetToDelete.type} dataset. You must have at least 2 datasets of the same type to delete one.`;
    }
    return `Are you sure you want to delete "${datasetToDelete.name}"? This action cannot be undone.`;
  };

  const handleImportClick = (type: "questions" | "geography") => {
    setImportType(type);
    setImportDialogOpen(true);
  };

  const columns: ColumnDef<UnifiedDataset>[] = React.useMemo(
    () => [
      {
        accessorKey: "type",
        header: "Type",
        cell: ({ row }) => (
          <Badge
            variant={
              row.getValue("type") === "questions" ? "default" : "secondary"
            }
          >
            {row.getValue("type") === "questions" ? (
              <>
                <IconFileText className="h-3 w-3" /> Questions
              </>
            ) : (
              <>
                <IconWorld className="h-3 w-3" /> Geography
              </>
            )}
          </Badge>
        ),
      },
      {
        accessorKey: "id",
        header: "ID",
        cell: ({ row }) => (
          <div
            className="font-mono text-xs truncate max-w-xs"
            title={row.getValue("id")}
          >
            {(row.getValue("id") as string).slice(0, 8)}...
          </div>
        ),
      },
      {
        accessorKey: "name",
        header: "Name",
        cell: ({ row }) => {
          const dataset = row.original;
          return (
            <div className="flex items-center gap-2">
              {dataset.is_default && <IconStar className="h-4 w-4" />}
              <div>
                <div className="font-medium">{dataset.name}</div>
                {dataset.description && (
                  <div className="text-sm text-muted-foreground line-clamp-1">
                    {dataset.description}
                  </div>
                )}
              </div>
            </div>
          );
        },
      },
      {
        accessorKey: "slug",
        header: "Slug",
        cell: ({ row }) => (
          <div
            className="font-mono text-sm truncate max-w-xs"
            title={row.getValue("slug")}
          >
            {row.getValue("slug") as string}
          </div>
        ),
      },
      {
        accessorKey: "version",
        header: "Version",
        cell: ({ row }) => (
          <Badge variant="outline">{row.getValue("version") as string}</Badge>
        ),
      },
      {
        accessorKey: "source",
        header: "Source",
        cell: ({ row }) => (
          <Badge variant="outline">{row.getValue("source") as string}</Badge>
        ),
      },
      {
        accessorKey: "manifest_url",
        header: "Manifest URL",
        cell: ({ row }) => {
          const url = row.getValue("manifest_url") as string;
          return url ? (
            <a
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-600 hover:underline text-xs truncate max-w-xs block"
              title={url}
            >
              {new URL(url).hostname}
            </a>
          ) : (
            <span className="text-muted-foreground text-xs">-</span>
          );
        },
      },
      {
        id: "stats",
        header: "Stats",
        cell: ({ row }) => {
          const dataset = row.original;
          return dataset.type === "questions" ? (
            <div className="flex items-center gap-1 text-sm">
              <IconFileText className="h-4 w-4 text-muted-foreground" />
              {dataset.question_count || 0}
            </div>
          ) : (
            <div className="flex items-center gap-2 text-sm">
              <IconFlag className="h-4 w-4 text-muted-foreground" />
              {dataset.country_count || 0}
              <IconMapPin className="h-4 w-4 text-muted-foreground ml-2" />
              {dataset.continent_count || 0}
            </div>
          );
        },
      },
      {
        accessorKey: "imported_at",
        header: "Imported",
        cell: ({ row }) => (
          <div className="flex items-center gap-1 text-sm text-muted-foreground">
            <IconCalendar className="h-4 w-4" />
            {formatDistanceToNow(new Date(row.getValue("imported_at")), {
              addSuffix: true,
              locale: enUS,
            })}
          </div>
        ),
      },
      {
        accessorKey: "update_available",
        header: "Update",
        cell: ({ row }) => {
          const hasUpdate = row.getValue("update_available");
          return (
            <Badge variant={hasUpdate ? "secondary" : "outline"}>
              {hasUpdate ? (
                <>
                  <IconRefresh className="mr-1 h-3 w-3" /> Available
                </>
              ) : (
                <>Up to date</>
              )}
            </Badge>
          );
        },
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => {
          const isActive = row.original.is_active;
          return (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => handleToggleActive(row.original.id)}
            >
              {isActive ? (
                <IconCheck className="h-4 w-4 text-green-500 mr-1" />
              ) : (
                <IconX className="h-4 w-4 text-muted-foreground mr-1" />
              )}
              <span className="text-sm">
                {isActive ? "Active" : "Inactive"}
              </span>
            </Button>
          );
        },
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => (
          <DatasetActionsCell
            dataset={row.original}
            onSetDefault={handleSetDefault}
            onToggleActive={handleToggleActive}
            onDeleteClick={handleDeleteClick}
            setDatasets={setDatasets}
          />
        ),
      },
    ],
    [handleSetDefault, handleToggleActive, handleDeleteClick, setDatasets],
  );

  const table = useReactTable({
    data: datasets,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    state: { columnVisibility },
  });

  const hasActiveFilters =
    filters.type !== "all" || filters.source !== "" || filters.status !== "all";

  if (loading && !refreshing && datasets.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Datasets</CardTitle>
              <CardDescription>
                Manage imported datasets from manifest
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <div className="space-y-3 p-4">
              {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
                <Skeleton key={`ds-skel-${i}`} className="h-10 w-full" />
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
          <CardTitle>Datasets</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <Alert variant="destructive" className="mb-2">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
              <Button variant="outline" size="sm" onClick={refresh}>
                Retry
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card className="border-0 dark:border">
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div>
              <CardTitle>All Datasets</CardTitle>
              <CardDescription>
                Manage imported datasets from manifest
              </CardDescription>
            </div>
            <div className="flex flex-col sm:flex-row gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={checkAllUpdates}
                disabled={checkingUpdate}
              >
                <IconRefresh
                  className={`h-4 w-4 ${checkingUpdate ? "animate-spin" : ""}`}
                />
                Check All Updates
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleImportClick("geography")}
              >
                <IconWorld className="mr-2 h-4 w-4" />
                Import Geography
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleImportClick("questions")}
              >
                <IconDownload className="mr-2 h-4 w-4" />
                Import Questions
              </Button>
              <Button
                size="sm"
                onClick={() => {
                  setCreateDrawerOpen(true);
                }}
              >
                <IconPlus className="mr-2 h-4 w-4" />
                New Dataset
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
            <div className="relative flex-1 max-w-sm w-full lg:w-auto">
              <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search datasets..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
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
                  <Button variant="outline" size="sm">
                    <IconFilter className="h-4 w-4" />
                    Filter
                    {hasActiveFilters && (
                      <Badge
                        variant="secondary"
                        className="ml-1 h-5 px-1 text-xs"
                      >
                        {[
                          filters.type !== "all" ? 1 : 0,
                          filters.source ? 1 : 0,
                          filters.status !== "all" ? 1 : 0,
                        ].reduce((a, b) => a + b, 0)}
                      </Badge>
                    )}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <div className="px-2 py-1.5 text-sm font-semibold">Type</div>
                  {["questions", "geography"].map((t) => (
                    <DropdownMenuItem
                      key={t}
                      onSelect={(e) => e.preventDefault()}
                    >
                      <Checkbox
                        checked={filters.type === t}
                        onCheckedChange={(checked) =>
                          setFilter(
                            "type",
                            checked ? (t as "questions" | "geography") : "all",
                          )
                        }
                        className="mr-2"
                      />
                      {t.charAt(0).toUpperCase() + t.slice(1)}
                    </DropdownMenuItem>
                  ))}
                  <DropdownMenuSeparator />
                  <div className="px-2 py-1.5 text-sm font-semibold">
                    Default
                  </div>
                  {["default"].map((s) => (
                    <DropdownMenuItem
                      key={s}
                      onSelect={(e) => e.preventDefault()}
                    >
                      <Checkbox
                        checked={filters.status === s}
                        onCheckedChange={(checked) =>
                          setFilter(
                            "status",
                            checked ? (s as "default") : "all",
                          )
                        }
                        className="mr-2"
                      />
                      Default only
                    </DropdownMenuItem>
                  ))}
                  {hasActiveFilters && (
                    <>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onSelect={clearFilters}>
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
                <DropdownMenuContent align="end">
                  {table
                    .getAllColumns()
                    .filter((col) => col.getCanHide())
                    .map((col) => (
                      <DropdownMenuCheckboxItem
                        key={col.id}
                        className="capitalize"
                        checked={col.getIsVisible()}
                        onCheckedChange={(value) =>
                          col.toggleVisibility(!!value)
                        }
                      >
                        {col.id}
                      </DropdownMenuCheckboxItem>
                    ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>

          {datasets.length === 0 && !loading && !refreshing ? (
            <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
              <IconDatabase className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p className="font-medium">No datasets found</p>
              <p className="text-sm mt-1">Import a dataset to get started</p>
            </div>
          ) : isMobile ? (
            <div className="space-y-3">
              {datasets.map((dataset) => (
                <MobileDatasetCard
                  key={`${dataset.type}-${dataset.id}`}
                  dataset={dataset}
                  onSetDefault={handleSetDefault}
                  onToggleActive={handleToggleActive}
                  onDeleteClick={handleDeleteClick}
                  onViewDetails={(slug) => navigate(`?datasetView=${slug}`)}
                  setDatasets={setDatasets}
                />
              ))}
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
                      <TableRow
                        key={row.id}
                        onClick={() =>
                          navigate(`?datasetView=${row.original.slug}`)
                        }
                        className="cursor-pointer hover:bg-muted/50 transition-colors"
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
                        No datasets found.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>

        <CardContent className="pt-2">
          <DataTablePagination
            pageIndex={currentPage - 1}
            pageCount={totalPages}
            pageSize={currentLimit}
            totalCount={totalCount}
            onPageChange={(page) => goToPage(page + 1)}
            onPageSizeChange={setPageSize}
            loading={loading}
          />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Delete Dataset"
        description={getDeleteDescription()}
        onConfirm={handleConfirmDelete}
        confirmText="Delete Dataset"
        variant="destructive"
      />

      <ImportDatasetDialog
        open={importDialogOpen}
        onOpenChange={setImportDialogOpen}
        defaultType={importType}
        onImportComplete={() => {
          setImportDialogOpen(false);
          refresh();
          onRefresh?.();
        }}
      />

      <CreateDatasetDrawer
        open={createDrawerOpen}
        onOpenChange={setCreateDrawerOpen}
        onCreate={createDataset}
        creating={creating}
      />
    </>
  );
}
