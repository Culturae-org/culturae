"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { useViewDialog } from "@/hooks/useViewDialog";
import type { Continent } from "@/lib/types/geography.types";
import {
  IconClock,
  IconMapPin,
  IconMaximize,
  IconUsers,
  IconWorld,
} from "@tabler/icons-react";

interface ContinentViewDialogProps {
  continent: Continent | null;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onEdit?: (continent: Continent) => void;
}

export function ContinentViewDialog({
  continent,
  open: controlledOpen,
  onOpenChange,
  onEdit,
}: ContinentViewDialogProps) {
  const viewDialog = useViewDialog("view");
  const isOpen =
    controlledOpen !== undefined ? controlledOpen : viewDialog.isOpen;
  const handleIsOpenChange = (open: boolean) => {
    if (onOpenChange) {
      onOpenChange(open);
    } else if (!open) {
      viewDialog.close();
    }
  };
  if (!continent) return null;

  const name = continent.name as Record<string, string>;

  const formatNumber = (num: number | undefined) => {
    if (num === undefined || num === null) return "N/A";
    return num.toLocaleString();
  };

  const formatArea = (area: number | undefined) => {
    if (area === undefined || area === null) return "N/A";
    return `${area.toLocaleString()} km²`;
  };

  const formatDate = (date: string | undefined) => {
    if (!date) return "N/A";
    return new Date(date).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleIsOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-3">
            <IconWorld className="h-6 w-6 text-primary" />
            <span>{name?.en || continent.slug}</span>
          </DialogTitle>
          <DialogDescription>Continent Details</DialogDescription>
        </DialogHeader>

        <ScrollArea className="max-h-[60vh] pr-4">
          <div className="space-y-6">
            <div className="space-y-2">
              <h4 className="text-sm font-semibold flex items-center gap-2">
                <IconWorld className="h-4 w-4" />
                Names (i18n)
              </h4>
              <div className="grid grid-cols-2 gap-2">
                {Object.entries(name || {}).map(([lang, value]) => (
                  <div key={lang} className="flex items-center gap-2">
                    <Badge
                      variant="outline"
                      className="uppercase text-xs w-10 justify-center"
                    >
                      {lang}
                    </Badge>
                    <span className="text-sm">{value}</span>
                  </div>
                ))}
              </div>
            </div>

            <Separator />

            <div className="space-y-2">
              <h4 className="text-sm font-semibold flex items-center gap-2">
                <IconMapPin className="h-4 w-4" />
                Statistics
              </h4>
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-muted/50 rounded-lg p-4">
                  <div className="flex items-center gap-2 text-muted-foreground text-xs">
                    <IconUsers className="h-3 w-3" />
                    Population
                  </div>
                  <div className="text-lg font-semibold">
                    {formatNumber(continent.population)}
                  </div>
                </div>
                <div className="bg-muted/50 rounded-lg p-4">
                  <div className="flex items-center gap-2 text-muted-foreground text-xs">
                    <IconMaximize className="h-3 w-3" />
                    Area
                  </div>
                  <div className="text-lg font-semibold">
                    {formatArea(continent.area_km2)}
                  </div>
                </div>
              </div>
            </div>

            <Separator />

            <div className="space-y-2">
              <h4 className="text-sm font-semibold flex items-center gap-2">
                <IconMapPin className="h-4 w-4" />
                Countries ({continent.countries?.length || 0})
              </h4>
              <div className="flex flex-wrap gap-1">
                {continent.countries?.slice(0, 50).map((country) => (
                  <Badge key={country} variant="secondary" className="text-xs">
                    {country}
                  </Badge>
                ))}
                {(continent.countries?.length || 0) > 50 && (
                  <Badge variant="outline" className="text-xs">
                    +{(continent.countries?.length || 0) - 50} more
                  </Badge>
                )}
              </div>
            </div>

            <Separator />

            <div className="space-y-2">
              <h4 className="text-sm font-semibold flex items-center gap-2">
                <IconClock className="h-4 w-4" />
                Technical Details
              </h4>
              <div className="grid grid-cols-2 gap-y-2 text-sm">
                <div className="text-muted-foreground">ID</div>
                <div className="font-mono text-xs">{continent.id}</div>
                <div className="text-muted-foreground">Slug</div>
                <div className="font-mono text-xs">{continent.slug}</div>
                <div className="text-muted-foreground">Dataset ID</div>
                <div className="font-mono text-xs">{continent.dataset_id}</div>
                <div className="text-muted-foreground">Created</div>
                <div>{formatDate(continent.created_at)}</div>
                <div className="text-muted-foreground">Updated</div>
                <div>{formatDate(continent.updated_at)}</div>
              </div>
            </div>
          </div>
        </ScrollArea>

        <div className="flex justify-end gap-2 mt-4">
          {onEdit && (
            <Button variant="outline" onClick={() => onEdit(continent)}>
              Edit
            </Button>
          )}
          <Button onClick={() => handleIsOpenChange(false)}>Close</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
