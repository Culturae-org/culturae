"use client";

import { AvatarPreviewDialog } from "@/components/avatar/avatar-preview-dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
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
import { useUsersList } from "@/hooks/useUsers";
import { useXPSettings } from "@/hooks/useXPSettings";
import { AVATAR_ENDPOINTS } from "@/lib/api/endpoints";
import type { AdminUser } from "@/lib/types/user.types";
import {
  IconColumns,
  IconDotsVertical,
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
import { useNavigate } from "react-router";
import { UserBanDialog } from "./user-ban-dialog";
import { UserCreateDialog } from "./user-create-dialog";
import { UserDeleteDialog } from "./user-delete-dialog";
import { UserEditDialog } from "./user-edit-dialog";

interface UsersDataTableProps {
  onTotalCountChange?: (count: number) => void;
}

function UserIdCell({ id }: { id: string }) {
  const [isExpanded, setIsExpanded] = React.useState(false);

  return (
    <button
      type="button"
      className="font-mono text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors bg-transparent border-none p-0 text-left"
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

interface UserActionsCellProps {
  user: AdminUser;
  onStatusChange: (userId: string, status: string) => void;
  onUnban: (userId: string) => Promise<AdminUser>;
  onBanClick: (user: { id: string; username: string }) => void;
  onUserUpdated: () => void;
  onUserDeleted: (userId: string) => void;
  setData: React.Dispatch<React.SetStateAction<AdminUser[]>>;
}

function UserActionsCell({
  user,
  onStatusChange,
  onUnban,
  onBanClick,
  onUserUpdated,
  onUserDeleted,
  setData,
}: UserActionsCellProps) {
  const navigate = useNavigate();
  const [editOpen, setEditOpen] = React.useState(false);
  const [deleteOpen, setDeleteOpen] = React.useState(false);

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
          <DropdownMenuItem onSelect={() => navigate(`?view=${user.id}`)}>
            View User
          </DropdownMenuItem>
          <DropdownMenuItem onSelect={() => setEditOpen(true)}>
            Edit User
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => onStatusChange(user.id, "active")}
            disabled={user.account_status === "active"}
          >
            Activate Account
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => onStatusChange(user.id, "suspended")}
            disabled={user.account_status === "suspended"}
          >
            Suspend Account
          </DropdownMenuItem>
          {user.account_status === "banned" ? (
            <DropdownMenuItem
              onClick={() => {
                onUnban(user.id).then((updated) => {
                  setData((prev) =>
                    prev.map((u) => (u.id === user.id ? updated : u)),
                  );
                });
              }}
            >
              Unban Account
            </DropdownMenuItem>
          ) : (
            <DropdownMenuItem
              onClick={() =>
                onBanClick({ id: user.id, username: user.username })
              }
            >
              Ban Account
            </DropdownMenuItem>
          )}
          <DropdownMenuItem
            onClick={() => onStatusChange(user.id, "inactive")}
            disabled={user.account_status === "inactive"}
          >
            Deactivate Account
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            className="text-destructive"
            onSelect={() => setDeleteOpen(true)}
          >
            Delete User
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      {!deleteOpen && (
        <UserEditDialog
          user={user}
          onUserUpdated={onUserUpdated}
          open={editOpen}
          onOpenChange={setEditOpen}
        />
      )}
      {!editOpen && (
        <UserDeleteDialog
          user={user}
          onUserDeleted={onUserDeleted}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
        />
      )}
    </>
  );
}

export function UsersDataTable({
  onTotalCountChange,
}: UsersDataTableProps = {}) {
  const {
    users,
    setUsers,
    loading,
    refreshing,
    error,
    currentPage,
    totalPages,
    totalCount,
    currentLimit,
    filters,
    search,
    setSearchQuery: setSearch,
    setFilter,
    setPage: goToPage,
    setLimit: setPageSize,
    refetch: refresh,
    updateUserStatus,
    deactivateUser,
    banUser,
    unbanUser,
  } = useUsersList({ onTotalCountChange });

  const { ranks } = useXPSettings();

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    [],
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      language: false,
      has_avatar: false,
      avatar: true,
      public_id: false,
      updated_at: false,
      score: false,
    });
  const [removingIds, setRemovingIds] = React.useState<Set<string>>(new Set());
  const [previewUser, setPreviewUser] = React.useState<{
    username: string;
    avatarUrl: string;
  } | null>(null);
  const [previewOpen, setPreviewOpen] = React.useState(false);
  const [banDialogOpen, setBanDialogOpen] = React.useState(false);
  const [banTargetUser, setBanTargetUser] = React.useState<{
    id: string;
    username: string;
  } | null>(null);
  const [banLoading, setBanLoading] = React.useState(false);
  const navigate = useNavigate();
  const [editUser, setEditUser] = React.useState<AdminUser | null>(null);
  const [editOpen, setEditOpen] = React.useState(false);

  const animateRemove = React.useCallback(
    (ids: string[], onDone: () => void) => {
      setRemovingIds(new Set(ids));
      setTimeout(() => {
        onDone();
        setRemovingIds(new Set());
      }, 300);
    },
    [],
  );

  const recentPresenceRef = React.useRef<
    Record<string, { is_online: boolean; ts: number }>
  >({});

  React.useEffect(() => {
    const presenceTimeouts: Record<string, number | undefined> = {};

    const handler = (e: Event) => {
      try {
        const detail = (e as CustomEvent).detail;
        if (!detail?.user_id) return;
        const { user_id, is_online } = detail;

        recentPresenceRef.current[user_id] = { is_online, ts: Date.now() };
        setTimeout(() => {
          const rec = recentPresenceRef.current[user_id];
          if (rec && Date.now() - rec.ts > 15000)
            delete recentPresenceRef.current[user_id];
        }, 16000);

        if (is_online) {
          const timeoutId = presenceTimeouts[user_id];
          if (timeoutId !== undefined) {
            window.clearTimeout(timeoutId);
            presenceTimeouts[user_id] = undefined;
          }
          setUsers((prev) => {
            let changed = false;
            const next = prev.map((u) => {
              if (u.id === user_id && u.is_online !== true) {
                changed = true;
                return { ...u, is_online: true };
              }
              return u;
            });
            return changed ? next : prev;
          });
          return;
        }

        if (presenceTimeouts[user_id] !== undefined)
          window.clearTimeout(presenceTimeouts[user_id]);
        presenceTimeouts[user_id] = window.setTimeout(() => {
          setUsers((prev) => {
            let changed = false;
            const next = prev.map((u) => {
              if (u.id === user_id && u.is_online !== false) {
                changed = true;
                return { ...u, is_online: false };
              }
              return u;
            });
            return changed ? next : prev;
          });
          presenceTimeouts[user_id] = undefined;
        }, 5000);
      } catch (err) {
        console.error("users-data-table: presence handler error", err);
      }
    };

    window.addEventListener("user-presence-change", handler as EventListener);
    return () => {
      window.removeEventListener(
        "user-presence-change",
        handler as EventListener,
      );
      for (const t of Object.values(presenceTimeouts)) {
        if (t) window.clearTimeout(t);
      }
    };
  }, [setUsers]);

  const handleStatusChange = React.useCallback(
    async (userId: string, newStatus: string) => {
      try {
        if (newStatus === "inactive") {
          await deactivateUser(userId);
        } else {
          await updateUserStatus(userId, { account_status: newStatus });
        }
      } catch {
        // errors handled inside hook with toast
      }
    },
    [updateUserStatus, deactivateUser],
  );

  const columns: ColumnDef<AdminUser>[] = React.useMemo(
    () => [
      {
        accessorKey: "id",
        header: "ID",
        cell: ({ row }) => <UserIdCell id={row.getValue("id")} />,
      },
      {
        accessorKey: "username",
        header: "Username",
        cell: ({ row }) => (
          <div className="font-medium">{row.getValue("username")}</div>
        ),
      },
      {
        accessorKey: "email",
        header: "Email",
        cell: ({ row }) => (
          <div className="text-muted-foreground">{row.getValue("email")}</div>
        ),
      },
      {
        accessorKey: "public_id",
        header: "Public ID",
        cell: ({ row }) => <UserIdCell id={row.getValue("public_id")} />,
      },
      {
        accessorKey: "role",
        header: "Role",
        cell: ({ row }) => {
          const role = row.getValue("role") as string;
          return (
            <Badge
              variant={
                role === "administrator" || role === "moderator"
                  ? "default"
                  : "secondary"
              }
              className={
                role === "moderator" ? "bg-purple-600 hover:bg-purple-700" : ""
              }
            >
              {role.charAt(0).toUpperCase() + role.slice(1)}
            </Badge>
          );
        },
      },
      {
        accessorKey: "rank",
        header: "Rank",
        cell: ({ row }) => {
          const rank = (row.getValue("rank") as string) || "";
          const rankColors: Record<string, string> = {
            beginner:
              "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
            intermediate:
              "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
            pro: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300",
            expert:
              "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300",
            legend:
              "bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-300",
          };
          return (
            <Badge
              variant="outline"
              className={rankColors[rank.toLowerCase()] || ""}
            >
              {rank || "N/A"}
            </Badge>
          );
        },
      },
      {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => {
          const isOnline = row.original.is_online;
          return (
            <div className="flex items-center gap-2">
              <div
                className={`h-2 w-2 rounded-full ${isOnline ? "bg-green-500" : "bg-gray-400"}`}
              />
              <span className="text-sm">{isOnline ? "Online" : "Offline"}</span>
            </div>
          );
        },
      },
      {
        accessorKey: "account_status",
        header: "Account Status",
        cell: ({ row }) => {
          const accountStatus =
            (row.getValue("account_status") as string) || "active";
          const bannedUntil = row.original.banned_until;
          const banReason = row.original.ban_reason;
          const statusColors = {
            active:
              "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
            suspended:
              "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
            banned: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
            inactive:
              "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300",
          };
          const banInfo =
            accountStatus === "banned" && bannedUntil
              ? `Until ${new Date(bannedUntil).toLocaleString()}`
              : accountStatus === "banned" && !bannedUntil
                ? "Permanent"
                : null;

          return (
            <div className="flex flex-col gap-0.5">
              <Badge
                variant="outline"
                className={
                  statusColors[accountStatus as keyof typeof statusColors] || ""
                }
              >
                {accountStatus.charAt(0).toUpperCase() + accountStatus.slice(1)}
              </Badge>
              {banInfo && (
                <span
                  className="text-[10px] text-muted-foreground"
                  title={banReason || undefined}
                >
                  {banInfo}
                </span>
              )}
            </div>
          );
        },
      },
      {
        accessorKey: "score",
        header: "Score",
        cell: ({ row }) => {
          const score = row.original.game_stats?.total_score ?? 0;
          return (
            <div className="text-sm font-medium">{score.toLocaleString()}</div>
          );
        },
      },
      {
        accessorKey: "created_at",
        header: "Created Date",
        cell: ({ row }) => (
          <div className="text-sm">
            {new Date(row.getValue("created_at")).toLocaleDateString("en-US", {
              day: "2-digit",
              month: "2-digit",
              year: "numeric",
            })}
          </div>
        ),
      },
      {
        accessorKey: "language",
        header: "Language",
        cell: ({ row }) => (
          <div className="text-sm capitalize">{row.getValue("language")}</div>
        ),
      },
      {
        accessorKey: "has_avatar",
        header: "Has Avatar",
        cell: ({ row }) => {
          const hasAvatar = row.getValue("has_avatar") as boolean;
          return (
            <Badge
              variant={hasAvatar ? "default" : "secondary"}
              className="text-xs"
            >
              {hasAvatar ? "✓" : "✗"}
            </Badge>
          );
        },
      },
      {
        id: "avatar",
        header: "Avatar",
        cell: ({ row }) => {
          const hasAvatar = row.getValue("has_avatar") as boolean;
          const userId = row.getValue("id") as string;
          const username = row.getValue("username") as string;
          if (!hasAvatar) return null;
          return (
            <button
              type="button"
              className="cursor-pointer hover:opacity-80 transition-opacity bg-transparent border-none p-0"
              onClick={(e) => {
                e.stopPropagation();
                setPreviewUser({
                  username,
                  avatarUrl: AVATAR_ENDPOINTS.GET(userId),
                });
                setPreviewOpen(true);
              }}
            >
              <Avatar className="h-8 w-8">
                <AvatarImage
                  src={AVATAR_ENDPOINTS.GET(userId)}
                  alt={`${username}'s avatar`}
                />
                <AvatarFallback className="text-xs">
                  {username.charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
            </button>
          );
        },
      },
      {
        accessorKey: "updated_at",
        header: "Last Updated",
        cell: ({ row }) => (
          <div className="text-sm">
            {new Date(row.getValue("updated_at")).toLocaleDateString("en-US", {
              day: "2-digit",
              month: "2-digit",
              year: "numeric",
            })}
          </div>
        ),
      },
      {
        id: "actions",
        header: "Actions",
        enableHiding: false,
        cell: ({ row }) => {
          const user = row.original;
          return (
            <UserActionsCell
              user={user}
              onStatusChange={handleStatusChange}
              onUnban={unbanUser}
              onBanClick={(u) => {
                setBanTargetUser(u);
                setBanDialogOpen(true);
              }}
              onUserUpdated={refresh}
              onUserDeleted={(userId) => {
                animateRemove([userId], () => {
                  setUsers((prev) => prev.filter((u) => u.id !== userId));
                });
              }}
              setData={setUsers}
            />
          );
        },
      },
    ],
    [refresh, handleStatusChange, animateRemove, unbanUser, setUsers],
  );

  const table = useReactTable({
    data: users,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    globalFilterFn: () => true,
    manualPagination: true,
    pageCount: totalPages,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      pagination: { pageIndex: currentPage - 1, pageSize: currentLimit },
    },
    onPaginationChange: (updater) => {
      const newPagination =
        typeof updater === "function"
          ? updater(table.getState().pagination)
          : updater;
      const newPage = newPagination.pageIndex + 1;
      const newLimit = newPagination.pageSize;
      if (newPage !== currentPage) goToPage(newPage);
      else if (newLimit !== currentLimit) setPageSize(newLimit);
    },
  });

  const _isMobile = useIsMobile();
  const hasActiveFilters =
    filters.account_status !== "" ||
    filters.role !== "" ||
    filters.rank !== "" ||
    filters.status !== "";

  if (loading && !refreshing && users.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <Skeleton className="h-6 w-32" />
              <Skeleton className="h-4 w-48 mt-1" />
            </div>
            <Skeleton className="h-9 w-32" />
          </div>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <div className="space-y-3 p-4">
              {[0, 1, 2, 3, 4, 5, 6, 7].map((i) => (
                <Skeleton key={`user-skel-${i}`} className="h-10 w-full" />
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
        <CardContent className="pt-6">
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
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <CardTitle>Users</CardTitle>
            <CardDescription>Manage and view user accounts</CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-start lg:items-center py-4 gap-4 mb-4">
          <div className="relative flex-1 max-w-sm w-full lg:w-auto">
            <IconSearch className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search users by username, email..."
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
              <DropdownMenuLabel asChild>
                <UserCreateDialog onUserCreated={refresh} />
              </DropdownMenuLabel>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <IconFilter className="h-4 w-4" />
                  Filter
                  {hasActiveFilters && (
                    <Badge
                      variant="secondary"
                      className="ml-1 h-5 px-1 text-xs"
                    >
                      {
                        [
                          filters.role,
                          filters.rank,
                          filters.account_status,
                          filters.status,
                        ].filter(Boolean).length
                      }
                    </Badge>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Role Filter
                </div>
                {["administrator", "user"].map((r) => (
                  <DropdownMenuItem
                    key={r}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={filters.role === r}
                      onCheckedChange={(checked) =>
                        setFilter("role", checked ? r : "")
                      }
                      className="mr-2"
                    />
                    {r.charAt(0).toUpperCase() + r.slice(1)}
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Rank Filter
                </div>
                {ranks.map((r) => (
                  <DropdownMenuItem
                    key={r.name}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={filters.rank === r.name}
                      onCheckedChange={(checked) =>
                        setFilter("rank", checked ? r.name : "")
                      }
                      className="mr-2"
                    />
                    {r.name}
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Account Status Filter
                </div>
                {["active", "suspended", "banned", "inactive"].map((s) => (
                  <DropdownMenuItem
                    key={s}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={filters.account_status === s}
                      onCheckedChange={(checked) =>
                        setFilter("account_status", checked ? s : "")
                      }
                      className="mr-2"
                    />
                    {s.charAt(0).toUpperCase() + s.slice(1)}
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm font-semibold">
                  Status Filter
                </div>
                {["online", "offline"].map((s) => (
                  <DropdownMenuItem
                    key={s}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <Checkbox
                      checked={filters.status === s}
                      onCheckedChange={(checked) =>
                        setFilter("status", checked ? s : "")
                      }
                      className="mr-2"
                    />
                    {s.charAt(0).toUpperCase() + s.slice(1)}
                  </DropdownMenuItem>
                ))}
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
                      onCheckedChange={(value) => col.toggleVisibility(!!value)}
                    >
                      {col.id}
                    </DropdownMenuCheckboxItem>
                  ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

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
                    className={`transition-all duration-300 cursor-pointer ${removingIds.has(row.original.id) ? "opacity-0 scale-y-0 h-0" : ""}`}
                    onClick={(e) => {
                      const target = e.target as HTMLElement;
                      if (
                        target.closest(
                          'button, [role="checkbox"], [role="menuitem"], [data-radix-collection-item]',
                        )
                      )
                        return;
                      if (
                        document.querySelector(
                          '[data-vaul-overlay], [data-radix-dialog-overlay], [role="dialog"]',
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
                    className="h-24 text-center"
                  >
                    No users found.
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
          loading={loading}
        />

        <AvatarPreviewDialog
          open={previewOpen}
          onOpenChange={setPreviewOpen}
          user={previewUser}
        />
        <UserBanDialog
          open={banDialogOpen}
          onOpenChange={setBanDialogOpen}
          username={banTargetUser?.username ?? ""}
          loading={banLoading}
          onConfirm={async (duration, reason) => {
            if (!banTargetUser) return;
            setBanLoading(true);
            try {
              await banUser(banTargetUser.id, { duration, reason });
              setBanDialogOpen(false);
            } finally {
              setBanLoading(false);
            }
          }}
        />
        {editUser && (
          <UserEditDialog
            user={editUser}
            open={editOpen}
            onOpenChange={(open) => {
              setEditOpen(open);
              if (!open) setEditUser(null);
            }}
            onUserUpdated={refresh}
          />
        )}
      </CardContent>
    </Card>
  );
}
