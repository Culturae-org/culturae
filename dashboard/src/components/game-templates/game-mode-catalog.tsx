"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useGameTemplates } from "@/hooks/useGameTemplates";
import {
  FLAG_VARIANT_OPTIONS,
  QUESTION_TYPE_OPTIONS,
} from "@/lib/constants/game-template.constants";
import type { CreateGameTemplateRequest } from "@/lib/types/game-template.types";
import {
  IconBook,
  IconFlag,
  IconGlobe,
  IconPlus,
  IconUser,
  IconUsers,
} from "@tabler/icons-react";
import { useState } from "react";
import { GameTemplateCreateDialog } from "./game-template-create-dialog";

interface ModeCard {
  id: string;
  title: string;
  description: string;
  engine: string;
  tags: string[];
  icon: React.ReactNode;
  preset: Partial<CreateGameTemplateRequest>;
}

const MODE_CATALOG: ModeCard[] = [
  // ─── SOLO ─────────────────────────────────────────────────────────────────
  {
    id: "solo-general",
    title: "Solo — General Knowledge",
    description:
      "Single-player quiz with general knowledge questions from a dataset.",
    engine: "VersusGame",
    tags: ["solo", "1 player", "time bonus"],
    icon: <IconUser className="h-5 w-5" />,
    preset: {
      mode: "solo",
      min_players: 1,
      max_players: 1,
      score_mode: "time_bonus",
      time_bonus: true,
      category: "general",
      question_count: 10,
      points_per_correct: 100,
    },
  },
  {
    id: "solo-flags",
    title: "Solo — Flags Quiz",
    description: "Single-player quiz: identify flags, capitals, and countries.",
    engine: "VersusGame",
    tags: ["solo", "1 player", "geography", "flags"],
    icon: <IconFlag className="h-5 w-5" />,
    preset: {
      mode: "solo",
      min_players: 1,
      max_players: 1,
      score_mode: "time_bonus",
      time_bonus: true,
      category: "flags",
      flag_variant: "mix",
      question_count: 10,
      points_per_correct: 100,
    },
  },
  {
    id: "solo-geography",
    title: "Solo — Geography",
    description:
      "Single-player geography quiz with map and location questions.",
    engine: "VersusGame",
    tags: ["solo", "1 player", "geography"],
    icon: <IconGlobe className="h-5 w-5" />,
    preset: {
      mode: "solo",
      min_players: 1,
      max_players: 1,
      score_mode: "time_bonus",
      time_bonus: true,
      category: "geography",
      flag_variant: "mix",
      question_count: 10,
      points_per_correct: 100,
    },
  },
  // ─── 1v1 ──────────────────────────────────────────────────────────────────
  {
    id: "1v1-general",
    title: "1v1 Duel — General Knowledge",
    description:
      "Head-to-head quiz: the fastest correct answer wins each question.",
    engine: "VersusGame",
    tags: ["1v1", "2 players", "fastest wins"],
    icon: <IconUsers className="h-5 w-5" />,
    preset: {
      mode: "1v1",
      min_players: 2,
      max_players: 2,
      score_mode: "fastest_wins",
      time_bonus: false,
      category: "general",
      question_count: 10,
      points_per_correct: 100,
    },
  },
  {
    id: "1v1-flags",
    title: "1v1 Duel — Flags",
    description:
      "Head-to-head flag quiz: fastest player to identify the country wins.",
    engine: "VersusGame",
    tags: ["1v1", "2 players", "fastest wins", "flags"],
    icon: <IconFlag className="h-5 w-5" />,
    preset: {
      mode: "1v1",
      min_players: 2,
      max_players: 2,
      score_mode: "fastest_wins",
      time_bonus: false,
      category: "flags",
      flag_variant: "mix",
      question_count: 10,
      points_per_correct: 100,
    },
  },
  {
    id: "1v1-geography",
    title: "1v1 Duel — Geography",
    description: "Head-to-head geography duel.",
    engine: "VersusGame",
    tags: ["1v1", "2 players", "fastest wins", "geography"],
    icon: <IconGlobe className="h-5 w-5" />,
    preset: {
      mode: "1v1",
      min_players: 2,
      max_players: 2,
      score_mode: "fastest_wins",
      time_bonus: false,
      category: "geography",
      flag_variant: "mix",
      question_count: 10,
      points_per_correct: 100,
    },
  },
  // ─── QUIZ LIBRE ───────────────────────────────────────────────────────────
  {
    id: "custom",
    title: "Custom",
    description:
      "Create a template with fully custom parameters: any player count, any scoring, any dataset.",
    engine: "VersusGame",
    tags: ["custom"],
    icon: <IconBook className="h-5 w-5" />,
    preset: { mode: "" },
  },
];

interface Props {
  onTemplateCreated?: () => void;
}

export function GameModeCatalog({ onTemplateCreated }: Props) {
  const { createTemplate } = useGameTemplates();
  const [createOpen, setCreateOpen] = useState(false);
  const [activePreset, setActivePreset] = useState<
    Partial<CreateGameTemplateRequest>
  >({});

  function handleNewTemplate(preset: Partial<CreateGameTemplateRequest>) {
    setActivePreset(preset);
    setCreateOpen(true);
  }

  async function handleCreated(data: CreateGameTemplateRequest) {
    const result = await createTemplate(data);
    if (result) onTemplateCreated?.();
    return result;
  }

  return (
    <div className="px-4 lg:px-6 space-y-4">
      <div>
        <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
          Available game modes
        </h2>
        <p className="text-xs text-muted-foreground mt-0.5">
          All modes supported by the server. Create a template for any mode to
          configure it and expose it to players.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-3">
        {MODE_CATALOG.map((card) => (
          <Card key={card.id} className="flex flex-col">
            <CardHeader className="pb-2">
              <div className="flex items-start justify-between gap-2">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 rounded-md bg-muted text-muted-foreground">
                    {card.icon}
                  </div>
                  <div>
                    <CardTitle className="text-sm leading-snug">
                      {card.title}
                    </CardTitle>
                    <span className="text-xs text-muted-foreground font-mono">
                      {card.engine}
                    </span>
                  </div>
                </div>
              </div>
            </CardHeader>
            <CardContent className="flex flex-col flex-1 gap-3 pt-0">
              <CardDescription className="text-xs leading-relaxed">
                {card.description}
              </CardDescription>
              <div className="flex flex-wrap gap-1">
                {card.tags.map((tag) => (
                  <Badge
                    key={tag}
                    variant="secondary"
                    className="text-xs px-1.5 py-0"
                  >
                    {tag}
                  </Badge>
                ))}
              </div>
              {(card.id === "solo-flags" ||
                card.id === "1v1-flags" ||
                card.id === "solo-geography" ||
                card.id === "1v1-geography") && (
                <div className="text-xs text-muted-foreground">
                  <span className="font-medium">Variants: </span>
                  {FLAG_VARIANT_OPTIONS.slice(0, 4)
                    .map((v) => v.label)
                    .join(", ")}{" "}
                  +{FLAG_VARIANT_OPTIONS.length - 4} more
                </div>
              )}
              {(card.id === "solo-general" || card.id === "1v1-general") && (
                <div className="text-xs text-muted-foreground">
                  <span className="font-medium">Formats: </span>
                  Any (mix),{" "}
                  {QUESTION_TYPE_OPTIONS.map((v) => v.label).join(", ")}
                </div>
              )}
              <Button
                variant="outline"
                size="sm"
                className="mt-auto w-full"
                onClick={() => handleNewTemplate(card.preset)}
              >
                <IconPlus className="mr-1 h-3.5 w-3.5" />
                New template
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>

      <GameTemplateCreateDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreated={handleCreated}
        initialValues={activePreset}
      />
    </div>
  );
}
