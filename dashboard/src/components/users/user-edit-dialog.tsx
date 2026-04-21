"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { useIsMobile } from "@/hooks/useMobile";
import { useUsers } from "@/hooks/useUsers";
import { apiPost } from "@/lib/api-client";
import { AVATAR_ENDPOINTS, USERS_ENDPOINTS } from "@/lib/api/endpoints";
import type { AdminUser } from "@/lib/types/user.types";
import {
  IconAlertTriangle,
  IconEdit,
  IconRefresh,
  IconTrash,
  IconUser,
} from "@tabler/icons-react";
import * as React from "react";
import { toast } from "sonner";

type UserEditProps = AdminUser;

type UserFormData = {
  email: string;
  username: string;
  role: string;
  account_status: string;
  language: string;
  bio: string;
  is_profile_public: boolean;
  show_online_status: boolean;
  allow_friend_requests: boolean;
  allow_party_invites: boolean;
};

type PasswordUpdateData = {
  password: string;
  confirmPassword: string;
};

interface UserEditDialogProps {
  user: UserEditProps;
  onUserUpdated?: () => void;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function UserEditDialog({
  user,
  onUserUpdated,
  open: controlledOpen,
  onOpenChange,
}: UserEditDialogProps) {
  const isMobile = useIsMobile();
  const { updateUser, updateUserPassword, uploadAvatar, deleteAvatar } =
    useUsers();
  const [internalOpen, setInternalOpen] = React.useState(false);
  const isOpen = controlledOpen ?? internalOpen;
  const setIsOpen = onOpenChange ?? setInternalOpen;
  const [isLoading, setIsLoading] = React.useState(false);
  const [activeTab, setActiveTab] = React.useState("general");
  const [formData, setFormData] = React.useState<UserFormData>({
    email: user.email,
    username: user.username,
    role: user.role,
    account_status: user.account_status,
    language: user.language ?? "en",
    bio: user.bio ?? "",
    is_profile_public: user.is_profile_public ?? true,
    show_online_status: user.show_online_status ?? true,
    allow_friend_requests: user.allow_friend_requests ?? true,
    allow_party_invites: user.allow_party_invites ?? true,
  });
  const [passwordData, setPasswordData] = React.useState<PasswordUpdateData>({
    password: "",
    confirmPassword: "",
  });
  const [passwordErrors, setPasswordErrors] = React.useState<string[]>([]);
  const [showRegenerateDialog, setShowRegenerateDialog] = React.useState(false);
  const [isRegenerating, setIsRegenerating] = React.useState(false);

  const avatarUrl = user.has_avatar ? AVATAR_ENDPOINTS.GET(user.id) : undefined;

  React.useEffect(() => {
    setFormData({
      email: user.email,
      username: user.username,
      role: user.role,
      account_status: user.account_status,
      language: user.language ?? "en",
      bio: user.bio ?? "",
      is_profile_public: user.is_profile_public ?? true,
      show_online_status: user.show_online_status ?? true,
      allow_friend_requests: user.allow_friend_requests ?? true,
      allow_party_invites: user.allow_party_invites ?? true,
    });
  }, [user]);

  const handleInputChange = (
    field: keyof UserFormData,
    value: string | boolean,
  ) => {
    if (field === "bio" && typeof value === "string" && value.length > 50) {
      return;
    }
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handlePasswordChange = (
    field: keyof PasswordUpdateData,
    value: string,
  ) => {
    setPasswordData((prev) => ({
      ...prev,
      [field]: value,
    }));
    if (passwordErrors.length > 0) {
      setPasswordErrors([]);
    }
  };

  const validatePassword = (): boolean => {
    const errors: string[] = [];

    if (!passwordData.password) {
      errors.push("Password is required");
    } else if (passwordData.password.length < 8) {
      errors.push("Password must be at least 8 characters long");
    }

    if (!passwordData.confirmPassword) {
      errors.push("Please confirm your password");
    } else if (passwordData.password !== passwordData.confirmPassword) {
      errors.push("Passwords do not match");
    }

    setPasswordErrors(errors);
    return errors.length === 0;
  };

  const handleSubmitUser = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      if (!formData.username.trim()) {
        toast.error("Username is required");
        return;
      }
      if (!formData.email.trim()) {
        toast.error("Email is required");
        return;
      }
      if (!formData.role) {
        toast.error("Role is required");
        return;
      }
      if (!formData.account_status) {
        toast.error("Account status is required");
        return;
      }
      if (!formData.language) {
        toast.error("Language is required");
        return;
      }
      if (formData.bio && formData.bio.length > 50) {
        toast.error("Bio must not exceed 50 characters");
        return;
      }

      await updateUser(user.id, formData);

      setIsOpen(false);

      if (onUserUpdated) {
        onUserUpdated();
      }
    } catch (error) {
      console.error("Error updating user:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmitPassword = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validatePassword()) {
      return;
    }

    setIsLoading(true);

    try {
      await updateUserPassword(user.id, { password: passwordData.password });

      setPasswordData({ password: "", confirmPassword: "" });
      setPasswordErrors([]);

      setActiveTab("general");
    } catch (error) {
      console.error("Error updating password:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      toast.error("File too large", {
        description: "Avatar must be less than 5MB",
      });
      e.target.value = "";
      return;
    }

    if (!file.type.startsWith("image/")) {
      toast.error("Invalid file type", {
        description: "Please upload an image file (JPG, PNG, GIF)",
      });
      e.target.value = "";
      return;
    }

    setIsLoading(true);

    try {
      await uploadAvatar(user.id, file);

      if (onUserUpdated) {
        onUserUpdated();
      }

      if (e.target) {
        e.target.value = "";
      }
    } catch (error) {
      console.error("Error uploading avatar:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleAvatarDelete = async () => {
    if (!user.has_avatar) return;

    setIsLoading(true);

    try {
      await deleteAvatar(user.id);

      if (onUserUpdated) {
        onUserUpdated();
      }
    } catch (error) {
      console.error("Error deleting avatar:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleRegeneratePublicId = async () => {
    setIsRegenerating(true);

    try {
      const response = await apiPost(
        USERS_ENDPOINTS.REGENERATE_PUBLIC_ID(user.id),
        {},
      );

      if (!response.ok) {
        const errorData = await response.json();
        toast.error("Failed to regenerate public ID", {
          description: errorData.error || "An error occurred",
        });
        throw new Error("Failed to regenerate public ID");
      }

      const data = await response.json();
      toast.success("Public ID regenerated successfully", {
        description: `New public ID: ${data.public_id}`,
      });

      setShowRegenerateDialog(false);

      if (onUserUpdated) {
        onUserUpdated();
      }
    } catch (error) {
      console.error("Error regenerating public ID:", error);
    } finally {
      setIsRegenerating(false);
    }
  };

  return (
    <>
      <Sheet open={isOpen} onOpenChange={setIsOpen}>
        {controlledOpen === undefined && (
          <SheetTrigger asChild>
            <Button variant="ghost" size="sm" onClick={() => setIsOpen(true)}>
              <IconEdit className="h-4 w-4" />
            </Button>
          </SheetTrigger>
        )}
        <SheetContent
          side={isMobile ? "bottom" : "right"}
          className="flex flex-col p-0 sm:max-w-sm overflow-hidden"
        >
          <SheetHeader className="gap-1 px-4 pt-4">
            <SheetTitle>Edit User</SheetTitle>
            <SheetDescription>
              Modify user information for <strong>{user.username}</strong>
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-col gap-4 overflow-y-auto flex-1 px-4 text-sm">
            <Tabs
              value={activeTab}
              onValueChange={setActiveTab}
              className="w-full"
            >
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="general">General Info</TabsTrigger>
                <TabsTrigger value="password">Change Password</TabsTrigger>
              </TabsList>

              <TabsContent value="general" className="space-y-4 mt-4">
                <form
                  onSubmit={handleSubmitUser}
                  className="flex flex-col gap-4"
                >
                  <div className="flex flex-col gap-3">
                    <Label htmlFor="username">Username</Label>
                    <Input
                      id="username"
                      value={formData.username}
                      onChange={(e) =>
                        handleInputChange("username", e.target.value)
                      }
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
                      onChange={(e) =>
                        handleInputChange("email", e.target.value)
                      }
                      placeholder="Enter email"
                      required
                    />
                  </div>

                  <div className="flex flex-col gap-3">
                    <Label htmlFor="role">Role</Label>
                    <Select
                      value={formData.role}
                      onValueChange={(value) =>
                        handleInputChange("role", value)
                      }
                    >
                      <SelectTrigger id="role">
                        <SelectValue placeholder="Select role" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="user">User</SelectItem>
                        <SelectItem value="administrator">
                          Administrator
                        </SelectItem>
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
                      <SelectTrigger id="account_status">
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

                  <div className="flex flex-col gap-3">
                    <Label htmlFor="language">Language</Label>
                    <Select
                      value={formData.language}
                      onValueChange={(value) =>
                        handleInputChange("language", value)
                      }
                    >
                      <SelectTrigger id="language">
                        <SelectValue placeholder="Select language" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="en">English</SelectItem>
                        <SelectItem value="fr">Français</SelectItem>
                        <SelectItem value="es">Español</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="flex flex-col gap-3">
                    <Label htmlFor="bio">Bio</Label>
                    <Textarea
                      id="bio"
                      value={formData.bio}
                      onChange={(e) => handleInputChange("bio", e.target.value)}
                      placeholder="Enter user bio (optional)"
                      rows={3}
                      className="resize-none"
                      maxLength={50}
                    />
                    <div className="flex justify-between items-center">
                      <p className="text-xs text-muted-foreground">
                        Description visible on the user&apos;s public profile
                      </p>
                      <span
                        className={`text-xs ${formData.bio.length > 45 ? "text-orange-500" : "text-muted-foreground"}`}
                      >
                        {formData.bio.length}/50
                      </span>
                    </div>
                  </div>

                  <div className="pt-4 border-t">
                    <div className="flex flex-col gap-4">
                      <div className="flex items-center">
                        <Label className="text-base font-semibold">
                          Privacy Settings
                        </Label>
                      </div>

                      <div className="space-y-4">
                        <div className="flex items-center justify-between p-3 border rounded-lg">
                          <div className="flex-1">
                            <Label
                              htmlFor="is_profile_public"
                              className="text-sm font-medium cursor-pointer"
                            >
                              Profile Public
                            </Label>
                            <p className="text-xs text-muted-foreground">
                              Allow others to view this user&apos;s profile
                            </p>
                          </div>
                          <Switch
                            id="is_profile_public"
                            checked={formData.is_profile_public ?? true}
                            onCheckedChange={(checked: boolean) =>
                              handleInputChange(
                                "is_profile_public" as keyof UserFormData,
                                checked,
                              )
                            }
                          />
                        </div>

                        <div className="flex items-center justify-between p-3 border rounded-lg">
                          <div className="flex-1">
                            <Label
                              htmlFor="show_online_status"
                              className="text-sm font-medium cursor-pointer"
                            >
                              Show Online Status
                            </Label>
                            <p className="text-xs text-muted-foreground">
                              Display when this user is online
                            </p>
                          </div>
                          <Switch
                            id="show_online_status"
                            checked={formData.show_online_status ?? true}
                            onCheckedChange={(checked: boolean) =>
                              handleInputChange(
                                "show_online_status" as keyof UserFormData,
                                checked,
                              )
                            }
                          />
                        </div>

                        <div className="flex items-center justify-between p-3 border rounded-lg">
                          <div className="flex-1">
                            <Label
                              htmlFor="allow_friend_requests"
                              className="text-sm font-medium cursor-pointer"
                            >
                              Allow Friend Requests
                            </Label>
                            <p className="text-xs text-muted-foreground">
                              Allow others to send friend requests
                            </p>
                          </div>
                          <Switch
                            id="allow_friend_requests"
                            checked={formData.allow_friend_requests ?? true}
                            onCheckedChange={(checked: boolean) =>
                              handleInputChange(
                                "allow_friend_requests" as keyof UserFormData,
                                checked,
                              )
                            }
                          />
                        </div>

                        <div className="flex items-center justify-between p-3 border rounded-lg">
                          <div className="flex-1">
                            <Label
                              htmlFor="allow_party_invites"
                              className="text-sm font-medium cursor-pointer"
                            >
                              Allow Party Invites
                            </Label>
                            <p className="text-xs text-muted-foreground">
                              Allow others to invite to parties
                            </p>
                          </div>
                          <Switch
                            id="allow_party_invites"
                            checked={formData.allow_party_invites ?? true}
                            onCheckedChange={(checked: boolean) =>
                              handleInputChange(
                                "allow_party_invites" as keyof UserFormData,
                                checked,
                              )
                            }
                          />
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="pt-4 border-t">
                    <div className="flex flex-col gap-4">
                      <div className="flex items-center gap-4">
                        <Avatar className="h-16 w-16">
                          <AvatarImage
                            src={avatarUrl}
                            alt={`${user.username}'s avatar`}
                          />
                          <AvatarFallback className="text-lg">
                            <IconUser className="h-6 w-6" />
                          </AvatarFallback>
                        </Avatar>
                        <div className="flex flex-col gap-2 flex-1">
                          <Label className="text-sm font-medium">Avatar</Label>
                          <p className="text-xs text-muted-foreground">
                            {user.has_avatar
                              ? "User has an avatar"
                              : "No avatar uploaded"}
                          </p>
                        </div>
                      </div>

                      <div className="flex gap-2">
                        <div className="flex-1">
                          <Input
                            id="avatar-upload"
                            type="file"
                            accept="image/*"
                            onChange={handleAvatarUpload}
                            disabled={isLoading}
                            className="text-xs"
                          />
                        </div>
                        {user.has_avatar && (
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={handleAvatarDelete}
                            disabled={isLoading}
                          >
                            <IconTrash className="h-4 w-4" />
                          </Button>
                        )}
                      </div>

                      <p className="text-xs text-muted-foreground">
                        Supported formats: JPG, PNG, GIF. Max size: 5MB
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-col gap-4 pt-4 border-t">
                    <div className="flex flex-col gap-2">
                      <Label className="text-xs text-muted-foreground">
                        User ID
                      </Label>
                      <div className="text-xs font-mono bg-muted px-2 py-1 rounded">
                        {user.id}
                      </div>
                    </div>
                    <div className="flex flex-col gap-2">
                      <Label className="text-xs text-muted-foreground">
                        Public ID
                      </Label>
                      <div className="flex items-center gap-2">
                        <div className="text-xs font-mono bg-muted px-2 py-1 rounded flex-1">
                          {user.public_id}
                        </div>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => setShowRegenerateDialog(true)}
                          disabled={isLoading}
                        >
                          <IconRefresh className="h-3 w-3" />
                        </Button>
                      </div>
                      <p className="text-xs text-muted-foreground">
                        Public ID is used for public-facing features
                      </p>
                    </div>
                    <div className="flex flex-col gap-2">
                      <Label className="text-xs text-muted-foreground">
                        Created At
                      </Label>
                      <div className="text-xs bg-muted px-2 py-1 rounded">
                        {user.created_at
                          ? new Date(user.created_at).toLocaleDateString()
                          : "N/A"}
                      </div>
                    </div>
                  </div>
                </form>
              </TabsContent>

              <TabsContent value="password" className="space-y-4 mt-4">
                <form
                  onSubmit={handleSubmitPassword}
                  className="flex flex-col gap-4"
                >
                  <div className="flex items-center gap-2 p-3 border rounded-lg">
                    <p className="text-sm flex-1">
                      You are about to change the password for{" "}
                      <strong>{user.username}</strong>. This action is{" "}
                      <strong>irreversible</strong>.
                    </p>
                  </div>

                  {user.role === "administrator" && (
                    <div className="flex items-center gap-2 p-3 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 rounded-lg">
                      <p className="text-sm text-red-800 dark:text-red-200 flex-1">
                        <strong>Attention:</strong> You are about to change the
                        password of an <strong>administrator</strong>. This
                        action requires special care and should only be done if
                        absolutely necessary.
                      </p>
                    </div>
                  )}

                  <div className="flex flex-col gap-3">
                    <Label htmlFor="password">New Password</Label>
                    <Input
                      id="password"
                      type="password"
                      value={passwordData.password}
                      onChange={(e) =>
                        handlePasswordChange("password", e.target.value)
                      }
                      placeholder="Enter new password"
                      required
                      minLength={8}
                    />
                  </div>

                  <div className="flex flex-col gap-3">
                    <Label htmlFor="confirmPassword">
                      Confirm New Password
                    </Label>
                    <Input
                      id="confirmPassword"
                      type="password"
                      value={passwordData.confirmPassword}
                      onChange={(e) =>
                        handlePasswordChange("confirmPassword", e.target.value)
                      }
                      placeholder="Confirm new password"
                      required
                      minLength={8}
                    />
                  </div>

                  {passwordErrors.length > 0 && (
                    <div className="p-3 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 rounded-lg">
                      <ul className="text-sm text-red-800 dark:text-red-200 space-y-1">
                        {passwordErrors.map((error) => (
                          <li key={error}>• {error}</li>
                        ))}
                      </ul>
                    </div>
                  )}

                  <div className="text-xs text-muted-foreground">
                    Password must be at least 8 characters long.
                  </div>
                </form>
              </TabsContent>
            </Tabs>
          </div>

          <SheetFooter className="px-4 pb-4">
            <div className="flex gap-2">
              {activeTab === "general" ? (
                <Button
                  onClick={handleSubmitUser}
                  disabled={isLoading}
                  className="flex-1"
                >
                  {isLoading ? "Updating..." : "Update User"}
                </Button>
              ) : (
                <Button
                  onClick={handleSubmitPassword}
                  disabled={isLoading}
                  className="flex-1"
                  variant="destructive"
                >
                  {isLoading ? "Updating..." : "Update Password"}
                </Button>
              )}
            </div>
            <SheetClose asChild>
              <Button variant="outline" disabled={isLoading}>
                Cancel
              </Button>
            </SheetClose>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <Dialog
        open={showRegenerateDialog}
        onOpenChange={setShowRegenerateDialog}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Regenerate Public ID</DialogTitle>
            <DialogDescription>
              Are you sure you want to regenerate the public ID for{" "}
              <strong>{user.username}</strong>?
            </DialogDescription>
          </DialogHeader>

          <div className="flex items-start gap-2 p-3 bg-yellow-50 dark:bg-yellow-950/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
            <IconAlertTriangle className="h-4 w-4 text-yellow-600 dark:text-yellow-400 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm text-yellow-800 dark:text-yellow-200">
                <strong>Warning:</strong> This action is{" "}
                <strong>irreversible</strong>. The old public ID will no longer
                work.
              </p>
              <p className="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
                Current public ID:{" "}
                <code className="font-mono">{user.public_id}</code>
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowRegenerateDialog(false)}
              disabled={isRegenerating}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleRegeneratePublicId}
              disabled={isRegenerating}
            >
              {isRegenerating ? "Regenerating..." : "Regenerate Public ID"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
