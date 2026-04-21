"use client";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useSettings } from "@/hooks/useSettings";
import { SETTINGS_ENDPOINTS } from "@/lib/api/endpoints";
import type { AvatarConfig } from "@/lib/types/settings.types";
import { IconRefresh } from "@tabler/icons-react";
import * as React from "react";
import { InfoHover } from "./info-hover";

const DEFAULTS: AvatarConfig = {
  max_file_size_mb: 5,
  allowed_mime_types: ["image/jpeg", "image/png"],
  allowed_extensions: [".png", ".jpeg", ".jpg"],
};

export function AvatarConfigCard() {
  const [mimeTypesStr, setMimeTypesStr] = React.useState(
    DEFAULTS.allowed_mime_types.join(", "),
  );
  const [extensionsStr, setExtensionsStr] = React.useState(
    DEFAULTS.allowed_extensions.join(", "),
  );

  const {
    config,
    updateProperty,
    save: baseSave,
    loading,
    initialLoading,
  } = useSettings<AvatarConfig>({
    endpoint: SETTINGS_ENDPOINTS.AVATAR,
    defaults: DEFAULTS,
    validate: (cfg) => {
      if (cfg.max_file_size_mb <= 0) {
        return "Max file size must be positive";
      }
      if (
        cfg.allowed_mime_types.length === 0 ||
        cfg.allowed_extensions.length === 0
      ) {
        return "At least one MIME type and extension are required";
      }
      return true;
    },
  });

  // Sync string representations when config changes
  React.useEffect(() => {
    setMimeTypesStr(config.allowed_mime_types.join(", "));
    setExtensionsStr(config.allowed_extensions.join(", "));
  }, [config.allowed_mime_types, config.allowed_extensions]);

  const save = async () => {
    const parsedMimes = mimeTypesStr
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    const parsedExts = extensionsStr
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);

    return baseSave({
      ...config,
      allowed_mime_types: parsedMimes,
      allowed_extensions: parsedExts,
    });
  };

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Avatar Upload Settings
        </CardTitle>
        <CardDescription>
          Configure file size limits and allowed formats for avatar uploads.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Label htmlFor="avatar-max-size">Max file size</Label>
            <InfoHover description="Maximum allowed file size for avatar uploads. Larger files will be rejected during upload." />
          </div>
          <Input
            id="avatar-max-size"
            type="number"
            min={1}
            value={config.max_file_size_mb}
            onChange={(e) =>
              updateProperty("max_file_size_mb", Number(e.target.value))
            }
            disabled={initialLoading}
            className="max-w-xs"
          />
          <p className="text-xs text-muted-foreground">MB</p>
        </div>
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Label htmlFor="avatar-mime-types">Allowed MIME types</Label>
            <InfoHover description="Comma-separated list of accepted MIME types (e.g., image/jpeg, image/png). Files with other MIME types will be rejected." />
          </div>
          <Input
            id="avatar-mime-types"
            value={mimeTypesStr}
            onChange={(e) => setMimeTypesStr(e.target.value)}
            disabled={initialLoading}
            placeholder="image/jpeg, image/png"
            className="max-w-xs"
          />
        </div>
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Label htmlFor="avatar-extensions">Allowed extensions</Label>
            <InfoHover description="Comma-separated list of accepted file extensions (include the dot, e.g., .jpg, .jpeg, .png). Files with other extensions will be rejected." />
          </div>
          <Input
            id="avatar-extensions"
            value={extensionsStr}
            onChange={(e) => setExtensionsStr(e.target.value)}
            disabled={initialLoading}
            placeholder=".jpg, .jpeg, .png"
            className="max-w-xs"
          />
        </div>
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Max size: {config.max_file_size_mb} MB · Types:{" "}
            {mimeTypesStr || "—"}
          </p>
          <Button onClick={() => save()} disabled={loading || initialLoading}>
            {loading && <IconRefresh className="h-4 w-4 mr-2 animate-spin" />}
            Save Changes
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
