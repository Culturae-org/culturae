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
import { useUserConnectionLogs } from "@/hooks/useLogs";
import { useIsMobile } from "@/hooks/useMobile";
import type { ConnectionLog } from "@/lib/types/user.types";
import {
  IconAlertTriangle,
  IconCalendar,
  IconCheck,
  IconColumns,
  IconCopy,
  IconDeviceDesktop,
  IconDotsVertical,
  IconFilter,
  IconMapPin,
  IconRefresh,
} from "@tabler/icons-react";
import {
  type ColumnDef,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { format, formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import * as React from "react";

interface UserConnectionLogsDataTableProps {
  userId: string;
}

function LogIdCell({ id }: { id: string | undefined }) {
  const [isExpanded, setIsExpanded] = React.useState(false);

  if (!id) return <span className="text-muted-foreground">-</span>;

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

function StatusCell({ isSuccess }: { isSuccess: boolean | undefined }) {
  if (isSuccess === undefined) return <Badge variant="outline">Unknown</Badge>;
  return (
    <Badge variant={isSuccess ? "default" : "destructive"}>
      {isSuccess ? "Success" : "Failed"}
    </Badge>
  );
}

function IpCell({ ip }: { ip: string | undefined }) {
  const [copied, setCopied] = React.useState(false);

  if (!ip) return <span className="text-muted-foreground">-</span>;

  const cleanIp = ip.replace(/^::ffff:/, "");

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(cleanIp);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  return (
    <div className="flex items-center gap-1">
      <span className="text-sm font-mono">{cleanIp}</span>
      <Button
        variant="ghost"
        size="sm"
        className="h-6 w-6 p-0"
        onClick={copyToClipboard}
      >
        {copied ? (
          <IconCheck className="h-3 w-3 text-green-500" />
        ) : (
          <IconCopy className="h-3 w-3" />
        )}
      </Button>
    </div>
  );
}

export function UserConnectionLogsDataTable({
  userId,
}: UserConnectionLogsDataTableProps) {
  const { logs, loading, refreshing, error, refresh } =
    useUserConnectionLogs(userId);
  const isMobile = useIsMobile();

  const [logFilter, setLogFilter] = React.useState<
    "all" | "success" | "failed"
  >("all");
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      session_id: false,
      device_info: false,
      user_agent: false,
    });
  const [rowSelection, setRowSelection] = React.useState({});
  const [currentPage, setCurrentPage] = React.useState(1);
  const [currentLimit, setCurrentLimit] = React.useState(10);

  const filteredData = React.useMemo(() => {
    if (logFilter === "success")
      return logs.filter((l) => l.is_success === true);
    if (logFilter === "failed")
      return logs.filter((l) => l.is_success === false);
    return logs;
  }, [logs, logFilter]);

  const columns: ColumnDef<ConnectionLog>[] = React.useMemo(
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
        cell: ({ row }) => <LogIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "session_id",
        header: "Session ID",
        cell: ({ row }) => {
          const sessionId = row.getValue("session_id") as string | undefined;
          return sessionId ? (
            <LogIdCell id={sessionId} />
          ) : (
            <span className="text-muted-foreground">-</span>
          );
        },
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => <StatusCell isSuccess={row.original.is_success} />,
      },
      {
        accessorKey: "ip_address",
        header: "IP Address",
        cell: ({ row }) => <IpCell ip={row.getValue("ip_address")} />,
      },
      {
        accessorKey: "location",
        header: "Location",
        cell: ({ row }) => {
          const location = row.getValue("location") as string | undefined;
          return location ? (
            <div className="flex items-center gap-1 text-sm">
              <IconMapPin className="h-3 w-3 text-muted-foreground" />
              {location}
            </div>
          ) : (
            <span className="text-muted-foreground text-sm">-</span>
          );
        },
      },
      {
        accessorKey: "user_agent",
        header: "User Agent",
        cell: ({ row }) => {
          const userAgent = row.getValue("user_agent") as string;
          return (
            <div className="flex items-center gap-1 text-sm truncate max-w-[200px]">
              <IconDeviceDesktop className="h-3 w-3 text-muted-foreground flex-shrink-0" />
              <span className="truncate" title={userAgent}>
                {userAgent || "Unknown"}
              </span>
            </div>
          );
        },
      },
      {
        accessorKey: "device_info",
        header: "Device Info",
        cell: ({ row }) => {
          const deviceInfo = row.getValue("device_info") as
            | Record<string, unknown>
            | undefined;
          return deviceInfo ? (
            <div
              className="text-xs text-muted-foreground truncate max-w-[150px]"
              title={JSON.stringify(deviceInfo)}
            >
              {JSON.stringify(deviceInfo)}
            </div>
          ) : (
            <span className="text-muted-foreground text-sm">-</span>
          );
        },
      },
      {
        accessorKey: "failure_reason",
        header: "Failure Reason",
        cell: ({ row }) => {
          const reason = row.getValue("failure_reason") as string | undefined;
          return reason ? (
            <div
              className="text-sm text-red-600 truncate max-w-[200px]"
              title={reason}
            >
              {reason}
            </div>
          ) : (
            <span className="text-muted-foreground text-sm">-</span>
          );
        },
      },
      {
        accessorKey: "created_at",
        header: "Date",
        cell: ({ row }) => {
          const dateValue = row.getValue("created_at") as string;
          if (
            !dateValue ||
            dateValue === "0001-01-01T00:00:00Z" ||
            Number.isNaN(new Date(dateValue).getTime())
          ) {
            return (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <IconCalendar className="h-4 w-4" />
                <span>-</span>
              </div>
            );
          }
          const date = new Date(dateValue);
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
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => {
          const log = row.original;
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
                  onClick={() => navigator.clipboard.writeText(log.id)}
                >
                  Copy Log ID
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => navigator.clipboard.writeText(log.ip_address)}
                >
                  Copy IP Address
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          );
        },
      },
    ],
    [],
  );

  const table = useReactTable({
    data: filteredData,
    columns,
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    pageCount: Math.ceil(filteredData.length / currentLimit),
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      pagination: { pageIndex: currentPage - 1, pageSize: currentLimit },
    },
    onPaginationChange: (updater) => {
      const newPagination =
        typeof updater === "function"
          ? updater(table.getState().pagination)
          : updater;
      setCurrentPage(newPagination.pageIndex + 1);
      setCurrentLimit(newPagination.pageSize);
    },
  });

  if (loading && !refreshing && logs.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>Connection Logs</CardTitle>
          <CardDescription>
            User authentication and connection history
          </CardDescription>
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
                  <Skeleton key={`conn-skel-${i}`} className="h-10 w-full" />
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
          <CardTitle>Connection Logs</CardTitle>
          <CardDescription>
            User authentication and connection history
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <IconAlertTriangle className="h-8 w-8 text-red-500 mx-auto" />
              <p className="mt-2 text-sm text-destructive">{error}</p>
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
        <CardTitle>Connection Logs</CardTitle>
        <CardDescription>
          User authentication and connection history
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="flex flex-wrap gap-2 w-full lg:w-auto">
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
                  <IconFilter className="h-4 w-4" />
                  Filter
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>Status Filter</DropdownMenuLabel>
                {(["all", "success", "failed"] as const).map((f) => (
                  <DropdownMenuItem
                    key={f}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={logFilter === f}
                      onCheckedChange={(checked) =>
                        setLogFilter(checked ? f : "all")
                      }
                      className="mr-2"
                    />
                    <span className="capitalize">{f}</span>
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
            {logFilter !== "all" && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setLogFilter("all")}
              >
                Clear Filter
              </Button>
            )}
          </div>
          <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
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
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => {
                const log = row.original;
                return (
                  <div
                    key={row.id}
                    className="rounded-lg border bg-card p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <StatusCell isSuccess={log.is_success} />
                      <span className="text-xs text-muted-foreground">
                        {formatDistanceToNow(new Date(log.created_at), {
                          addSuffix: true,
                          locale: enUS,
                        })}
                      </span>
                    </div>
                    <div className="text-sm font-mono">
                      {log.ip_address.replace(/^::ffff:/, "")}
                    </div>
                    {log.location && (
                      <div className="flex items-center gap-1 text-sm text-muted-foreground">
                        <IconMapPin className="h-3 w-3" />
                        {log.location}
                      </div>
                    )}
                    {!log.is_success && log.failure_reason && (
                      <div className="text-sm text-red-600">
                        {log.failure_reason}
                      </div>
                    )}
                  </div>
                );
              })
            ) : (
              <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                No connection logs found.
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
                      No connection logs found.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}

        <DataTablePagination
          pageIndex={currentPage - 1}
          pageCount={Math.ceil(filteredData.length / currentLimit) || 1}
          pageSize={currentLimit}
          totalCount={filteredData.length}
          onPageChange={(pageIndex) => {
            setCurrentPage(pageIndex + 1);
            table.setPageIndex(pageIndex);
          }}
          onPageSizeChange={(size) => {
            setCurrentLimit(size);
            setCurrentPage(1);
            table.setPageSize(size);
            table.setPageIndex(0);
          }}
        />
      </CardContent>
    </Card>
  );
}
