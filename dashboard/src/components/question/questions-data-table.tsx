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
import { useIsMobile } from "@/hooks/useMobile";
import { useQuestions } from "@/hooks/useQuestions";
import { questionsService } from "@/lib/services/questions.service";
import {
  IconAlertTriangle,
  IconColumns,
  IconDotsVertical,
  IconFilter,
  IconRefresh,
  IconSearch,
  IconTrash,
  IconX,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type ColumnFiltersState,
  type FilterFn,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import * as React from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { QuestionCreateDialog } from "./question-create-dialog";
import { QuestionEditDialog } from "./question-edit-dialog";

import type { Question } from "@/lib/types/question.types";

interface QuestionsDataTableProps {
  onTotalCountChange?: (count: number) => void;
  datasetId?: string;
}

function QuestionIdCell({ id }: { id: string }) {
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

function QuestionActionsCell({
  question,
  onQuestionUpdated,
  onDeleteClick,
}: {
  question: Question;
  onQuestionUpdated: (q: Question) => void;
  onDeleteClick: (q: Question) => void;
}) {
  const [editOpen, setEditOpen] = React.useState(false);
  const navigate = useNavigate();

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="h-8 w-8 p-0">
            <span className="sr-only">Open menu</span>
            <IconDotsVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel>Actions</DropdownMenuLabel>
          <DropdownMenuItem onSelect={() => navigate(`?view=${question.id}`)}>
            View question
          </DropdownMenuItem>
          <DropdownMenuItem onSelect={() => setEditOpen(true)}>
            Edit question
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => navigator.clipboard.writeText(question.id)}
          >
            Copy question ID
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            className="text-destructive"
            onClick={() => onDeleteClick(question)}
          >
            <IconTrash className="mr-2 h-4 w-4" />
            Delete question
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <QuestionEditDialog
        question={question}
        onQuestionUpdated={onQuestionUpdated}
        open={editOpen}
        onOpenChange={setEditOpen}
      />
    </>
  );
}

export function QuestionsDataTable({
  onTotalCountChange,
  datasetId,
}: QuestionsDataTableProps = {}) {
  const {
    questions: apiQuestions,
    loading,
    error,
    currentPage,
    totalPages,
    fetchQuestions: fetchQuestionsHook,
  } = useQuestions(datasetId);

  const [data, setData] = React.useState<Question[]>([]);

  React.useEffect(() => {
    setData(apiQuestions || []);
  }, [apiQuestions]);

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    [],
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      tags: false,
      kind: false,
      qtype: true,
      version: false,
      estimated_seconds: false,
      created_at: false,
    });
  const [rowSelection, setRowSelection] = React.useState({});
  const [globalFilter, setGlobalFilter] = React.useState("");
  const [currentLimit, setCurrentLimit] = React.useState(10);
  const [searchQuery, setSearchQuery] = React.useState<string>("");
  const [debouncedSearchQuery, setDebouncedSearchQuery] =
    React.useState<string>("");
  const [refreshing, setRefreshing] = React.useState(false);
  const [filterDifficulty, setFilterDifficulty] = React.useState("");
  const [filterQtype, setFilterQtype] = React.useState("");

  const [overallTotal, setOverallTotal] = React.useState(0);

  const currentPageRef = React.useRef(currentPage);
  const currentLimitRef = React.useRef(currentLimit);
  const searchQueryRef = React.useRef(searchQuery);
  const debouncedSearchQueryRef = React.useRef(debouncedSearchQuery);
  const filterDifficultyRef = React.useRef(filterDifficulty);
  const filterQtypeRef = React.useRef(filterQtype);

  const fetchQuestions = React.useCallback(
    async (page: number, limit: number, query?: string, isRefresh = false) => {
      try {
        if (isRefresh) {
          setRefreshing(true);
        }

        const result = await fetchQuestionsHook({
          page,
          limit,
          search: query,
          dataset_id: datasetId,
          difficulty: filterDifficultyRef.current || undefined,
          qtype: filterQtypeRef.current || undefined,
        });

        if (onTotalCountChange) {
          onTotalCountChange(result.total);
        }

        setCurrentLimit(result.limit || limit);
      } catch (err) {
        console.error("Error fetching questions:", err);
      } finally {
        setRefreshing(false);
      }
    },
    [fetchQuestionsHook, onTotalCountChange, datasetId],
  );

  const updateQuestionInList = React.useCallback(
    (updatedQuestion: Question) => {
      setData((prevData) =>
        prevData.map((q) =>
          q.id === updatedQuestion.id ? updatedQuestion : q,
        ),
      );
    },
    [],
  );

  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false);
  const [questionToDelete, setQuestionToDelete] =
    React.useState<Question | null>(null);
  const { deleteQuestion } = useQuestions(datasetId);

  const handleDeleteClick = React.useCallback((question: Question) => {
    setQuestionToDelete(question);
    setDeleteDialogOpen(true);
  }, []);

  const handleConfirmDelete = async () => {
    if (questionToDelete) {
      try {
        await deleteQuestion(questionToDelete.id);
        toast.success("Question deleted successfully");
        setOverallTotal((prev) => Math.max(0, prev - 1));
        fetchQuestions(
          currentPageRef.current,
          currentLimitRef.current,
          debouncedSearchQueryRef.current,
        );
      } catch (error) {
        console.error("Failed to delete question:", error);
        toast.error("Failed to delete question", {
          description:
            error instanceof Error ? error.message : "An error occurred",
        });
      }
    }
    setDeleteDialogOpen(false);
    setQuestionToDelete(null);
  };

  const columns: ColumnDef<Question>[] = React.useMemo(
    () => [
      {
        id: "select",
        header: ({ table }) => (
          <Checkbox
            checked={
              table.getIsAllPageRowsSelected() ||
              (table.getIsSomePageRowsSelected() && "indeterminate")
            }
            onCheckedChange={(value) =>
              table.toggleAllPageRowsSelected(!!value)
            }
            aria-label="Select all"
          />
        ),
        cell: ({ row }) => (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
          />
        ),
        enableSorting: false,
        enableHiding: false,
      },
      {
        accessorKey: "id",
        header: "ID",
        cell: ({ row }) => <QuestionIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "slug",
        header: "Slug",
        cell: ({ row }) => (
          <div className="font-medium">{row.getValue("slug")}</div>
        ),
      },
      {
        accessorKey: "difficulty",
        header: "Difficulty",
        cell: ({ row }) => {
          const difficulty = row.getValue("difficulty") as string;
          const difficultyColors = {
            beginner:
              "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
            intermediate:
              "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
            advanced:
              "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
          };
          return (
            <Badge
              variant="secondary"
              className={
                difficultyColors[difficulty as keyof typeof difficultyColors] ||
                "bg-gray-100 text-gray-800"
              }
            >
              {difficulty}
            </Badge>
          );
        },
      },
      {
        accessorKey: "theme.slug",
        header: "Theme",
        cell: ({ row }) => (
          <div className="text-muted-foreground">{row.original.theme.slug}</div>
        ),
      },
      {
        accessorKey: "subthemes",
        header: "Subthemes",
        cell: ({ row }) => {
          const subthemes = row.original.subthemes;
          if (!subthemes || subthemes.length === 0)
            return <span className="text-muted-foreground">-</span>;
          return (
            <div className="flex flex-wrap gap-1">
              {subthemes.slice(0, 2).map((st) => (
                <Badge key={st.slug} variant="outline" className="text-xs">
                  {st.slug}
                </Badge>
              ))}
              {subthemes.length > 2 && (
                <Badge variant="outline" className="text-xs">
                  +{subthemes.length - 2}
                </Badge>
              )}
            </div>
          );
        },
      },
      {
        accessorKey: "tags",
        header: "Tags",
        cell: ({ row }) => {
          const tags = row.original.tags;
          if (!tags || tags.length === 0)
            return <span className="text-muted-foreground">-</span>;
          return (
            <div className="flex flex-wrap gap-1">
              {tags.slice(0, 2).map((t) => (
                <Badge key={t.slug} variant="secondary" className="text-xs">
                  {t.slug}
                </Badge>
              ))}
              {tags.length > 2 && (
                <Badge variant="secondary" className="text-xs">
                  +{tags.length - 2}
                </Badge>
              )}
            </div>
          );
        },
      },
      {
        accessorKey: "kind",
        header: "Kind",
        cell: ({ row }) => (
          <div className="text-muted-foreground capitalize">
            {row.getValue("kind")}
          </div>
        ),
      },
      {
        accessorKey: "qtype",
        header: "Type",
        cell: ({ row }) => {
          const qtype = row.getValue("qtype") as string;
          const qtypeConfig: Record<
            string,
            { label: string; className: string }
          > = {
            single_choice: {
              label: "Single Choice",
              className:
                "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
            },
            "true-false": {
              label: "True/False",
              className:
                "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300",
            },
            multiple_choice: {
              label: "Multiple Choice",
              className:
                "bg-cyan-100 text-cyan-800 dark:bg-cyan-900 dark:text-cyan-300",
            },
          };
          const config = qtypeConfig[qtype] || {
            label: qtype.replace("-", " ").replace("_", " "),
            className: "bg-gray-100 text-gray-800",
          };
          return (
            <Badge variant="secondary" className={config.className}>
              {config.label}
            </Badge>
          );
        },
      },
      {
        accessorKey: "version",
        header: "Version",
        cell: ({ row }) => (
          <div className="text-muted-foreground">{row.getValue("version")}</div>
        ),
      },
      {
        accessorKey: "estimated_seconds",
        header: "Time (s)",
        cell: ({ row }) => (
          <div className="text-center">{row.getValue("estimated_seconds")}</div>
        ),
      },
      {
        accessorKey: "created_at",
        header: "Created",
        cell: ({ row }) => {
          const date = new Date(row.getValue("created_at"));
          return (
            <div className="text-sm text-muted-foreground">
              {date.toLocaleDateString()}
            </div>
          );
        },
      },
      {
        id: "actions",
        enableHiding: false,
        header: "Actions",
        cell: ({ row }) => {
          const question = row.original;
          return (
            <QuestionActionsCell
              question={question}
              onQuestionUpdated={updateQuestionInList}
              onDeleteClick={handleDeleteClick}
            />
          );
        },
      },
    ],
    [handleDeleteClick, updateQuestionInList],
  );

  React.useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  React.useEffect(() => {
    if (datasetId) {
      fetchQuestions(1, 10, "");
    }
  }, [datasetId, fetchQuestions]);

  React.useEffect(() => {
    fetchQuestions(1, currentLimitRef.current, debouncedSearchQueryRef.current);
  }, [fetchQuestions]);

  React.useEffect(() => {
    fetchQuestions(currentPage, currentLimit, debouncedSearchQuery);
  }, [currentLimit, currentPage, debouncedSearchQuery, fetchQuestions]);

  React.useEffect(() => {
    currentPageRef.current = currentPage;
  }, [currentPage]);

  React.useEffect(() => {
    currentLimitRef.current = currentLimit;
  }, [currentLimit]);

  React.useEffect(() => {
    searchQueryRef.current = searchQuery;
  }, [searchQuery]);

  React.useEffect(() => {
    debouncedSearchQueryRef.current = debouncedSearchQuery;
  }, [debouncedSearchQuery]);

  React.useEffect(() => {
    filterDifficultyRef.current = filterDifficulty;
  }, [filterDifficulty]);

  React.useEffect(() => {
    filterQtypeRef.current = filterQtype;
  }, [filterQtype]);

  React.useEffect(() => {
    const fetchOverallTotal = async () => {
      try {
        const result = await questionsService.getQuestions({ limit: 1 });
        setOverallTotal(result.total);
      } catch (err) {
        console.error("Error fetching overall total:", err);
      }
    };
    fetchOverallTotal();
  }, []);

  const isMobile = useIsMobile();

  const globalFilterFn: FilterFn<Question> = () => {
    return true;
  };

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn,
    manualPagination: true,
    pageCount: totalPages,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
      globalFilter,
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
        fetchQuestions(newPage, newLimit, debouncedSearchQueryRef.current);
      }
    },
  });

  if (loading && !refreshing && data.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Questions</CardTitle>
          <CardDescription>Manage and view quiz questions</CardDescription>
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
                  <Skeleton key={`q-skel-${i}`} className="h-10 w-full" />
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
          <CardTitle>Questions</CardTitle>
          <CardDescription>Manage and view quiz questions</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <IconAlertTriangle className="h-8 w-8 text-red-500 mx-auto" />
              <p className="mt-2 text-sm text-destructive">{error}</p>
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  fetchQuestions(
                    currentPageRef.current,
                    currentLimitRef.current,
                    debouncedSearchQueryRef.current,
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
        <CardTitle>Questions</CardTitle>
        <CardDescription>Manage and view quiz questions</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="relative flex-1 max-w-sm w-full lg:w-auto">
            <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search by slug, UUID, tag, subtheme..."
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              className="pl-10 pr-10"
            />
            {searchQuery && (
              <IconX
                className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground cursor-pointer"
                onClick={() => setSearchQuery("")}
              />
            )}
          </div>
          <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
            <Button
              variant="outline"
              size="sm"
              onClick={() =>
                fetchQuestions(
                  currentPageRef.current,
                  currentLimitRef.current,
                  debouncedSearchQueryRef.current,
                  true,
                )
              }
              disabled={refreshing}
            >
              <IconRefresh
                className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
              />
            </Button>
            <QuestionCreateDialog
              datasetId={datasetId}
              onQuestionCreated={() =>
                fetchQuestions(
                  currentPageRef.current,
                  currentLimitRef.current,
                  debouncedSearchQueryRef.current,
                  true,
                )
              }
            />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconFilter className="h-4 w-4" />
                  Filter
                  {(filterDifficulty || filterQtype) && (
                    <Badge
                      variant="secondary"
                      className="ml-1 h-5 px-1 text-xs"
                    >
                      {(filterDifficulty ? 1 : 0) + (filterQtype ? 1 : 0)}
                    </Badge>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>Difficulty</DropdownMenuLabel>
                {(["beginner", "intermediate", "advanced"] as const).map(
                  (d) => (
                    <DropdownMenuItem
                      key={d}
                      onSelect={(e) => e.preventDefault()}
                    >
                      <Checkbox
                        checked={filterDifficulty === d}
                        onCheckedChange={(checked) => {
                          setFilterDifficulty(checked ? d : "");
                        }}
                        className="mr-2"
                      />
                      {d.charAt(0).toUpperCase() + d.slice(1)}
                    </DropdownMenuItem>
                  ),
                )}
                <DropdownMenuSeparator />
                <DropdownMenuLabel>Question type</DropdownMenuLabel>
                {[
                  { value: "single_choice", label: "Single choice" },
                  { value: "multiple_choice", label: "Multiple choice" },
                  { value: "true-false", label: "True / False" },
                ].map(({ value, label }) => (
                  <DropdownMenuItem
                    key={value}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={filterQtype === value}
                      onCheckedChange={(checked) => {
                        setFilterQtype(checked ? value : "");
                      }}
                      className="mr-2"
                    />
                    {label}
                  </DropdownMenuItem>
                ))}
                {(filterDifficulty || filterQtype) && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      onSelect={() => {
                        setFilterDifficulty("");
                        setFilterQtype("");
                      }}
                    >
                      Clear filters
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
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
                  .map((column) => (
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
                  ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {isMobile ? (
          <div className="space-y-3">
            {(data || []).length ? (
              data.map((question) => {
                const difficultyColors: Record<string, string> = {
                  beginner:
                    "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
                  intermediate:
                    "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
                  advanced:
                    "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
                };
                return (
                  <div
                    key={question.id}
                    className="rounded-lg border bg-card p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-sm">
                        {question.slug}
                      </span>
                      <QuestionActionsCell
                        question={question}
                        onQuestionUpdated={updateQuestionInList}
                        onDeleteClick={handleDeleteClick}
                      />
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge
                        variant="secondary"
                        className={
                          difficultyColors[question.difficulty] ||
                          "bg-gray-100 text-gray-800"
                        }
                      >
                        {question.difficulty}
                      </Badge>
                      <span className="text-sm text-muted-foreground">
                        {question.theme.slug}
                      </span>
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                No questions found.
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
                      data-state={row.getIsSelected() && "selected"}
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
                      No questions found.
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
          totalCount={overallTotal}
          onPageChange={(page) =>
            fetchQuestions(
              page + 1,
              currentLimitRef.current,
              debouncedSearchQueryRef.current,
            )
          }
          onPageSizeChange={(newSize) => {
            setCurrentLimit(newSize);
            fetchQuestions(1, newSize, debouncedSearchQueryRef.current);
          }}
          loading={loading}
        />

        <ConfirmDialog
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          title="Delete Question"
          description={`Are you sure you want to delete the question "${questionToDelete?.slug}"? This action cannot be undone.`}
          onConfirm={handleConfirmDelete}
          confirmText="Delete"
          variant="destructive"
        />
      </CardContent>
    </Card>
  );
}
