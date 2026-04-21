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
import { useUserSessions } from "@/hooks/useLogs";
import { useIsMobile } from "@/hooks/useMobile";
import type { Session } from "@/lib/types/user.types";
import {
  IconAlertTriangle,
  IconCalendar,
  IconCheck,
  IconClock,
  IconColumns,
  IconCopy,
  IconDeviceDesktop,
  IconDotsVertical,
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
import { differenceInSeconds, format, formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import * as React from "react";

interface UserActiveSessionsDataTableProps {
  userId: string;
}

function SessionIdCell({ id }: { id: string | undefined }) {
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

function StatusCell({ isActive }: { isActive: boolean | undefined }) {
  if (isActive === undefined) return <Badge variant="outline">Unknown</Badge>;
  return (
    <Badge variant={isActive ? "default" : "secondary"}>
      {isActive ? "Active" : "Inactive"}
    </Badge>
  );
}

function IpCell({ ip }: { ip: string | undefined }) {
  const [copied, setCopied] = React.useState(false);

  if (!ip) return <span className="text-muted-foreground">-</span>;

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(ip);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  return (
    <div className="flex items-center gap-1">
      <span className="text-sm font-mono">{ip}</span>
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

function DurationCell({ expiresAt }: { expiresAt: string }) {
  const expires = new Date(expiresAt);
  const now = new Date();

  if (now > expires)
    return <span className="text-muted-foreground text-sm">Expired</span>;

  const remainingSeconds = differenceInSeconds(expires, now);
  const remainingMinutes = Math.floor(remainingSeconds / 60);
  const remainingHours = Math.floor(remainingMinutes / 60);

  let durationText: string;
  if (remainingHours > 0) {
    durationText = `${remainingHours}h ${remainingMinutes % 60}m`;
  } else if (remainingMinutes > 0) {
    durationText = `${remainingMinutes}m ${remainingSeconds % 60}s`;
  } else {
    durationText = `${remainingSeconds}s`;
  }

  return (
    <div className="flex items-center gap-1 text-sm">
      <span>{durationText} remaining</span>
    </div>
  );
}

export function UserActiveSessionsDataTable({
  userId,
}: UserActiveSessionsDataTableProps) {
  const { sessions, loading, refreshing, error, refresh } =
    useUserSessions(userId);
  const isMobile = useIsMobile();

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      user_agent: false,
      device_fingerprint: false,
    });
  const [rowSelection, setRowSelection] = React.useState({});
  const [currentPage, setCurrentPage] = React.useState(1);
  const [currentLimit, setCurrentLimit] = React.useState(10);

  const columns: ColumnDef<Session>[] = React.useMemo(
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
        cell: ({ row }) => <SessionIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "device_fingerprint",
        header: "Device",
        cell: ({ row }) => (
          <div className="font-mono text-xs text-muted-foreground truncate max-w-[120px]">
            {row.getValue("device_fingerprint") || "-"}
          </div>
        ),
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => <StatusCell isActive={row.original.is_active} />,
      },
      {
        accessorKey: "ip_address",
        header: "IP Address",
        cell: ({ row }) => <IpCell ip={row.getValue("ip_address")} />,
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
        accessorKey: "created_at",
        header: "Created",
        cell: ({ row }) => {
          const dateValue = row.getValue("created_at") as string;
          if (
            !dateValue ||
            dateValue === "0001-01-01T00:00:00Z" ||
            Number.isNaN(new Date(dateValue).getTime())
          ) {
            return (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <span>-</span>
              </div>
            );
          }
          const date = new Date(dateValue);
          return (
            <div className="flex items-center gap-1 text-sm">
              <span title={format(date, "PPpp", { locale: enUS })}>
                {formatDistanceToNow(date, { addSuffix: true, locale: enUS })}
              </span>
            </div>
          );
        },
      },
      {
        accessorKey: "expires_at",
        header: "Expires",
        cell: ({ row }) => {
          const dateValue = row.getValue("expires_at") as string;
          if (
            !dateValue ||
            dateValue === "0001-01-01T00:00:00Z" ||
            Number.isNaN(new Date(dateValue).getTime())
          ) {
            return (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <IconClock className="h-4 w-4" />
                <span>-</span>
              </div>
            );
          }
          const date = new Date(dateValue);
          const isExpired = new Date() > date;
          return (
            <div className="flex items-center gap-1 text-sm">
              <IconClock className="h-4 w-4 text-muted-foreground" />
              <span
                title={format(date, "PPpp", { locale: enUS })}
                className={isExpired ? "text-red-600" : ""}
              >
                {isExpired
                  ? "Expired"
                  : formatDistanceToNow(date, {
                      addSuffix: true,
                      locale: enUS,
                    })}
              </span>
            </div>
          );
        },
      },
      {
        id: "duration",
        header: "Duration",
        cell: ({ row }) => <DurationCell expiresAt={row.original.expires_at} />,
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => {
          const session = row.original;
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
                  onClick={() => navigator.clipboard.writeText(session.id)}
                >
                  Copy Session ID
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() =>
                    navigator.clipboard.writeText(session.ip_address)
                  }
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
    data: sessions,
    columns,
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    pageCount: Math.ceil(sessions.length / currentLimit),
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

  if (loading && !refreshing && sessions.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle>User Active Sessions</CardTitle>
          <CardDescription>View active sessions for this user</CardDescription>
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
                  <Skeleton key={`session-skel-${i}`} className="h-10 w-full" />
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
          <CardTitle>User Active Sessions</CardTitle>
          <CardDescription>View active sessions for this user</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
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
        <CardTitle>User Active Sessions</CardTitle>
        <CardDescription>View active sessions for this user</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="relative">
          <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
            <div className="flex flex-wrap gap-2 w-full lg:w-auto lg:ml-auto">
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
                  const session = row.original;
                  const isExpired = new Date() > new Date(session.expires_at);
                  return (
                    <div
                      key={row.id}
                      className="rounded-lg border bg-card p-4 space-y-3"
                    >
                      <div className="flex items-center justify-between">
                        <StatusCell isActive={session.is_active} />
                        <span
                          className={`text-xs ${isExpired ? "text-red-600" : "text-muted-foreground"}`}
                        >
                          {isExpired
                            ? "Expired"
                            : formatDistanceToNow(
                                new Date(session.expires_at),
                                { addSuffix: true, locale: enUS },
                              )}
                        </span>
                      </div>
                      <div className="text-sm font-mono">
                        {session.ip_address}
                      </div>
                      <div className="flex items-center gap-1 text-sm text-muted-foreground">
                        Created{" "}
                        {formatDistanceToNow(new Date(session.created_at), {
                          addSuffix: true,
                          locale: enUS,
                        })}
                      </div>
                      <DurationCell expiresAt={session.expires_at} />
                    </div>
                  );
                })
              ) : (
                <div className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
                  No active sessions found.
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
                        No active sessions found.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          )}

          <DataTablePagination
            pageIndex={currentPage - 1}
            pageCount={table.getPageCount() || 1}
            pageSize={currentLimit}
            totalCount={sessions.length}
            onPageChange={(page) => table.setPageIndex(page)}
            onPageSizeChange={(newSize) => {
              setCurrentLimit(newSize);
              setCurrentPage(1);
            }}
          />
        </div>
      </CardContent>
    </Card>
  );
}
