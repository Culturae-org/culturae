"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { IconBan, IconClock } from "@tabler/icons-react";
import * as React from "react";

const BAN_PRESETS = [
  { label: "1 hour", value: "1h" },
  { label: "5 hours", value: "5h" },
  { label: "24 hours", value: "24h" },
  { label: "3 days", value: "72h" },
  { label: "7 days", value: "168h" },
  { label: "30 days", value: "720h" },
  { label: "Permanent", value: "permanent" },
] as const;

interface UserBanDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  username: string;
  onConfirm: (duration: string, reason: string) => void;
  loading?: boolean;
}

export function UserBanDialog({
  open,
  onOpenChange,
  username,
  onConfirm,
  loading,
}: UserBanDialogProps) {
  const [selectedPreset, setSelectedPreset] = React.useState<string | null>(
    null,
  );
  const [customDuration, setCustomDuration] = React.useState("");
  const [customUnit, setCustomUnit] = React.useState<"h" | "d">("h");
  const [reason, setReason] = React.useState("");
  const [useCustom, setUseCustom] = React.useState(false);

  React.useEffect(() => {
    if (open) {
      setSelectedPreset(null);
      setCustomDuration("");
      setCustomUnit("h");
      setReason("");
      setUseCustom(false);
    }
  }, [open]);

  const getDuration = (): string | null => {
    if (useCustom) {
      const num = Number.parseInt(customDuration, 10);
      if (!num || num <= 0) return null;
      return `${num}${customUnit}`;
    }
    return selectedPreset;
  };

  const handleConfirm = () => {
    const duration = getDuration();
    if (!duration) return;
    onConfirm(duration, reason);
  };

  const isValid = getDuration() !== null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <IconBan className="h-5 w-5 text-red-500" />
            Ban {username}
          </DialogTitle>
          <DialogDescription>
            Choose a ban duration. The user will be immediately logged out and
            unable to access the platform.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label>Duration</Label>
            <div className="grid grid-cols-4 gap-2">
              {BAN_PRESETS.map((preset) => (
                <Button
                  key={preset.value}
                  type="button"
                  variant={
                    !useCustom && selectedPreset === preset.value
                      ? "default"
                      : "outline"
                  }
                  size="sm"
                  className={`text-xs ${preset.value === "permanent" ? "col-span-2 text-red-600 border-red-200 hover:bg-red-50 dark:text-red-400 dark:border-red-800 dark:hover:bg-red-950" : ""} ${!useCustom && selectedPreset === preset.value && preset.value === "permanent" ? "bg-red-600 text-white hover:bg-red-700 dark:bg-red-700" : ""}`}
                  onClick={() => {
                    setSelectedPreset(preset.value);
                    setUseCustom(false);
                  }}
                >
                  {preset.label}
                </Button>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant={useCustom ? "default" : "outline"}
                size="sm"
                className="text-xs"
                onClick={() => setUseCustom(true)}
              >
                <IconClock className="h-3 w-3 mr-1" />
                Custom duration
              </Button>
            </div>
            {useCustom && (
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  min={1}
                  placeholder="Duration"
                  value={customDuration}
                  onChange={(e) => setCustomDuration(e.target.value)}
                  className="w-24"
                />
                <div className="flex rounded-md border">
                  <Button
                    type="button"
                    variant={customUnit === "h" ? "default" : "ghost"}
                    size="sm"
                    className="rounded-r-none text-xs"
                    onClick={() => setCustomUnit("h")}
                  >
                    Hours
                  </Button>
                  <Button
                    type="button"
                    variant={customUnit === "d" ? "default" : "ghost"}
                    size="sm"
                    className="rounded-l-none text-xs"
                    onClick={() => setCustomUnit("d")}
                  >
                    Days
                  </Button>
                </div>
              </div>
            )}
          </div>

          <div className="space-y-2">
            <Label>Reason (optional)</Label>
            <Textarea
              placeholder="Reason for the ban..."
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={2}
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleConfirm}
            disabled={!isValid || loading}
          >
            {loading ? "Banning..." : "Ban User"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
