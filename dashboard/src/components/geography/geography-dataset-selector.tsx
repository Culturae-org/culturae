"use client";

import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { GeographyDataset } from "@/lib/types/geography.types";
import { IconDatabase } from "@tabler/icons-react";

interface GeographyDatasetSelectorProps {
  datasets: GeographyDataset[];
  selectedDatasetId: string;
  onDatasetChange: (datasetId: string) => void;
}

export function GeographyDatasetSelector({
  datasets,
  selectedDatasetId,
  onDatasetChange,
}: GeographyDatasetSelectorProps) {
  const activeDatasets = datasets.filter((d) => d.is_active);

  if (activeDatasets.length <= 1 && !selectedDatasetId) {
    return null;
  }

  return (
    <>
      <div className="h-8 w-px bg-border mx-2 hidden sm:block" />
      <IconDatabase className="h-4 w-4 text-muted-foreground" />
      <Select value={selectedDatasetId} onValueChange={onDatasetChange}>
        <SelectTrigger className="w-[250px]">
          <SelectValue placeholder="Select dataset" />
        </SelectTrigger>
        <SelectContent>
          {activeDatasets.map((dataset) => (
            <SelectItem key={dataset.id} value={dataset.id}>
              {dataset.name}
              {dataset.is_default && (
                <Badge variant="secondary" className="text-xs ml-2">
                  Default
                </Badge>
              )}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </>
  );
}
