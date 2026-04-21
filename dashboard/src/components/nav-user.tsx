"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { AVATAR_ENDPOINTS } from "@/lib/api/endpoints";
import { useAuth } from "@/lib/stores";
import {
  IconDotsVertical,
  IconLogout,
  IconUserCircle,
} from "@tabler/icons-react";
import { useState } from "react";
import * as React from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";

export function NavUser({
  user,
}: {
  user: {
    id: string;
    name: string;
    email: string;
    has_avatar?: boolean;
  };
}) {
  const { isMobile } = useSidebar();
  const navigate = useNavigate();
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const { logout } = useAuth();

  // Generate fresh timestamp for each render to bust avatar cache
  const timestamp = React.useMemo(() => Date.now(), []);
  const avatarUrl =
    user.id && user.has_avatar
      ? `${AVATAR_ENDPOINTS.GET(user.id)}?t=${timestamp}`
      : undefined;

  const handleDropdownOpenChange = (open: boolean) => {
    setIsDropdownOpen(open);
    if (!open) {
      setShowLogoutConfirm(false);
    }
  };

  const handleLogoutClick = () => {
    setShowLogoutConfirm(true);
  };

  const handleCancelClick = () => {
    setShowLogoutConfirm(false);
    setIsDropdownOpen(true);
  };

  const handleConfirmLogoutClick = async () => {
    document.cookie = "jwt=; Max-Age=0; path=/;";
    document.cookie = "refresh_token=; Max-Age=0; path=/;";

    setShowLogoutConfirm(false);
    setIsDropdownOpen(false);

    const success = await logout();
    if (success) {
      toast.success("Logged out successfully");
    } else {
      toast.error("Failed to logout");
    }

    navigate("/login");
  };

  return (
    <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu
            open={isDropdownOpen}
            onOpenChange={handleDropdownOpenChange}
          >
            <DropdownMenuTrigger asChild>
              <SidebarMenuButton
                size="lg"
                className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              >
                <Avatar key={avatarUrl} className="h-8 w-8 rounded-lg">
                  <AvatarImage src={avatarUrl} alt={user.name} />
                  <AvatarFallback className="rounded-lg">CR</AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">{user.name}</span>
                  <span className="text-muted-foreground truncate text-xs">
                    {user.email}
                  </span>
                </div>
                <IconDotsVertical className="ml-auto size-4" />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
              side={isMobile ? "bottom" : "right"}
              align="end"
              sideOffset={4}
            >
              <DropdownMenuLabel className="p-0 font-normal">
                <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                  <Avatar key={avatarUrl} className="h-8 w-8 rounded-lg">
                    <AvatarImage src={avatarUrl} alt={user.name} />
                    <AvatarFallback className="rounded-lg">CR</AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">{user.name}</span>
                    <span className="text-muted-foreground truncate text-xs">
                      {user.email}
                    </span>
                  </div>
                </div>
              </DropdownMenuLabel>

              <DropdownMenuSeparator />

              {!showLogoutConfirm ? (
                <>
                  <DropdownMenuGroup>
                    <DropdownMenuItem
                      onClick={() => navigate(`/users/${user.id}`)}
                    >
                      <IconUserCircle />
                      Account
                    </DropdownMenuItem>
                  </DropdownMenuGroup>

                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={(e) => {
                      e.preventDefault();
                      handleLogoutClick();
                    }}
                  >
                    <IconLogout />
                    Log out
                  </DropdownMenuItem>
                </>
              ) : (
                <>
                  <div className="px-2 py-1.5">
                    <div className="flex items-center gap-2 mb-2">
                      <IconLogout className="h-4 w-4 text-destructive" />
                      <span className="text-sm font-medium">
                        Confirm Logout
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground mb-3">
                      Are you sure you want to log out? You will need to sign in
                      again to access your account.
                    </p>
                  </div>
                  <DropdownMenuSeparator />
                  <div className="flex gap-1 p-1">
                    <DropdownMenuItem
                      onClick={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        handleCancelClick();
                      }}
                      className="flex-1 justify-center cursor-pointer"
                    >
                      Cancel
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={handleConfirmLogoutClick}
                      className="flex-1 justify-center cursor-pointer text-destructive focus:text-destructive"
                    >
                      Log out
                    </DropdownMenuItem>
                  </div>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
  );
}
