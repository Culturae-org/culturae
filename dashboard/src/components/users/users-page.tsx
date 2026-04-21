"use client";

import { UserEditDialog } from "@/components/users/user-edit-dialog";
import { UserViewDialog } from "@/components/users/user-view-dialog";
import { UsersDataTable } from "@/components/users/users-data-table";
import { apiGet } from "@/lib/api-client";
import { USERS_ENDPOINTS } from "@/lib/api/endpoints";
import type { AdminUser } from "@/lib/types/user.types";
import * as React from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router";

export default function UsersPage() {
  const [totalUsers, setTotalUsers] = React.useState<number>(0);

  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const viewId = searchParams.get("view");

  const [viewUser, setViewUser] = React.useState<AdminUser | null>(null);
  const [editUser, setEditUser] = React.useState<AdminUser | null>(null);
  const [editOpen, setEditOpen] = React.useState(false);

  // Fetch user when viewId changes
  React.useEffect(() => {
    if (!viewId) {
      setViewUser(null);
      return;
    }
    apiGet(USERS_ENDPOINTS.GET(viewId))
      .then((r) => (r.ok ? r.json() : null))
      .then((json) => {
        if (json) setViewUser(json.data ?? json);
      })
      .catch(() => {});
  }, [viewId]);

  const handleViewClose = (open: boolean) => {
    if (!open) {
      setViewUser(null);
      navigate(location.pathname, { replace: true });
    }
  };

  const handleEditClick = () => {
    if (!viewUser) return;
    setEditUser(viewUser);
    handleViewClose(false);
    setEditOpen(true);
  };

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <h1 className="text-3xl font-bold">Players Management</h1>
        <p className="text-muted-foreground">
          Manage players on your platform • {totalUsers} total users
        </p>
      </div>

      <div className="px-4 lg:px-6">
        <UsersDataTable onTotalCountChange={setTotalUsers} />
      </div>

      {viewUser && (
        <UserViewDialog
          user={viewUser}
          open={true}
          onOpenChange={handleViewClose}
          onEditClick={handleEditClick}
        />
      )}

      {editUser && (
        <UserEditDialog
          user={editUser}
          open={editOpen}
          onOpenChange={(open) => {
            setEditOpen(open);
            if (!open) setEditUser(null);
          }}
          onUserUpdated={() => {}}
        />
      )}
    </div>
  );
}
