"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
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
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { TableSkeletonRows } from "@/components/ui/table-skeleton-rows";
import { useDatasets } from "@/hooks/useDatasets";
import { useGameTemplates } from "@/hooks/useGameTemplates";
import {
  CATEGORY_LABELS,
  FLAG_VARIANT_LABELS,
  MODE_LABELS,
  QUESTION_TYPE_LABELS,
  SCORE_MODE_LABELS,
} from "@/lib/constants/game-template.constants";
import { gameTemplatesService } from "@/lib/services/game-templates.service";
import type {
  CreateGameTemplateRequest,
  GameTemplate,
  TemplateCategory,
  TemplateMode,
  UpdateGameTemplateRequest,
} from "@/lib/types/game-template.types";
import {
  IconDotsVertical,
  IconEdit,
  IconEye,
  IconFilter,
  IconPlus,
  IconRefresh,
  IconRestore,
  IconSearch,
  IconTrash,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import * as React from "react";
import { useNavigate } from "react-router";
import { GameTemplateCreateDialog } from "./game-template-create-dialog";
import { GameTemplateEditDialog } from "./game-template-edit-dialog";

function ActionsCell({
  template,
  onEdit,
  onDelete,
}: {
  template: GameTemplate;
  onEdit: (t: GameTemplate) => void;
  onDelete: (t: GameTemplate) => void;
}) {
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
        <DropdownMenuItem onSelect={() => navigate(`?view=${template.id}`)}>
          <IconEye className="mr-2 h-4 w-4" />
          View
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => onEdit(template)}>
          <IconEdit className="mr-2 h-4 w-4" />
          Edit
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className="text-destructive"
          onSelect={() => onDelete(template)}
        >
          <IconTrash className="mr-2 h-4 w-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

interface Props {
  onTotalCountChange?: (total: number) => void;
  refreshKey?: number;
}

export function GameTemplatesDataTable({
  onTotalCountChange,
  refreshKey,
}: Props) {
  const navigate = useNavigate();
  const {
    templates,
    loading,
    total,
    currentPage,
    totalPages,
    currentLimit,
    fetchTemplates,
    createTemplate,
    updateTemplate,
    deleteTemplate,
  } = useGameTemplates();
  const { datasets } = useDatasets();

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [search, setSearch] = React.useState("");
  const [modeFilter, setModeFilter] = React.useState<TemplateMode | "">("");
  const [categoryFilter, setCategoryFilter] = React.useState<
    TemplateCategory | ""
  >("");
  const [activeFilter, setActiveFilter] = React.useState<boolean | undefined>(
    undefined,
  );
  const [createOpen, setCreateOpen] = React.useState(false);
  const [editTarget, setEditTarget] = React.useState<GameTemplate | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<GameTemplate | null>(
    null,
  );
  const [deleting, setDeleting] = React.useState(false);
  const [seedConfirmOpen, setSeedConfirmOpen] = React.useState(false);
  const [seeding, setSeeding] = React.useState(false);
  const searchTimerRef = React.useRef<ReturnType<typeof setTimeout> | null>(
    null,
  );

  function buildParams(
    overrides?: Partial<Parameters<typeof fetchTemplates>[0]>,
  ) {
    return {
      page: 1,
      limit: currentLimit,
      query: search || undefined,
      mode: (modeFilter || undefined) as TemplateMode | undefined,
      category: (categoryFilter || undefined) as TemplateCategory | undefined,
      active_only: activeFilter,
      ...overrides,
    };
  }

  React.useEffect(() => {
    fetchTemplates({ page: 1, limit: currentLimit });
  }, [fetchTemplates, currentLimit]);

  React.useEffect(() => {
    onTotalCountChange?.(total);
  }, [total, onTotalCountChange]);

  function handleSearchChange(value: string) {
    setSearch(value);
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current);
    searchTimerRef.current = setTimeout(() => {
      fetchTemplates(buildParams({ query: value || undefined }));
    }, 400);
  }

  function handleRefresh() {
    fetchTemplates(buildParams({ page: currentPage }));
  }

  function goToPage(page: number) {
    fetchTemplates(buildParams({ page }));
  }

  function toggleMode(mode: TemplateMode) {
    const next = modeFilter === mode ? "" : mode;
    setModeFilter(next);
    fetchTemplates(buildParams({ mode: next || undefined }));
  }

  function toggleCategory(category: TemplateCategory) {
    const next = categoryFilter === category ? "" : category;
    setCategoryFilter(next);
    fetchTemplates(buildParams({ category: next || undefined }));
  }

  function toggleActive(value: boolean) {
    const next = activeFilter === value ? undefined : value;
    setActiveFilter(next);
    fetchTemplates(buildParams({ active_only: next }));
  }

  function clearFilters() {
    setModeFilter("");
    setCategoryFilter("");
    setActiveFilter(undefined);
    setSearch("");
    fetchTemplates({ page: 1, limit: currentLimit });
  }

  const hasActiveFilters = !!(
    modeFilter ||
    categoryFilter ||
    activeFilter !== undefined ||
    search
  );

  async function handleCreate(data: CreateGameTemplateRequest) {
    return createTemplate(data);
  }

  async function handleUpdate(id: string, data: UpdateGameTemplateRequest) {
    return updateTemplate(id, data);
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    setDeleting(true);
    await deleteTemplate(deleteTarget.id, deleteTarget.name);
    setDeleting(false);
    setDeleteTarget(null);
  }

  async function handleSeedDefaults() {
    setSeeding(true);
    try {
      await gameTemplatesService.seedDefaultTemplates();
      fetchTemplates(buildParams({ page: currentPage }));
    } finally {
      setSeeding(false);
      setSeedConfirmOpen(false);
    }
  }

  const columns: ColumnDef<GameTemplate>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <div>
          <div className="font-medium">{row.original.name}</div>
          <div className="text-xs text-muted-foreground font-mono">
            {row.original.slug}
          </div>
        </div>
      ),
    },
    {
      accessorKey: "mode",
      header: "Mode",
      cell: ({ row }) => {
        const mode = row.original.mode;
        return mode ? (
          <Badge variant="outline">{MODE_LABELS[mode] ?? mode}</Badge>
        ) : (
          <span className="text-muted-foreground text-xs">Custom</span>
        );
      },
    },
    {
      accessorKey: "score_mode",
      header: "Score mode",
      cell: ({ row }) => (
        <Badge
          variant={
            row.original.score_mode === "fastest_wins" ? "secondary" : "outline"
          }
        >
          {SCORE_MODE_LABELS[row.original.score_mode] ??
            row.original.score_mode}
        </Badge>
      ),
    },
    {
      id: "players",
      header: "Players",
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {row.original.min_players}–{row.original.max_players}
        </span>
      ),
    },
    {
      accessorKey: "question_count",
      header: "Questions",
      cell: ({ row }) => (
        <span className="tabular-nums">{row.original.question_count}</span>
      ),
    },
    {
      accessorKey: "points_per_correct",
      header: "Pts / correct",
      cell: ({ row }) => (
        <span className="tabular-nums">{row.original.points_per_correct}</span>
      ),
    },
    {
      id: "dataset",
      header: "Dataset",
      cell: ({ row }) => {
        const ds = datasets.find((d) => d.id === row.original.dataset_id);
        return ds ? (
          <Badge variant="outline" className="font-mono text-xs">
            {ds.name}
          </Badge>
        ) : (
          <span className="text-muted-foreground text-xs">Default</span>
        );
      },
    },
    {
      id: "category",
      header: "Category",
      cell: ({ row }) => {
        const { category, flag_variant, question_type, continent } =
          row.original;
        if (!category)
          return <span className="text-muted-foreground text-xs">—</span>;
        return (
          <div className="flex flex-col gap-0.5">
            <Badge variant="outline">
              {CATEGORY_LABELS[category] ?? category}
            </Badge>
            {flag_variant && (
              <span className="text-xs text-muted-foreground">
                {FLAG_VARIANT_LABELS[flag_variant] ?? flag_variant}
              </span>
            )}
            {question_type && (
              <span className="text-xs text-muted-foreground">
                {QUESTION_TYPE_LABELS[question_type] ?? question_type}
              </span>
            )}
            {continent && (
              <span className="text-xs text-muted-foreground capitalize">
                {continent}
              </span>
            )}
          </div>
        );
      },
    },
    {
      accessorKey: "is_active",
      header: "Status",
      cell: ({ row }) => (
        <Badge variant={row.original.is_active ? "default" : "secondary"}>
          {row.original.is_active ? "Active" : "Inactive"}
        </Badge>
      ),
    },
    {
      id: "actions",
      header: "Actions",
      enableHiding: false,
      cell: ({ row }) => (
        <ActionsCell
          template={row.original}
          onEdit={setEditTarget}
          onDelete={setDeleteTarget}
        />
      ),
    },
  ];

  const table = useReactTable({
    data: templates,
    columns,
    state: { sorting, columnVisibility },
    onSortingChange: setSorting,
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    manualPagination: true,
    pageCount: totalPages,
  });

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>Game Templates</CardTitle>
        <CardDescription>Create and manage game templates</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="relative flex-1 max-w-sm w-full lg:w-auto">
            <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search templates…"
              value={search}
              onChange={(e) => handleSearchChange(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
            <Button
              variant="outline"
              size="sm"
              onClick={handleRefresh}
              disabled={loading}
            >
              <IconRefresh
                className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
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
                      {(modeFilter ? 1 : 0) +
                        (categoryFilter ? 1 : 0) +
                        (activeFilter !== undefined ? 1 : 0)}
                    </Badge>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <div className="px-2 py-1.5 text-sm font-semibold">Mode</div>
                {(["solo", "1v1", "multi"] as TemplateMode[]).map((mode) => (
                  <DropdownMenuItem
                    key={mode}
                    onSelect={(e) => {
                      e.preventDefault();
                      toggleMode(mode);
                    }}
                  >
                    <Checkbox checked={modeFilter === mode} className="mr-2" />
                    {MODE_LABELS[mode] ?? mode}
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Category
                </div>
                {(["general", "flags", "geography"] as TemplateCategory[]).map(
                  (cat) => (
                    <DropdownMenuItem
                      key={cat}
                      onSelect={(e) => {
                        e.preventDefault();
                        toggleCategory(cat);
                      }}
                    >
                      <Checkbox
                        checked={categoryFilter === cat}
                        className="mr-2"
                      />
                      <span className="capitalize">{cat}</span>
                    </DropdownMenuItem>
                  ),
                )}
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm font-semibold">Status</div>
                <DropdownMenuItem
                  onSelect={(e) => {
                    e.preventDefault();
                    toggleActive(true);
                  }}
                >
                  <Checkbox checked={activeFilter === true} className="mr-2" />
                  Active
                </DropdownMenuItem>
                <DropdownMenuItem
                  onSelect={(e) => {
                    e.preventDefault();
                    toggleActive(false);
                  }}
                >
                  <Checkbox checked={activeFilter === false} className="mr-2" />
                  Inactive
                </DropdownMenuItem>
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
            <Button
              variant="outline"
              size="sm"
              onClick={() => setSeedConfirmOpen(true)}
            >
              <IconRestore className="mr-2 h-4 w-4" />
              Restore defaults
            </Button>
            <Button size="sm" onClick={() => setCreateOpen(true)}>
              <IconPlus className="mr-2 h-4 w-4" />
              New template
            </Button>
          </div>
        </div>

        <div className="rounded-md border">
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((hg) => (
                <TableRow key={hg.id}>
                  {hg.headers.map((header) => (
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
              {loading && templates.length === 0 ? (
                <TableSkeletonRows columnCount={columns.length} rowCount={5} />
              ) : table.getRowModel().rows.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    className="cursor-pointer"
                    onClick={(e) => {
                      if (
                        (e.target as HTMLElement).closest(
                          "button, [role='menu'], [role='menuitem']",
                        )
                      )
                        return;
                      navigate(`?view=${row.original.id}`);
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
                    className="h-24 text-center text-muted-foreground"
                  >
                    No templates found.
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
          totalCount={total}
          onPageChange={(page) => goToPage(page + 1)}
          onPageSizeChange={(size) => {
            fetchTemplates({ page: 1, limit: size });
          }}
          loading={loading}
        />

        <GameTemplateCreateDialog
          open={createOpen}
          onOpenChange={setCreateOpen}
          onCreated={handleCreate}
        />
        <GameTemplateEditDialog
          template={editTarget}
          open={!!editTarget}
          onOpenChange={(open) => {
            if (!open) setEditTarget(null);
          }}
          onUpdated={handleUpdate}
        />

        <AlertDialog open={seedConfirmOpen} onOpenChange={setSeedConfirmOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Restore default templates?</AlertDialogTitle>
              <AlertDialogDescription>
                This will add any missing default templates without removing
                existing ones. Templates you created manually will not be
                affected.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={seeding}>Cancel</AlertDialogCancel>
              <AlertDialogAction
                onClick={handleSeedDefaults}
                disabled={seeding}
              >
                {seeding ? "Restoring…" : "Restore defaults"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        <AlertDialog
          open={!!deleteTarget}
          onOpenChange={(open) => {
            if (!open) setDeleteTarget(null);
          }}
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                Delete "{deleteTarget?.name}"?
              </AlertDialogTitle>
              <AlertDialogDescription>
                This action cannot be undone. Existing games that use this
                template will keep their current settings.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={deleting}>Cancel</AlertDialogCancel>
              <AlertDialogAction
                onClick={handleDelete}
                disabled={deleting}
                className="bg-destructive text-white hover:bg-destructive/90"
              >
                {deleting ? "Deleting…" : "Delete"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardContent>
    </Card>
  );
}
