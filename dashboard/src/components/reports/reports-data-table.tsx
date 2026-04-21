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
import { useIsMobile } from "@/hooks/useMobile";
import { useReports } from "@/hooks/useReports";
import type { Report, ReportStatus } from "@/lib/types/reports.types";
import { IconDotsVertical, IconRefresh } from "@tabler/icons-react";
import { IconFilter } from "@tabler/icons-react";
import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { format } from "date-fns";
import { useNavigate } from "react-router";

interface ReportsDataTableProps {
  onViewRequest?: (report: Report) => void;
}

export function ReportsDataTable({
  onViewRequest,
}: ReportsDataTableProps = {}) {
  const {
    reports,
    loading,
    currentPage,
    totalPages,
    totalCount,
    currentLimit,
    filters,
    goToPage,
    setPageSize,
    refresh,
    setFilter,
  } = useReports();

  const navigate = useNavigate();
  const isMobile = useIsMobile();

  const getStatusColor = (status: ReportStatus) => {
    switch (status) {
      case "pending":
        return "default";
      case "in_progress":
        return "secondary";
      case "resolved":
        return "outline";
      default:
        return "outline";
    }
  };

  const columns: ColumnDef<Report>[] = [
    {
      accessorKey: "created_at",
      header: "Date",
      cell: ({ row }) =>
        format(new Date(row.getValue("created_at")), "MMM d, yyyy HH:mm"),
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => (
        <Badge
          variant={getStatusColor(row.getValue("status"))}
          className="capitalize"
        >
          {row.getValue("status")}
        </Badge>
      ),
    },
    {
      accessorKey: "reason",
      header: "Reason",
      cell: ({ row }) => {
        const reason = row.getValue("reason");
        return (
          <span className="capitalize">
            {typeof reason === "string" ? reason.replace(/_/g, " ") : ""}
          </span>
        );
      },
    },
    {
      accessorKey: "user_id",
      header: "Reporter",
      cell: ({ row }) => {
        const id = row.getValue("user_id");
        return (
          <span className="font-mono text-xs">
            {typeof id === "string" ? id.slice(0, 8) : String(id)}...
          </span>
        );
      },
    },
    {
      id: "question_ref",
      header: "Question",
      cell: ({ row }) => {
        const report = row.original;
        const id = report.question_id || report.game_question_id;
        if (!id)
          return <span className="text-muted-foreground text-xs">-</span>;
        const label = report.question_id ? "" : "gq:";
        return (
          <span className="font-mono text-xs">
            {label}
            {id.slice(0, 8)}…
          </span>
        );
      },
    },
    {
      id: "actions",
      header: "Actions",
      enableHiding: false,
      cell: ({ row }) => {
        const report = row.original;
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <IconDotsVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Actions</DropdownMenuLabel>
              <DropdownMenuItem
                onClick={() => {
                  navigate(`?view=${report.id}`);
                  onViewRequest?.(report);
                }}
              >
                View Details
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => {
                  navigate(`?view=${report.id}`);
                  onViewRequest?.(report);
                }}
              >
                Update Status
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];

  const table = useReactTable({
    data: reports,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: totalPages,
  });

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle>Reports</CardTitle>
        <CardDescription>Manage user reports and violations</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
            <Button
              variant="outline"
              size="sm"
              onClick={() => refresh()}
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
                </Button>
              </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <div className="px-2 py-1.5 text-sm font-semibold">
                Status Filter
              </div>
              {["pending", "in_progress", "resolved"].map((s) => (
                <DropdownMenuItem key={s} onSelect={(e) => e.preventDefault()}>
                  <Checkbox
                    checked={filters.status === s}
                    onCheckedChange={(checked) =>
                      setFilter("status", checked ? s : "")
                    }
                    className="mr-2"
                  />
                  <span className="capitalize">{s}</span>
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {isMobile ? (
          <div className="space-y-3">
            {reports.length
              ? reports.map((report) => (
                  <div
                    key={report.id}
                    className="rounded-lg border bg-card p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-medium capitalize">
                        {report.reason?.replace(/_/g, " ")}
                      </span>
                      <div className="flex items-center gap-2">
                        <Badge
                          variant={getStatusColor(report.status)}
                          className="capitalize"
                        >
                          {report.status}
                        </Badge>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="h-8 w-8 p-0">
                              <IconDotsVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => {
                                navigate(`?view=${report.id}`);
                                onViewRequest?.(report);
                              }}
                            >
                              View Details
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </div>
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <span>
                        {format(new Date(report.created_at), "MMM d, yyyy")}
                      </span>
                      <span>•</span>
                      <span className="font-mono text-xs">
                        {typeof report.user_id === "string"
                          ? report.user_id.slice(0, 8)
                          : String(report.user_id)}
                        ...
                      </span>
                    </div>
                  </div>
                ))
              : "No results."}
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
                      {loading ? (
                        <div className="space-y-2">
                          {[0, 1, 2, 3, 4].map((i) => (
                            <Skeleton
                              key={`report-skel-${i}`}
                              className="h-4 w-full"
                            />
                          ))}
                        </div>
                      ) : (
                        "No results."
                      )}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}

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
  );
}
