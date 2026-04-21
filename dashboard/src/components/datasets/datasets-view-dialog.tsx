"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Skeleton } from "@/components/ui/skeleton";
import type { UnifiedDataset } from "@/hooks/useDatasetsList";
import { useIsMobile } from "@/hooks/useMobile";
import { datasetsService } from "@/lib/services/datasets.service";
import {
  IconArrowRight,
  IconCalendar,
  IconCheck,
  IconCopy,
  IconFileText,
  IconFlag,
  IconMapPin,
  IconWorld,
} from "@tabler/icons-react";
import { formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import * as React from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";

interface DatasetDetailsDialogProps {
  datasetId: string | null;
}

export function DatasetsViewDialog({ datasetId }: DatasetDetailsDialogProps) {
  const navigate = useNavigate();
  const isMobile = useIsMobile();
  const [dataset, setDataset] = React.useState<UnifiedDataset | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [copiedField, setCopiedField] = React.useState<string | null>(null);

  const isOpen = !!datasetId;

  const handleIsOpenChange = (open: boolean) => {
    if (!open) {
      navigate("?");
    }
  };

  React.useEffect(() => {
    if (!datasetId) {
      setDataset(null);
      return;
    }

    const fetchDataset = async () => {
      setLoading(true);
      try {
        const data = await datasetsService.getDatasetBySlug(datasetId);
        setDataset(data);
      } catch (error) {
        console.error("Failed to fetch dataset details:", error);
        toast.error("Failed to load dataset details");
      } finally {
        setLoading(false);
      }
    };

    fetchDataset();
  }, [datasetId]);

  const copyToClipboard = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch (_err) {
      toast.error("Failed to copy");
    }
  };

  if (!dataset) return null;

  const isGeography = dataset.type === "geography";
  const _statsLabel = isGeography ? "Countries" : "Questions";
  const _statsValue = isGeography
    ? dataset.country_count || 0
    : dataset.question_count || 0;

  const content = (
    <div className="space-y-4">
      {loading && (
        <div className="space-y-3">
          <Skeleton className="h-8 w-1/2" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
      )}

      {!loading && (
        <>
          <div className="pb-4">
            <div className="flex items-start justify-between gap-4 mb-2">
              <h2 className="text-2xl font-bold">{dataset.name}</h2>
              <Badge variant={isGeography ? "secondary" : "default"}>
                {isGeography ? (
                  <>
                    <IconWorld className="mr-1 h-3 w-3" /> Geography
                  </>
                ) : (
                  <>
                    <IconFileText className="mr-1 h-3 w-3" /> Questions
                  </>
                )}
              </Badge>
            </div>
            {dataset.description && (
              <p className="text-sm text-muted-foreground">
                {dataset.description}
              </p>
            )}
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between py-2">
              <span className="text-sm text-muted-foreground">Name</span>
              <span className="text-sm font-medium">{dataset.name}</span>
            </div>

            <div className="flex items-center justify-between py-2">
              <span className="text-sm text-muted-foreground">Type</span>
              <Badge variant="outline">
                {isGeography ? "Geography" : "Questions"}
              </Badge>
            </div>

            <div className="flex items-center justify-between py-2">
              <span className="text-sm text-muted-foreground">Source</span>
              <Badge variant="outline">{dataset.source}</Badge>
            </div>

            <div className="flex items-center justify-between py-2">
              <span className="text-sm text-muted-foreground">Version</span>
              <div className="flex items-center gap-2">
                <code className="text-sm bg-muted px-2 py-1 rounded font-mono">
                  {dataset.version}
                </code>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 w-6 p-0"
                  onClick={() => copyToClipboard(dataset.version, "version")}
                >
                  {copiedField === "version" ? (
                    <IconCheck className="h-4 w-4 text-green-500" />
                  ) : (
                    <IconCopy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>

            <div className="flex items-center justify-between py-2">
              <span className="text-sm text-muted-foreground">Status</span>
              <Badge variant={dataset.is_active ? "default" : "secondary"}>
                {dataset.is_active ? "Active" : "Inactive"}
              </Badge>
            </div>

            <div className="py-2 space-y-2">
              <div className="text-sm text-muted-foreground font-medium">
                Statistics
              </div>
              {isGeography ? (
                // Geography statistics
                <div className="space-y-2 pl-4">
                  {dataset.country_count !== undefined && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground flex items-center gap-1">
                        <IconWorld className="h-4 w-4" /> Countries
                      </span>
                      <span className="text-sm font-medium">
                        {dataset.country_count}
                      </span>
                    </div>
                  )}
                  {dataset.continent_count !== undefined && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground flex items-center gap-1">
                        <IconMapPin className="h-4 w-4" /> Continents
                      </span>
                      <span className="text-sm font-medium">
                        {dataset.continent_count}
                      </span>
                    </div>
                  )}
                  {dataset.region_count !== undefined && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground flex items-center gap-1">
                        <IconMapPin className="h-4 w-4" /> Regions
                      </span>
                      <span className="text-sm font-medium">
                        {dataset.region_count}
                      </span>
                    </div>
                  )}
                  {dataset.flag_count !== undefined && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground flex items-center gap-1">
                        <IconFlag className="h-4 w-4" /> Flags
                      </span>
                      <span className="text-sm font-medium">
                        {dataset.flag_count}
                      </span>
                    </div>
                  )}
                  {dataset.flag_png512_count !== undefined && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground text-xs">
                        Flag PNG 512px
                      </span>
                      <span className="text-sm font-medium">
                        {dataset.flag_png512_count}
                      </span>
                    </div>
                  )}
                  {dataset.flag_png1024_count !== undefined && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground text-xs">
                        Flag PNG 1024px
                      </span>
                      <span className="text-sm font-medium">
                        {dataset.flag_png1024_count}
                      </span>
                    </div>
                  )}
                </div>
              ) : (
                // Question statistics
                <div className="space-y-2 pl-4">
                  {dataset.question_count !== undefined && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground flex items-center gap-1">
                        <IconFileText className="h-4 w-4" /> Questions
                      </span>
                      <span className="text-sm font-medium">
                        {dataset.question_count}
                      </span>
                    </div>
                  )}
                  {dataset.theme_count !== undefined && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground">
                        Themes
                      </span>
                      <span className="text-sm font-medium">
                        {dataset.theme_count}
                      </span>
                    </div>
                  )}
                </div>
              )}
            </div>

            <div className="py-2 space-y-2">
              <div className="text-sm text-muted-foreground font-medium">
                Dates
              </div>
              <div className="flex items-center justify-between pl-4">
                <span className="text-sm text-muted-foreground flex items-center gap-1">
                  <IconCalendar className="h-4 w-4" /> Imported
                </span>
                <span className="text-sm">
                  {formatDistanceToNow(new Date(dataset.imported_at), {
                    addSuffix: true,
                    locale: enUS,
                  })}
                </span>
              </div>
              <div className="flex items-center justify-between pl-4">
                <span className="text-sm text-muted-foreground flex items-center gap-1">
                  <IconCalendar className="h-4 w-4" /> Created
                </span>
                <span className="text-sm">
                  {formatDistanceToNow(new Date(dataset.created_at), {
                    addSuffix: true,
                    locale: enUS,
                  })}
                </span>
              </div>
              <div className="flex items-center justify-between pl-4">
                <span className="text-sm text-muted-foreground flex items-center gap-1">
                  <IconCalendar className="h-4 w-4" /> Updated
                </span>
                <span className="text-sm">
                  {formatDistanceToNow(new Date(dataset.updated_at), {
                    addSuffix: true,
                    locale: enUS,
                  })}
                </span>
              </div>
            </div>

            <div className="py-2 space-y-2">
              <div className="text-sm text-muted-foreground font-medium">
                Identifiers
              </div>
              <div className="space-y-2 pl-4">
                <div className="flex flex-col gap-1">
                  <span className="text-xs text-muted-foreground">Slug</span>
                  <div className="flex items-center justify-between gap-2">
                    <code className="text-xs bg-muted px-2 py-1 rounded flex-1 overflow-auto font-mono">
                      {dataset.slug}
                    </code>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0"
                      onClick={() => copyToClipboard(dataset.slug, "slug")}
                    >
                      {copiedField === "slug" ? (
                        <IconCheck className="h-4 w-4 text-green-500" />
                      ) : (
                        <IconCopy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
                <div className="flex flex-col gap-1">
                  <span className="text-xs text-muted-foreground">ID</span>
                  <div className="flex items-center justify-between gap-2">
                    <code className="text-xs bg-muted px-2 py-1 rounded flex-1 overflow-auto font-mono">
                      {dataset.id}
                    </code>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0"
                      onClick={() => copyToClipboard(dataset.id, "id")}
                    >
                      {copiedField === "id" ? (
                        <IconCheck className="h-4 w-4 text-green-500" />
                      ) : (
                        <IconCopy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
                {(() => {
                  const jobId = dataset.import_job_id;
                  return jobId ? (
                    <>
                      <div className="flex flex-col gap-1">
                        <span className="text-xs text-muted-foreground">
                          Import Job ID
                        </span>
                        <div className="flex items-center justify-between gap-2">
                          <code className="text-xs bg-muted px-2 py-1 rounded flex-1 overflow-auto font-mono">
                            {jobId}
                          </code>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 w-6 p-0"
                            onClick={() =>
                              copyToClipboard(jobId, "import_job_id")
                            }
                          >
                            {copiedField === "import_job_id" ? (
                              <IconCheck className="h-4 w-4 text-green-500" />
                            ) : (
                              <IconCopy className="h-4 w-4" />
                            )}
                          </Button>
                        </div>
                      </div>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => navigate(`?view=${jobId}`)}
                        className="w-full gap-2"
                      >
                        Go to the job details
                        <IconArrowRight className="h-4 w-4" />
                      </Button>
                    </>
                  ) : null;
                })()}
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );

  if (isMobile) {
    return (
      <Drawer open={isOpen} onOpenChange={handleIsOpenChange}>
        <DrawerContent>
          <DrawerHeader>
            <DrawerTitle>Dataset Details</DrawerTitle>
            <DrawerDescription>
              View detailed information about this dataset
            </DrawerDescription>
          </DrawerHeader>
          <div className="px-4 pb-4 overflow-y-auto max-h-[70vh]">
            {content}
          </div>
          <DrawerFooter>
            <DrawerClose asChild>
              <Button variant="outline" className="w-full">
                Close
              </Button>
            </DrawerClose>
          </DrawerFooter>
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <Drawer direction="right" open={isOpen} onOpenChange={handleIsOpenChange}>
      <DrawerContent className="max-w-2xl">
        <DrawerHeader className="gap-1">
          <DrawerTitle>Dataset Details</DrawerTitle>
          <DrawerDescription>
            View detailed information about this dataset
          </DrawerDescription>
          <DrawerClose />
        </DrawerHeader>
        <div className="px-4 pb-4 overflow-y-auto">{content}</div>
        <DrawerFooter>
          <DrawerClose asChild>
            <Button variant="outline">Close</Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
