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
import { useIsMobile } from "@/hooks/useMobile";
import { importsService } from "@/lib/services/imports.service";
import type { ImportJob } from "@/lib/types/datasets.types";
import {
  IconCheck,
  IconCircleCheck,
  IconCircleX,
  IconClock,
  IconCopy,
} from "@tabler/icons-react";
import { format } from "date-fns";
import { enUS } from "date-fns/locale";
import * as React from "react";
import { useLocation, useNavigate } from "react-router";

interface ImportDetailsDialogProps {
  viewId: string | null;
}

export function ImportDetailsDialog({ viewId }: ImportDetailsDialogProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const isOpen = !!viewId;
  const isMobile = useIsMobile();
  const [job, setJob] = React.useState<ImportJob | null>(null);
  const [_jobLoading, setJobLoading] = React.useState(false);
  const [copiedField, setCopiedField] = React.useState<string | null>(null);

  const handleIsOpenChange = (open: boolean) => {
    if (!open) {
      navigate(location.pathname, { replace: true });
    }
  };

  React.useEffect(() => {
    if (!viewId) {
      setJob(null);
      return;
    }

    const fetchJob = async () => {
      setJobLoading(true);
      try {
        const jobData = await importsService.getImportById(viewId);
        setJob(jobData);
      } catch (error) {
        console.error("Failed to fetch job details:", error);
        setJob(null);
      } finally {
        setJobLoading(false);
      }
    };

    fetchJob();
  }, [viewId]);

  const copyToClipboard = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
  };

  if (!job) return null;

  const isGeography = job.dataset === "geography";
  const isFlagsInProgress =
    isGeography && job.flags_started_at && !job.flags_finished_at;

  const getDuration = () => {
    if (!job.finished_at) return "In progress...";
    const start = new Date(job.started_at);
    const end = new Date(job.finished_at);
    const diff = end.getTime() - start.getTime();
    const seconds = Math.floor(diff / 1000);

    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}m ${remainingSeconds}s`;
  };

  const getFlagsDuration = () => {
    if (!job.flags_started_at || !job.flags_finished_at) return null;
    const start = new Date(job.flags_started_at);
    const end = new Date(job.flags_finished_at);
    const diff = end.getTime() - start.getTime();
    const seconds = Math.floor(diff / 1000);

    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}m ${remainingSeconds}s`;
  };

  return (
    <Drawer
      direction={isMobile ? "bottom" : "right"}
      open={isOpen}
      onOpenChange={handleIsOpenChange}
    >
      <DrawerContent className="max-w-2xl">
        <DrawerHeader className="gap-1">
          <DrawerTitle>Import details</DrawerTitle>
          <DrawerDescription>
            Import of dataset {job.dataset} v{job.version}
          </DrawerDescription>
        </DrawerHeader>

        <div className="flex flex-col gap-6 overflow-y-auto px-4 pb-4">
          <div className="space-y-4">
            <div className="flex flex-col gap-4">
              <div>
                <div className="text-sm font-medium text-muted-foreground">
                  Status
                </div>
                <div className="mt-1">
                  {isFlagsInProgress ? (
                    <Badge variant="secondary">
                      <IconClock className="mr-1 h-3 w-3 animate-spin" />
                      Importing flags...
                    </Badge>
                  ) : job.finished_at ? (
                    job.success ? (
                      <Badge variant="outline">
                        <IconCircleCheck className="mr-1 h-3 w-3" />
                        Success
                      </Badge>
                    ) : (
                      <Badge variant="destructive">
                        <IconCircleX className="mr-1 h-3 w-3" />
                        Failed
                      </Badge>
                    )
                  ) : (
                    <Badge variant="secondary">
                      <IconClock className="mr-1 h-3 w-3 animate-spin" />
                      In progress
                    </Badge>
                  )}
                </div>
              </div>

              <div>
                <div className="text-sm font-medium text-muted-foreground">
                  Duration
                </div>
                <div className="mt-1 text-sm">{getDuration()}</div>
              </div>

              <div>
                <div className="text-sm font-medium text-muted-foreground">
                  Start
                </div>
                <div className="mt-1 text-sm">
                  {format(new Date(job.started_at), "PPpp", {
                    locale: enUS,
                  })}
                </div>
              </div>

              {job.finished_at && (
                <div>
                  <div className="text-sm font-medium text-muted-foreground">
                    End
                  </div>
                  <div className="mt-1 text-sm">
                    {format(new Date(job.finished_at), "PPpp", {
                      locale: enUS,
                    })}
                  </div>
                </div>
              )}

              <div>
                <div className="text-sm font-medium text-muted-foreground">
                  Manifest URL
                </div>
                <div className="mt-1 flex items-center gap-2">
                  <div
                    className="text-sm truncate flex-1"
                    title={job.manifest_url}
                  >
                    {job.manifest_url}
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() =>
                      copyToClipboard(job.manifest_url, "manifestUrl")
                    }
                  >
                    {copiedField === "manifestUrl" ? (
                      <IconCheck className="h-3 w-3" />
                    ) : (
                      <IconCopy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
            </div>

            <div className="border rounded-lg p-4">
              <h3 className="font-medium mb-3">Results</h3>
              <div className="grid grid-cols-2 gap-4 text-muted-foreground">
                <div>
                  <div className="text-2xl font-bold">{job.added}</div>
                  <div className="text-sm text-muted-foreground">
                    {isGeography ? "Countries added" : "Questions added"}
                  </div>
                </div>
                <div>
                  <div className="text-2xl font-bold">{job.errors}</div>
                  <div className="text-sm text-muted-foreground">Errors</div>
                </div>
              </div>

              {isGeography && (
                <div className="mt-4 pt-4 border-t">
                  <div className="text-sm font-medium text-muted-foreground mb-3">
                    Flags
                  </div>
                  {isFlagsInProgress ? (
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <IconClock className="h-4 w-4 animate-spin" />
                      Importing flags...
                    </div>
                  ) : (
                    <div className="grid grid-cols-3 gap-4">
                      <div>
                        <div className="text-xl font-bold">
                          {job.flags_svg_count}
                        </div>
                        <div className="text-xs text-muted-foreground">SVG</div>
                      </div>
                      <div>
                        <div className="text-xl font-bold">
                          {job.flags_png512_count}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          PNG 512
                        </div>
                      </div>
                      <div>
                        <div className="text-xl font-bold">
                          {job.flags_png1024_count}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          PNG 1024
                        </div>
                      </div>
                    </div>
                  )}
                  {job.flags_finished_at && getFlagsDuration() && (
                    <div className="mt-3 text-xs text-muted-foreground">
                      Flags imported in {getFlagsDuration()}
                    </div>
                  )}
                </div>
              )}
            </div>

            {job.message && (
              <div className="border rounded-lg p-4">
                <h3 className="font-medium mb-2">Message</h3>
                <p className="text-sm text-muted-foreground">{job.message}</p>
              </div>
            )}
          </div>
        </div>

        <DrawerFooter>
          <DrawerClose asChild>
            <Button variant="outline">Close</Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
