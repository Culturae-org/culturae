"use client";

import { Button } from "@/components/ui/button";
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
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Switch } from "@/components/ui/switch";
import { useDatasets, useGeographyDatasets } from "@/hooks/useDatasets";
import { useIsMobile } from "@/hooks/useMobile";
import {
  FIXED_PLAYER_COUNT_MODES,
  FLAG_VARIANT_OPTIONS,
  MODE_DEFAULTS,
  QUESTION_TYPE_OPTIONS,
  SCORE_MODE_OPTIONS,
} from "@/lib/constants/game-template.constants";
import type {
  CreateGameTemplateRequest,
  FlagVariant,
  GameTemplate,
  TemplateCategory,
  TemplateMode,
} from "@/lib/types/game-template.types";
import { useEffect, useMemo, useState } from "react";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: (data: CreateGameTemplateRequest) => Promise<GameTemplate | null>;
  initialValues?: Partial<CreateGameTemplateRequest>;
}

const SUPPORTED_LANGS = [
  { code: "fr", label: "French (fr)" },
  { code: "es", label: "Spanish (es)" },
];

const EMPTY_FORM: CreateGameTemplateRequest = {
  name: "",
  description: "",
  name_i18n: {},
  description_i18n: {},
  slug: "",
  mode: "solo",
  min_players: 1,
  max_players: 1,
  question_count: 10,
  dataset_id: undefined,
  points_per_correct: 100,
  time_bonus: true,
  score_mode: "time_bonus",
  is_active: true,
  category: "general",
  flag_variant: "",
  continent: "",
  include_territories: false,
};

function slugify(value: string): string {
  return value
    .toLowerCase()
    .replace(/\s+/g, "-")
    .replace(/[^a-z0-9-]/g, "");
}

const isGeo = (c: TemplateCategory | undefined) =>
  c === "flags" || c === "geography";

export function GameTemplateCreateDialog({
  open,
  onOpenChange,
  onCreated,
  initialValues,
}: Props) {
  const isMobile = useIsMobile();
  const { datasets: questionDatasets } = useDatasets();
  const { datasets: geoDatasets } = useGeographyDatasets();
  const [form, setForm] = useState<CreateGameTemplateRequest>({
    ...EMPTY_FORM,
    ...initialValues,
  });
  const [loading, setLoading] = useState(false);

  const stableInitialValues = useMemo(
    () => initialValues ?? {},
    [initialValues],
  );

  useEffect(() => {
    if (open) setForm({ ...EMPTY_FORM, ...stableInitialValues });
  }, [open, stableInitialValues]);

  function set<K extends keyof CreateGameTemplateRequest>(
    key: K,
    value: CreateGameTemplateRequest[K],
  ) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  function handleModeChange(value: string) {
    const mode = value === "custom" ? "" : value;
    const defaults = MODE_DEFAULTS[value] ?? {};
    setForm((prev) => ({ ...prev, mode: mode as TemplateMode, ...defaults }));
  }

  function handleCategoryChange(value: string) {
    const category = value === "none" ? "" : (value as TemplateCategory);
    const nextIsGeo = isGeo(category as TemplateCategory);
    const prevIsGeo = isGeo(form.category);
    setForm((prev) => ({
      ...prev,
      category,
      ...(nextIsGeo !== prevIsGeo ? { dataset_id: undefined } : {}),
      ...(!nextIsGeo
        ? { flag_variant: "", continent: "", include_territories: false }
        : {}),
    }));
  }

  function handleNameChange(value: string) {
    setForm((prev) => ({
      ...prev,
      name: value,
      slug:
        prev.slug === "" || prev.slug === slugify(prev.name)
          ? slugify(value)
          : prev.slug,
    }));
  }

  async function handleSubmit() {
    if (!form.name.trim() || !form.slug.trim()) return;
    setLoading(true);
    const result = await onCreated(form);
    setLoading(false);
    if (result) {
      setForm(EMPTY_FORM);
      onOpenChange(false);
    }
  }

  const selectedModeValue = form.mode === "" ? "custom" : (form.mode ?? "solo");
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
          <SheetTitle>New game template</SheetTitle>
        </SheetHeader>

        <div className="flex flex-col gap-5 overflow-y-auto flex-1 px-4 pb-4">
          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Identity
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>
                Name <span className="text-destructive">*</span>
              </Label>
              <Input
                value={form.name}
                onChange={(e) => handleNameChange(e.target.value)}
                placeholder="Quick Duel"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Description</Label>
              <Input
                value={form.description ?? ""}
                onChange={(e) => set("description", e.target.value)}
                placeholder="A fast 1v1 game…"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>
                Slug <span className="text-destructive">*</span>
              </Label>
              <Input
                value={form.slug}
                onChange={(e) => set("slug", slugify(e.target.value))}
                placeholder="quick-duel"
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
                    value={form.name_i18n?.[lang.code] ?? ""}
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
                    value={form.description_i18n?.[lang.code] ?? ""}
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
                    checked={form.include_territories ?? false}
                    onCheckedChange={(v) => set("include_territories", v)}
                    id="create_include_territories"
                  />
                  <Label htmlFor="create_include_territories">
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
                  onValueChange={(v) =>
                    set(
                      "score_mode",
                      v as "classic" | "time_bonus" | "fastest_wins",
                    )
                  }
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
                checked={form.time_bonus ?? false}
                onCheckedChange={(v) => set("time_bonus", v)}
                id="create_time_bonus"
              />
              <Label htmlFor="create_time_bonus">Time bonus</Label>
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
                checked={form.is_active ?? true}
                onCheckedChange={(v) => set("is_active", v)}
                id="create_is_active"
              />
              <Label htmlFor="create_is_active">Active</Label>
            </div>
          </div>
        </div>

        <SheetFooter className="flex flex-row gap-2 px-4 pb-4">
          <SheetClose asChild>
            <Button variant="outline" className="flex-1" disabled={loading}>
              Cancel
            </Button>
          </SheetClose>
          <Button
            className="flex-1"
            onClick={handleSubmit}
            disabled={loading || !form.name.trim() || !form.slug.trim()}
          >
            {loading ? "Creating…" : "Create template"}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
