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
import { ScrollArea } from "@/components/ui/scroll-area";
import { geographyService } from "@/lib/services/geography.service";
import type { Continent } from "@/lib/types/geography.types";
import {
  IconDeviceFloppy,
  IconRefresh,
  IconWorld,
} from "@tabler/icons-react";
import * as React from "react";
import { toast } from "sonner";

interface ContinentEditDialogProps {
  continent: Continent | null;
  datasetId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave?: (continent: Continent) => void;
}

export function ContinentEditDialog({
  continent,
  datasetId,
  open,
  onOpenChange,
  onSave,
}: ContinentEditDialogProps) {
  const [saving, setSaving] = React.useState(false);
  const [formData, setFormData] = React.useState<Partial<Continent>>({});

  React.useEffect(() => {
    if (continent) {
      setFormData({ ...continent });
    }
  }, [continent]);

  if (!continent) return null;

  const name = (formData.name || continent.name) as Record<string, string>;

  const updateName = (lang: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      name: { ...name, [lang]: value },
    }));
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const updatedContinent = await geographyService.updateContinent(
        datasetId,
        continent.slug,
        formData,
      );
      toast.success("Continent updated successfully");
      onSave?.(updatedContinent);
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to save continent:", error);
      toast.error("Failed to update continent");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-3">
            <IconWorld className="h-6 w-6 text-primary" />
            <span>Edit: {name?.en || continent.slug}</span>
          </DialogTitle>
          <DialogDescription>Modify continent information</DialogDescription>
        </DialogHeader>

        <ScrollArea className="max-h-[60vh] pr-4">
          <div className="space-y-6">
            <div className="space-y-4">
              <h4 className="text-sm font-semibold">Continent Names</h4>
              <div className="grid grid-cols-2 gap-4">
                {Object.entries(name || {}).map(([lang, value]) => (
                  <div key={lang} className="space-y-2">
                    <Label className="uppercase text-xs">{lang}</Label>
                    <Input
                      value={value}
                      onChange={(e) => updateName(lang, e.target.value)}
                    />
                  </div>
                ))}
              </div>
            </div>

            <div className="space-y-4">
              <h4 className="text-sm font-semibold">Statistics</h4>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Population</Label>
                  <Input
                    type="number"
                    value={formData.population ?? continent.population}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        population: Number.parseInt(e.target.value, 10) || 0,
                      }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Area (km²)</Label>
                  <Input
                    type="number"
                    value={formData.area_km2 ?? continent.area_km2}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        area_km2: Number.parseInt(e.target.value, 10) || 0,
                      }))
                    }
                  />
                </div>
              </div>
            </div>
          </div>
        </ScrollArea>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving ? (
              <>
                <IconRefresh className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <IconDeviceFloppy className="mr-2 h-4 w-4" />
                Save Changes
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
