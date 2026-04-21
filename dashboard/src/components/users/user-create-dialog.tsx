"use client";

import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useIsMobile } from "@/hooks/useMobile";
import { useUsers } from "@/hooks/useUsers";
import { IconPlus, IconUserPlus } from "@tabler/icons-react";
import * as React from "react";

type UserCreateData = {
  email: string;
  username: string;
  role: string;
  account_status: string;
  password: string;
  confirmPassword: string;
};

interface UserCreateDialogProps {
  onUserCreated?: () => void;
}

export function UserCreateDialog({ onUserCreated }: UserCreateDialogProps) {
  const isMobile = useIsMobile();
  const { createUser } = useUsers();
  const [isOpen, setIsOpen] = React.useState(false);
  const [isLoading, setIsLoading] = React.useState(false);
  const [formData, setFormData] = React.useState<UserCreateData>({
    email: "",
    username: "",
    role: "user",
    account_status: "active",
    password: "",
    confirmPassword: "",
  });
  const [passwordErrors, setPasswordErrors] = React.useState<string[]>([]);

  const handleInputChange = (field: keyof UserCreateData, value: string) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
    if (field === "password" || field === "confirmPassword") {
      if (passwordErrors.length > 0) {
        setPasswordErrors([]);
      }
    }
  };

  const validatePassword = (): boolean => {
    const errors: string[] = [];

    if (!formData.password) {
      errors.push("Password is required");
    } else if (formData.password.length < 8) {
      errors.push("Password must be at least 8 characters long");
    }

    if (!formData.confirmPassword) {
      errors.push("Please confirm your password");
    } else if (formData.password !== formData.confirmPassword) {
      errors.push("Passwords do not match");
    }

    setPasswordErrors(errors);
    return errors.length === 0;
  };

  const validateForm = (): boolean => {
    if (!formData.username.trim()) {
      return false;
    }
    if (!formData.email.trim()) {
      return false;
    }
    if (!formData.role) {
      return false;
    }
    if (!formData.account_status) {
      return false;
    }
    if (!validatePassword()) {
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsLoading(true);

    try {
      await createUser({
        email: formData.email,
        username: formData.username,
        role: formData.role,
        account_status: formData.account_status,
        password: formData.password,
      });

      setFormData({
        email: "",
        username: "",
        role: "user",
        account_status: "active",
        password: "",
        confirmPassword: "",
      });
      setPasswordErrors([]);

      setIsOpen(false);

      if (onUserCreated) {
        onUserCreated();
      }
    } catch (error) {
      console.error("Error creating user:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open);
    if (!open) {
      setFormData({
        email: "",
        username: "",
        role: "user",
        account_status: "active",
        password: "",
        confirmPassword: "",
      });
      setPasswordErrors([]);
    }
  };

  return (
    <Sheet modal={false} open={isOpen} onOpenChange={handleOpenChange}>
      <SheetTrigger asChild>
        <Button size="sm">
          <IconPlus className="h-4 w-4" />
          {!isMobile && " Create User"}
        </Button>
      </SheetTrigger>
      <SheetContent
        side={isMobile ? "bottom" : "right"}
        className="flex flex-col p-0 sm:max-w-sm overflow-hidden"
        onOpenAutoFocus={(e) => e.preventDefault()}
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        <SheetHeader className="gap-1 px-4 pt-4">
          <SheetTitle className="flex items-center gap-2">
            <IconUserPlus className="h-5 w-5" />
            Create New User
          </SheetTitle>
          <SheetDescription>Add a new user to the app</SheetDescription>
        </SheetHeader>

        <div className="flex flex-col gap-4 overflow-y-auto flex-1 px-4 text-sm">
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div className="flex flex-col gap-3">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                value={formData.username}
                onChange={(e) => handleInputChange("username", e.target.value)}
                placeholder="Enter username"
                required
              />
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={formData.email}
                onChange={(e) => handleInputChange("email", e.target.value)}
                placeholder="Enter email address"
                required
              />
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="role">Role</Label>
              <Select
                value={formData.role}
                onValueChange={(value) => handleInputChange("role", value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select role" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="administrator">Administrator</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="account_status">Account Status</Label>
              <Select
                value={formData.account_status}
                onValueChange={(value) =>
                  handleInputChange("account_status", value)
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select account status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="suspended">Suspended</SelectItem>
                  <SelectItem value="banned">Banned</SelectItem>
                  <SelectItem value="inactive">Inactive</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="pt-4 border-t">
              <div className="flex flex-col gap-3">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) =>
                    handleInputChange("password", e.target.value)
                  }
                  placeholder="Enter password"
                  required
                  minLength={8}
                />
              </div>

              <div className="flex flex-col gap-3 mt-4">
                <Label htmlFor="confirmPassword">Confirm Password</Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  value={formData.confirmPassword}
                  onChange={(e) =>
                    handleInputChange("confirmPassword", e.target.value)
                  }
                  placeholder="Confirm password"
                  required
                  minLength={8}
                />
              </div>

              {passwordErrors.length > 0 && (
                <div className="mt-2 p-3 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 rounded-lg">
                  <ul className="text-sm text-red-600 dark:text-red-400 space-y-1">
                    {passwordErrors.map((error) => (
                      <li key={error}>• {error}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </form>
        </div>

        <SheetFooter className="px-4 pb-4">
          <div className="flex gap-2">
            <Button
              onClick={handleSubmit}
              disabled={isLoading}
              className="flex-1"
            >
              {isLoading ? "Creating..." : "Create User"}
            </Button>
          </div>
          <SheetClose asChild>
            <Button variant="outline" disabled={isLoading}>
              Cancel
            </Button>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
