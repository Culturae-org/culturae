"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { GameAnswer, GameQuestion } from "@/lib/types/games.types";
import { IconCircleCheck } from "@tabler/icons-react";

function getQuestionStem(
  q: {
    i18n: Record<string, { stem?: string; Stem?: string }>;
    answers: { slug: string; is_correct: boolean }[];
  } | null,
): string {
  if (!q) return "";
  const i18n = q.i18n || {};
  return (
    i18n.en?.stem ||
    i18n.en?.Stem ||
    i18n.fr?.stem ||
    i18n.fr?.Stem ||
    Object.values(i18n)[0]?.stem ||
    Object.values(i18n)[0]?.Stem ||
    ""
  );
}

function getCorrectSlug(
  q: {
    i18n: Record<string, { stem?: string; Stem?: string }>;
    answers: { slug: string; is_correct: boolean }[];
  } | null,
): string {
  if (!q || !Array.isArray(q.answers)) return "";
  const correct = q.answers.find((a) => a.is_correct);
  return correct?.slug ?? "";
}

function getGeoCountryName(
  targetName: Record<string, unknown> | undefined,
  slug: string | undefined,
): string {
  if (targetName) {
    const en = targetName.en;
    if (typeof en === "string") return en;
    if (typeof en === "object" && en !== null) {
      const common = (en as Record<string, unknown>).common;
      if (typeof common === "string") return common;
    }
    const fr = targetName.fr;
    if (typeof fr === "string") return fr;
  }
  return (slug ?? "").toUpperCase();
}

function getGeoStem(gq: GameQuestion): string {
  const d = gq.data;
  if (!d) return `Q${gq.order_number}`;
  const name = getGeoCountryName(d.target_name, d.target_slug ?? d.target_iso2);
  if (gq.type === "text_input")
    return `Name this country (${d.flag ? "flag" : "capital"})`;
  return `${name}`;
}

function getGeoTargetSlug(gq: GameQuestion): string {
  return gq.data?.target_slug ?? gq.data?.target_iso2 ?? "";
}

interface GameQuestionsListProps {
  questions: GameQuestion[];
  answers: GameAnswer[];
}

export function GameQuestionsList({
  questions,
  answers,
}: GameQuestionsListProps) {
  if (questions.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle className="text-base">Questions (0)</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-center py-6 text-sm">
            No questions recorded
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">
          Questions ({questions.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {[...questions]
            .sort((a, b) => a.order_number - b.order_number)
            .map((gq) => {
              const q = gq.question;
              const isGeoQ = gq.type === "text_input" || (!q && !!gq.data);
              const stem = isGeoQ
                ? getGeoStem(gq)
                : getQuestionStem(q as never) || `Q${gq.order_number}`;
              const correctSlug = isGeoQ
                ? getGeoTargetSlug(gq)
                : getCorrectSlug(q as never);
              const answerCount = answers.filter(
                (a) =>
                  a.question_id === gq.question_id || a.question_id === gq.id,
              ).length;

              return (
                <div key={gq.id} className="p-3 border rounded-lg">
                  <div className="flex items-start gap-3">
                    <span className="text-xs font-mono text-muted-foreground bg-muted px-1.5 py-0.5 rounded shrink-0">
                      {gq.order_number}
                    </span>
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium">{stem}</div>
                      <div className="flex items-center gap-3 mt-1.5 text-xs text-muted-foreground flex-wrap">
                        <Badge variant="outline" className="text-xs capitalize">
                          {gq.type || q?.qtype || "?"}
                        </Badge>
                        {q?.difficulty && (
                          <Badge
                            variant="secondary"
                            className="text-xs capitalize"
                          >
                            {q.difficulty}
                          </Badge>
                        )}
                        {correctSlug && (
                          <span className="flex items-center gap-1">
                            <IconCircleCheck className="h-3 w-3 text-green-600" />
                            {correctSlug}
                          </span>
                        )}
                        <span>{answerCount} answer(s) submitted</span>
                      </div>
                    </div>
                    <div className="text-right shrink-0">
                      <span className="text-sm font-bold">
                        {q?.points ?? 100}
                      </span>
                      <div className="text-xs text-muted-foreground">pts</div>
                    </div>
                  </div>
                </div>
              );
            })}
        </div>
      </CardContent>
    </Card>
  );
}
