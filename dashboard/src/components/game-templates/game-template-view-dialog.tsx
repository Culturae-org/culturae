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
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useDatasets } from "@/hooks/useDatasets";
import { useIsMobile } from "@/hooks/useMobile";
import {
  CATEGORY_LABELS,
  FLAG_VARIANT_LABELS,
  MODE_LABELS,
  QUESTION_TYPE_LABELS,
  SCORE_MODE_LABELS,
} from "@/lib/constants/game-template.constants";
import type { GameTemplate } from "@/lib/types/game-template.types";
import {
  IconCheck,
  IconCopy,
  IconEdit,
  IconFlag,
  IconLayoutGrid,
  IconWorld,
  IconX,
} from "@tabler/icons-react";
import * as React from "react";

function StatCard({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex flex-col items-center gap-1 p-3 rounded-lg bg-muted/50">
      <div className="text-sm font-bold">{value}</div>
      <div className="text-[10px] text-muted-foreground">{label}</div>
    </div>
  );
}

interface Props {
  template: GameTemplate | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onEditClick?: () => void;
}

export function GameTemplateViewDialog({
  template,
  open,
  onOpenChange,
  onEditClick,
}: Props) {
  const isMobile = useIsMobile();
  const { datasets } = useDatasets();
  const [copiedSlug, setCopiedSlug] = React.useState(false);
  const [copiedId, setCopiedId] = React.useState(false);

  if (!template) return null;

  const dataset = datasets.find((d) => d.id === template.dataset_id);
  const templateSlug = template.slug;
  const templateId = template.id;

  function copyToClipboard(text: string | undefined, field: string) {
    if (!text) return;
    navigator.clipboard.writeText(text).then(() => {
      if (field === "slug") setCopiedSlug(true);
      if (field === "id") setCopiedId(true);
      setTimeout(() => {
        if (field === "slug") setCopiedSlug(false);
        if (field === "id") setCopiedId(false);
      }, 2000);
    });
  }

  function copySlugID() {
    copyToClipboard(templateSlug, "slug");
  }

  function copyTemplateID() {
    copyToClipboard(templateId, "id");
  }

  function formatDate(s: string) {
    return new Date(s).toLocaleDateString("fr-FR", {
      day: "numeric",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  return (
    <Drawer
      direction={isMobile ? "bottom" : "right"}
      open={open}
      onOpenChange={onOpenChange}
    >
      <DrawerContent>
        <DrawerHeader className="gap-1">
          <div className="flex items-center justify-between">
            <DrawerTitle className="flex items-center gap-2">
              <IconLayoutGrid className="h-5 w-5" />
              Game Template
            </DrawerTitle>
            {onEditClick && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  onOpenChange(false);
                  onEditClick();
                }}
              >
                <IconEdit className="h-4 w-4" />
              </Button>
            )}
          </div>
          <DrawerDescription>
            <span className="font-semibold text-foreground">
              {template.name}
            </span>
          </DrawerDescription>
        </DrawerHeader>

        <div className="flex flex-col gap-5 overflow-y-auto px-4 pb-4">
          <div className="flex flex-wrap gap-2">
            <Badge variant={template.is_active ? "default" : "secondary"}>
              {template.is_active ? "Active" : "Inactive"}
            </Badge>
            {template.mode && (
              <Badge variant="outline">
                {MODE_LABELS[template.mode] ?? template.mode}
              </Badge>
            )}
            {template.category && (
              <Badge variant="outline">
                {CATEGORY_LABELS[template.category] ?? template.category}
              </Badge>
            )}
            {template.flag_variant && (
              <Badge variant="secondary" className="text-xs">
                <IconFlag className="h-3 w-3 mr-1" />
                {FLAG_VARIANT_LABELS[template.flag_variant] ??
                  template.flag_variant}
              </Badge>
            )}
          </div>

          <div className="grid grid-cols-4 gap-2">
            <StatCard
              label="Players"
              value={`${template.min_players}–${template.max_players}`}
            />
            <StatCard label="Questions" value={template.question_count} />
            <StatCard
              label="Pts / correct"
              value={template.points_per_correct}
            />
            <StatCard
              label="Score mode"
              value={
                SCORE_MODE_LABELS[template.score_mode] ?? template.score_mode
              }
            />
          </div>

          {template.description && (
            <div className="text-sm text-muted-foreground leading-relaxed bg-muted/30 rounded-lg p-3">
              {template.description}
            </div>
          )}

          <Separator />

          <div className="space-y-3">
            <h4 className="font-medium flex items-center text-sm">Scoring</h4>
            <div className="space-y-2 pl-6 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Mode</span>
                <span className="font-medium">
                  {SCORE_MODE_LABELS[template.score_mode] ??
                    template.score_mode}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">
                  Points per correct
                </span>
                <span className="font-medium tabular-nums">
                  {template.points_per_correct}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Time bonus</span>
                <span>
                  {template.time_bonus ? (
                    <IconCheck className="h-4 w-4" />
                  ) : (
                    <IconX className="h-4 w-4 text-muted-foreground" />
                  )}
                </span>
              </div>
              {template.xp_multiplier != null && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">XP multiplier</span>
                  <span className="font-medium tabular-nums">
                    ×{template.xp_multiplier}
                  </span>
                </div>
              )}
            </div>
          </div>

          <Separator />

          <div className="space-y-3">
            <h4 className="font-medium flex items-center text-sm">
              Questions & Dataset
            </h4>
            <div className="space-y-2 pl-6 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Count</span>
                <span className="font-medium tabular-nums">
                  {template.question_count}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Dataset</span>
                <span className="font-medium">
                  {dataset ? (
                    <Badge variant="outline" className="font-mono text-xs">
                      {dataset.name}
                    </Badge>
                  ) : (
                    <span className="text-muted-foreground">Default</span>
                  )}
                </span>
              </div>
              {template.category && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Category</span>
                  <span className="font-medium capitalize">
                    {template.category}
                  </span>
                </div>
              )}
              {template.flag_variant && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Variant</span>
                  <span className="font-medium">
                    {FLAG_VARIANT_LABELS[template.flag_variant] ??
                      template.flag_variant}
                  </span>
                </div>
              )}
              {template.question_type && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Format</span>
                  <span className="font-medium">
                    {QUESTION_TYPE_LABELS[template.question_type] ??
                      template.question_type}
                  </span>
                </div>
              )}
            </div>
          </div>

          {(template.continent || template.include_territories) && (
            <>
              <Separator />
              <div className="space-y-3">
                <h4 className="font-medium flex items-center gap-2 text-sm">
                  <IconWorld className="h-4 w-4" />
                  Geography
                </h4>
                <div className="space-y-2 pl-6 text-sm">
                  {template.continent && (
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Continent</span>
                      <span className="font-medium capitalize">
                        {template.continent}
                      </span>
                    </div>
                  )}
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">
                      Include territories
                    </span>
                    <span>
                      {template.include_territories ? (
                        <IconCheck className="h-4 w-4" />
                      ) : (
                        <IconX className="h-4 w-4 text-muted-foreground" />
                      )}
                    </span>
                  </div>
                </div>
              </div>
            </>
          )}

          <Separator />

          <div className="space-y-3">
            <h4 className="font-medium text-sm text-muted-foreground uppercase tracking-wide">
              Technical
            </h4>
            <div className="space-y-2 text-sm">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">Slug</Label>
                <div className="flex items-center gap-2">
                  <code className="flex-1 text-xs bg-muted px-2 py-1 rounded font-mono">
                    {template.slug}
                  </code>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0"
                    onClick={copySlugID}
                  >
                    {copiedSlug ? (
                      <IconCheck className="h-3 w-3" />
                    ) : (
                      <IconCopy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">ID</Label>
                <div className="flex items-center gap-2">
                  <code className="flex-1 text-xs bg-muted px-2 py-1 rounded font-mono">
                    {template.id}
                  </code>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0"
                    onClick={copyTemplateID}
                  >
                    {copiedId ? (
                      <IconCheck className="h-3 w-3" />
                    ) : (
                      <IconCopy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>Created</span>
                <span>{formatDate(template.created_at)}</span>
              </div>
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>Updated</span>
                <span>{formatDate(template.updated_at)}</span>
              </div>
            </div>
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
