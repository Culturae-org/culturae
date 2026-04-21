"use client";

import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
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
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { useDatasets, useGeographyDatasets } from "@/hooks/useDatasets";
import { useIsMobile } from "@/hooks/useMobile";
import {
  FIXED_PLAYER_COUNT_MODES,
  FLAG_VARIANT_OPTIONS,
  QUESTION_TYPE_OPTIONS,
  SCORE_MODE_OPTIONS,
} from "@/lib/constants/game-template.constants";
import type {
  FlagVariant,
  GameTemplate,
  ScoreMode,
  TemplateCategory,
  TemplateMode,
  UpdateGameTemplateRequest,
} from "@/lib/types/game-template.types";
import { useEffect, useState } from "react";

interface Props {
  template: GameTemplate | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onUpdated: (
    id: string,
    data: UpdateGameTemplateRequest,
  ) => Promise<GameTemplate | null>;
}

const SUPPORTED_LANGS = [
  { code: "fr", label: "French (fr)" },
  { code: "es", label: "Spanish (es)" },
];

type FormState = {
  name: string;
  description: string;
  name_i18n: Record<string, string>;
  description_i18n: Record<string, string>;
  slug: string;
  mode: TemplateMode;
  min_players: number;
  max_players: number;
  question_count: number;
  dataset_id?: string;
  points_per_correct: number;
  time_bonus: boolean;
  score_mode: ScoreMode;
  is_active: boolean;
  category: TemplateCategory;
  flag_variant: FlagVariant;
  question_type: string;
  xp_multiplier?: number;
  continent: string;
  include_territories: boolean;
};

function toForm(t: GameTemplate): FormState {
  return {
    name: t.name,
    description: t.description ?? "",
    name_i18n: t.name_i18n ?? {},
    description_i18n: t.description_i18n ?? {},
    slug: t.slug,
    mode: t.mode ?? "",
    min_players: t.min_players,
    max_players: t.max_players,
    question_count: t.question_count,
    dataset_id: t.dataset_id,
    points_per_correct: t.points_per_correct,
    time_bonus: t.time_bonus,
    score_mode: t.score_mode,
    is_active: t.is_active,
    category: t.category ?? "general",
    flag_variant: t.flag_variant ?? "",
    question_type: t.question_type ?? "",
    xp_multiplier: t.xp_multiplier,
    continent: t.continent ?? "",
    include_territories: t.include_territories ?? false,
  };
}

const isGeo = (c: TemplateCategory) => c === "flags" || c === "geography";

export function GameTemplateEditDialog({
  template,
  open,
  onOpenChange,
  onUpdated,
}: Props) {
  const isMobile = useIsMobile();
  const { datasets: questionDatasets } = useDatasets();
  const { datasets: geoDatasets } = useGeographyDatasets();
  const [form, setForm] = useState<FormState | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (template) setForm(toForm(template));
  }, [template]);

  function set<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => (prev ? { ...prev, [key]: value } : prev));
  }

  function handleModeChange(value: string) {
    set("mode", (value === "custom" ? "" : value) as TemplateMode);
  }

  function handleCategoryChange(value: string) {
    const category = value === "none" ? "" : (value as TemplateCategory);
    setForm((prev) => {
      if (!prev) return prev;
      const nextIsGeo = isGeo(category as TemplateCategory);
      const prevIsGeo = isGeo(prev.category);
      return {
        ...prev,
        category,
        ...(nextIsGeo !== prevIsGeo ? { dataset_id: undefined } : {}),
        ...(!nextIsGeo
          ? { flag_variant: "", continent: "", include_territories: false }
          : {}),
      };
    });
  }

  async function handleSubmit() {
    if (!form || !template) return;
    setLoading(true);
    try {
      const result = await onUpdated(template.id, form);
      if (result) onOpenChange(false);
    } catch (error) {
      console.error("Error updating template:", error);
    } finally {
      setLoading(false);
    }
  }

  if (!form) return null;

  const selectedModeValue =
    form.mode === "" ? "custom" : (form.mode ?? "custom");
  const selectedCategory =
    form.category === "" ? "none" : (form.category ?? "general");
  const showGeoFields = isGeo(form.category);

  return (
    <Sheet modal={false} open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side={isMobile ? "bottom" : "right"}
        className="flex flex-col p-0 sm:max-w-sm overflow-hidden"
        onOpenAutoFocus={(e) => e.preventDefault()}
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        <SheetHeader className="gap-1 px-4 pt-4">
          <SheetTitle>Edit template</SheetTitle>
          <p className="text-sm text-muted-foreground font-mono">
            {template?.slug}
          </p>
        </SheetHeader>

        <div className="flex flex-col gap-5 overflow-y-auto flex-1 px-4 pb-4">
          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Identity
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Name</Label>
              <Input
                value={form.name}
                onChange={(e) => set("name", e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Description</Label>
              <Input
                value={form.description}
                onChange={(e) => set("description", e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Slug</Label>
              <Input
                value={form.slug}
                onChange={(e) => set("slug", e.target.value)}
              />
            </div>
          </div>

          <Separator />

          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Translations
            </div>
            <p className="text-xs text-muted-foreground -mt-1">
              English (en) is always the default. Add overrides for other
              languages below.
            </p>
            {SUPPORTED_LANGS.map((lang) => (
              <div
                key={lang.code}
                className="flex flex-col gap-2 border rounded-lg p-3"
              >
                <div className="text-xs font-medium text-muted-foreground">
                  {lang.label}
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label className="text-xs">Name</Label>
                  <Input
                    value={form.name_i18n[lang.code] ?? ""}
                    onChange={(e) =>
                      set("name_i18n", {
                        ...form.name_i18n,
                        [lang.code]: e.target.value,
                      })
                    }
                    placeholder={`${form.name || "Template name"} (${lang.code})`}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label className="text-xs">Description</Label>
                  <Input
                    value={form.description_i18n[lang.code] ?? ""}
                    onChange={(e) =>
                      set("description_i18n", {
                        ...form.description_i18n,
                        [lang.code]: e.target.value,
                      })
                    }
                    placeholder={`${form.description || "Description"} (${lang.code})`}
                  />
                </div>
              </div>
            ))}
          </div>

          <Separator />

          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Engine
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Game mode</Label>
              <Select
                value={selectedModeValue}
                onValueChange={handleModeChange}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="solo">Solo</SelectItem>
                  <SelectItem value="1v1">1v1 Duel</SelectItem>
                  <SelectItem value="multi">Multiplayer</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-3 gap-3">
              <div className="flex flex-col gap-1.5">
                <Label>Min players</Label>
                <Input
                  type="number"
                  min={1}
                  value={form.min_players}
                  onChange={(e) => set("min_players", Number(e.target.value))}
                  disabled={FIXED_PLAYER_COUNT_MODES.includes(form.mode ?? "")}
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label>Max players</Label>
                <Input
                  type="number"
                  min={1}
                  value={form.max_players}
                  onChange={(e) => set("max_players", Number(e.target.value))}
                  disabled={FIXED_PLAYER_COUNT_MODES.includes(form.mode ?? "")}
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label>Questions</Label>
                <Input
                  type="number"
                  min={1}
                  value={form.question_count}
                  onChange={(e) =>
                    set("question_count", Number(e.target.value))
                  }
                />
              </div>
            </div>
          </div>

          <Separator />

          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Questions
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>
                Dataset{" "}
                <span className="text-muted-foreground text-xs">
                  (optional)
                </span>
              </Label>
              <Select
                value={form.dataset_id ?? "none"}
                onValueChange={(v) =>
                  set("dataset_id", v === "none" ? undefined : v)
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Use default dataset" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">Default (auto)</SelectItem>
                  {(showGeoFields ? geoDatasets : questionDatasets)
                    .filter((d) => d.is_active)
                    .map((d) => (
                      <SelectItem key={d.id} value={d.id}>
                        {d.name}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Category</Label>
              <Select
                value={selectedCategory}
                onValueChange={handleCategoryChange}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="general">General knowledge</SelectItem>
                  <SelectItem value="flags">Flags</SelectItem>
                  <SelectItem value="geography">Geography</SelectItem>
                  <SelectItem value="none">None / custom</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {form.category === "general" && (
              <div className="flex flex-col gap-1.5">
                <Label>
                  Question format{" "}
                  <span className="text-muted-foreground text-xs">
                    (optional)
                  </span>
                </Label>
                <Select
                  value={form.question_type || "any"}
                  onValueChange={(v) =>
                    set("question_type", v === "any" ? "" : v)
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="any">Any (mix of formats)</SelectItem>
                    {QUESTION_TYPE_OPTIONS.map((o) => (
                      <SelectItem key={o.value} value={o.value}>
                        {o.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {showGeoFields && (
              <>
                <div className="flex flex-col gap-1.5">
                  <Label>Question variant</Label>
                  <Select
                    value={form.flag_variant || "none"}
                    onValueChange={(v) =>
                      set(
                        "flag_variant",
                        v === "none" ? "" : (v as FlagVariant),
                      )
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select…" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">Mix (random)</SelectItem>
                      {FLAG_VARIANT_OPTIONS.map((o) => (
                        <SelectItem key={o.value} value={o.value}>
                          {o.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <Label>Continent</Label>
                    <Select
                      value={form.continent || "world"}
                      onValueChange={(v) =>
                        set("continent", v === "world" ? "" : v)
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="world">World</SelectItem>
                        <SelectItem value="europe">Europe</SelectItem>
                        <SelectItem value="asia">Asia</SelectItem>
                        <SelectItem value="africa">Africa</SelectItem>
                        <SelectItem value="americas">Americas</SelectItem>
                        <SelectItem value="oceania">Oceania</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    checked={form.include_territories}
                    onCheckedChange={(v) => set("include_territories", v)}
                    id="edit_include_territories"
                  />
                  <Label htmlFor="edit_include_territories">
                    Include territories
                  </Label>
                </div>
              </>
            )}
          </div>

          <Separator />

          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Scoring
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <Label>Score mode</Label>
                <Select
                  value={form.score_mode}
                  onValueChange={(v) => set("score_mode", v as ScoreMode)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {SCORE_MODE_OPTIONS.map((o) => (
                      <SelectItem key={o.value} value={o.value}>
                        {o.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex flex-col gap-1.5">
                <Label>Pts / correct</Label>
                <Input
                  type="number"
                  min={0}
                  value={form.points_per_correct}
                  onChange={(e) =>
                    set("points_per_correct", Number(e.target.value))
                  }
                />
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Switch
                checked={form.time_bonus}
                onCheckedChange={(v) => set("time_bonus", v)}
                id="edit_time_bonus"
              />
              <Label htmlFor="edit_time_bonus">Time bonus</Label>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>
                XP multiplier{" "}
                <a
                  href="/settings#xp"
                  className="text-muted-foreground text-xs"
                >
                  (optional — overrides global mode setting)
                </a>
              </Label>
              <Input
                type="number"
                min={0}
                step={0.1}
                value={form.xp_multiplier ?? ""}
                placeholder="e.g. 1.5 (leave empty = global default)"
                onChange={(e) =>
                  set(
                    "xp_multiplier",
                    e.target.value === "" ? undefined : Number(e.target.value),
                  )
                }
              />
            </div>
          </div>

          <Separator />

          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Options
            </div>
            <div className="flex items-center gap-2">
              <Switch
                checked={form.is_active}
                onCheckedChange={(v) => set("is_active", v)}
                id="edit_is_active"
              />
              <Label htmlFor="edit_is_active">Active</Label>
            </div>
          </div>
        </div>

        <SheetFooter className="flex flex-row gap-2 px-4 pb-4">
          <SheetClose asChild>
            <Button variant="outline" className="flex-1" disabled={loading}>
              Cancel
            </Button>
          </SheetClose>
          <Button className="flex-1" onClick={handleSubmit} disabled={loading}>
            {loading ? "Saving…" : "Save changes"}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
