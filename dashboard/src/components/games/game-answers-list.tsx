"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DATASETS_ENDPOINTS } from "@/lib/api/endpoints";
import type {
  FlagAnswerData,
  GameAnswerDetail,
  GamePlayer,
  GameQuestion,
} from "@/lib/types/games.types";
import {
  IconCircleCheck,
  IconCircleX,
  IconClock,
  IconFlag,
  IconTargetArrow,
} from "@tabler/icons-react";
import * as React from "react";
import { Link } from "react-router";

function formatTimeSpent(ms: number) {
  if (ms <= 0) return "—";
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

function FlagImg({
  code,
  className = "h-4 w-auto rounded-sm",
}: { code?: string; className?: string }) {
  if (!code || code.length < 2) return null;
  return (
    <img
      src={DATASETS_ENDPOINTS.GET_FLAG(code.toLowerCase())}
      alt={code.toUpperCase()}
      className={`inline-block shrink-0 ${className}`}
      loading="lazy"
      onError={(e) => {
        (e.target as HTMLImageElement).style.display = "none";
      }}
    />
  );
}

function normalizeQType(qt: string): string {
  return qt.replace(/_(4|2|text)$/, "");
}

const QUESTION_TYPE_LABELS: Record<string, string> = {
  mcq: "MCQ",
  text_input: "Text",
  flag_to_name: "Flag → Name",
  flag_to_text: "Flag → Name (text)",
  name_to_flag: "Name → Flag",
  flag_to_capital: "Flag → Capital",
  capital_to_flag: "Capital → Flag",
  capital_to_name: "Capital → Name",
};

function qTypeLabel(qt: string): string {
  return (
    QUESTION_TYPE_LABELS[qt] || QUESTION_TYPE_LABELS[normalizeQType(qt)] || qt
  );
}

const FLAG_SHOWN_AS_QUESTION = new Set([
  "flag_to_name",
  "flag_to_text",
  "flag_to_capital",
]);
const ANSWER_IS_FLAG = new Set(["name_to_flag", "capital_to_flag"]);
const ANSWER_HAS_COUNTRY_SLUG = new Set([
  "flag_to_name",
  "flag_to_text",
  "capital_to_name",
]);

function FlagAnswerDetails({
  answer,
  countryNames,
  flagVariant,
}: {
  answer: GameAnswerDetail;
  countryNames: Record<string, string>;
  flagVariant?: string;
}) {
  const data = answer.data as FlagAnswerData | undefined;
  const base = flagVariant
    ? normalizeQType(flagVariant)
    : normalizeQType(answer.question_type);

  const dataCountrySlug = data?.question_slug || "";
  const userSlug =
    data?.user_answer?.submitted_slug ||
    data?.match_slug ||
    answer.answer_slug ||
    "";
  const correctSlug = answer.correct_answer_slug || dataCountrySlug;
  const questionSlug = dataCountrySlug || answer.question_slug || correctSlug;

  const userLabel =
    countryNames[userSlug.toLowerCase()] ||
    answer.answer_label ||
    userSlug.toUpperCase();
  const correctLabel =
    countryNames[correctSlug.toLowerCase()] ||
    answer.correct_answer_label ||
    correctSlug.toUpperCase();
  const questionTitle =
    answer.question_title ||
    countryNames[questionSlug.toLowerCase()] ||
    questionSlug.toUpperCase();

  const showFlagAsQuestion = FLAG_SHOWN_AS_QUESTION.has(base);
  const answerIsFlag = ANSWER_IS_FLAG.has(base);
  const showFlagForAnswer = answerIsFlag || ANSWER_HAS_COUNTRY_SLUG.has(base);

  const questionPrompt =
    base === "flag_to_capital"
      ? "Capital of this country?"
      : base === "capital_to_flag" || base === "capital_to_name"
        ? `Capital: ${questionTitle}`
        : base === "name_to_flag"
          ? `Which flag for: ${questionTitle}`
          : "Which country?";

  return (
    <div className="space-y-4">
      {showFlagAsQuestion ? (
        <div className="flex items-center gap-4">
          <img
            src={DATASETS_ENDPOINTS.GET_FLAG(questionSlug.toLowerCase())}
            alt={questionTitle}
            className="h-20 w-auto rounded border border-border shrink-0"
            onError={(e) => {
              (e.target as HTMLImageElement).style.display = "none";
            }}
          />
          <div>
            <div className="text-xs text-muted-foreground mb-1">
              {questionPrompt}
            </div>
            {answer.is_correct && (
              <div className="text-sm font-semibold flex items-center gap-2">
                {showFlagForAnswer && (
                  <FlagImg
                    code={correctSlug}
                    className="h-8 w-auto rounded-sm border border-border"
                  />
                )}
                {correctLabel}
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-2">
          <IconFlag className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
          <span className="text-sm font-medium">{questionPrompt}</span>
        </div>
      )}

      <div className="grid grid-cols-2 gap-x-6 gap-y-2">
        <div>
          <div className="text-xs text-muted-foreground mb-1.5">Answered</div>
          {showFlagForAnswer ? (
            <div className="flex items-center gap-2">
              <FlagImg
                code={userSlug}
                className="h-12 w-auto rounded border border-border shrink-0"
              />
              <span className="text-sm font-medium">{userLabel}</span>
            </div>
          ) : (
            <div className="text-sm font-medium">{userLabel}</div>
          )}
        </div>
        {!answer.is_correct && (
          <div>
            <div className="text-xs text-muted-foreground mb-1.5">Correct</div>
            {showFlagForAnswer ? (
              <div className="flex items-center gap-2">
                <FlagImg
                  code={correctSlug}
                  className="h-12 w-auto rounded border border-border shrink-0"
                />
                <span className="text-sm font-medium">{correctLabel}</span>
              </div>
            ) : (
              <div className="text-sm font-medium">{correctLabel}</div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Generic Answer Details ─────────────────────────────────────────────────

function GenericAnswerDetails({ answer }: { answer: GameAnswerDetail }) {
  const submittedLabel = answer.answer_label || answer.answer_slug || "—";
  const correctLabel =
    answer.correct_answer_label || answer.correct_answer_slug || "—";

  return (
    <div className="grid grid-cols-2 gap-x-6 gap-y-2">
      <div>
        <div className="text-xs text-muted-foreground mb-0.5">Submitted</div>
        <div className="text-sm font-medium">{submittedLabel}</div>
      </div>
      <div>
        <div className="text-xs text-muted-foreground mb-0.5">Expected</div>
        <div className="text-sm font-medium text-muted-foreground">
          {correctLabel}
        </div>
      </div>
    </div>
  );
}

interface GameAnswersListProps {
  enrichedAnswers: GameAnswerDetail[];
  playerMap: Record<string, GamePlayer>;
  questions?: GameQuestion[];
  gameCategory?: string;
  flagVariant?: string;
  countryNames?: Record<string, string>;
}

export function GameAnswersList({
  enrichedAnswers,
  playerMap,
  questions = [],
  gameCategory,
  flagVariant,
  countryNames = {},
}: GameAnswersListProps) {
  const questionOrderMap = React.useMemo(() => {
    const m: Record<string, number> = {};
    for (const q of questions) {
      if (q.question_id) m[q.question_id] = q.order_number;
      m[q.id] = q.order_number;
    }
    return m;
  }, [questions]);

  const isMultiPlayer = Object.keys(playerMap).length > 1;

  const sortedAnswers = React.useMemo(() => {
    return [...enrichedAnswers].sort(
      (a, b) =>
        new Date(a.answered_at).getTime() - new Date(b.answered_at).getTime(),
    );
  }, [enrichedAnswers]);

  const groupedByQuestion = React.useMemo(() => {
    if (!isMultiPlayer) return null;
    const groups: Map<string | null, GameAnswerDetail[]> = new Map();
    for (const ans of sortedAnswers) {
      const key = ans.question_id ?? "__null__";
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key)?.push(ans);
    }
    return Array.from(groups.entries()).sort(([ka], [kb]) => {
      const oa = ka && ka !== "__null__" ? (questionOrderMap[ka] ?? 999) : 999;
      const ob = kb && kb !== "__null__" ? (questionOrderMap[kb] ?? 999) : 999;
      return oa - ob;
    });
  }, [isMultiPlayer, sortedAnswers, questionOrderMap]);

  function getAnswerDisplayType(answer: GameAnswerDetail): "flags" | "generic" {
    const base = flagVariant
      ? normalizeQType(flagVariant)
      : normalizeQType(answer.question_type);
    if (
      base === "flag_to_name" ||
      base === "flag_to_text" ||
      base === "name_to_flag" ||
      base === "flag_to_capital" ||
      base === "capital_to_flag" ||
      base === "capital_to_name"
    )
      return "flags";
    return "generic";
  }

  if (sortedAnswers.length === 0) {
    return (
      <Card className="border-0 dark:border">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <IconTargetArrow className="h-4 w-4" />
            Answer Details
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-center py-6 text-sm">
            No answers recorded
          </p>
        </CardContent>
      </Card>
    );
  }

  function renderAnswerCard(
    answer: GameAnswerDetail,
    qNumber: number,
    showPlayer = true,
  ) {
    const player = playerMap[answer.player_id];
    const playerName = player?.user?.username ?? answer.player_id.slice(0, 8);
    const displayType = getAnswerDisplayType(answer);

    return (
      <div
        key={answer.id}
        className={`p-4 border rounded-lg border-l-2 ${answer.is_correct ? "border-l-foreground/40" : "border-l-muted-foreground/20"}`}
      >
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-2.5 flex-wrap">
              <span className="text-xs font-mono text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                Q{qNumber}
              </span>
              {answer.question_type && (
                <Badge variant="outline" className="text-xs">
                  {qTypeLabel(answer.question_type)}
                </Badge>
              )}
            </div>

            {displayType === "flags" ? (
              <FlagAnswerDetails
                answer={answer}
                countryNames={countryNames}
                flagVariant={flagVariant}
              />
            ) : (
              <GenericAnswerDetails answer={answer} />
            )}

            <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground border-t pt-2">
              <span className="flex items-center gap-1">
                <IconClock className="h-3 w-3" />
                {formatTimeSpent(answer.time_spent)}
              </span>
              <span>{new Date(answer.answered_at).toLocaleTimeString()}</span>
              {showPlayer && (
                <span>
                  Player:{" "}
                  {player?.user ? (
                    <Link
                      to={`/users?view=${player.user_public_id}`}
                      className="font-medium hover:underline text-foreground"
                    >
                      {playerName}
                    </Link>
                  ) : (
                    playerName
                  )}
                </span>
              )}
            </div>
          </div>

          <div className="flex flex-col items-end gap-1 shrink-0">
            {answer.is_correct ? (
              <div className="flex items-center gap-1 text-sm font-medium text-foreground/70">
                <IconCircleCheck className="h-4 w-4" />
                Correct
              </div>
            ) : (
              <div className="flex items-center gap-1 text-sm font-medium text-muted-foreground">
                <IconCircleX className="h-4 w-4" />
                Wrong
              </div>
            )}
            <div className="text-right">
              <span
                className={`text-lg font-bold ${answer.points > 0 ? "text-foreground" : "text-muted-foreground"}`}
              >
                +{answer.points}
              </span>
              <div className="text-xs text-muted-foreground">pts</div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <IconTargetArrow className="h-4 w-4" />
          Answer Details
        </CardTitle>
      </CardHeader>
      <CardContent>
        {groupedByQuestion ? (
          <div className="space-y-6">
            {groupedByQuestion.map(([questionId, answers]) => {
              const qId = questionId === "__null__" ? null : questionId;
              const qNumber = qId ? (questionOrderMap[qId] ?? "?") : "?";
              const firstAnswer = answers[0];
              const headerBase = firstAnswer
                ? flagVariant
                  ? normalizeQType(flagVariant)
                  : normalizeQType(firstAnswer.question_type)
                : "";
              const headerData = firstAnswer?.data as
                | FlagAnswerData
                | undefined;
              const headerQSlug =
                headerData?.question_slug ||
                firstAnswer?.question_slug ||
                firstAnswer?.correct_answer_slug ||
                "";
              const showFlagInHeader = FLAG_SHOWN_AS_QUESTION.has(headerBase);

              return (
                <div key={questionId ?? "null"}>
                  <div className="flex items-center gap-2.5 mb-3">
                    <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wide shrink-0">
                      Question {qNumber}
                    </span>
                    {showFlagInHeader && headerQSlug && (
                      <img
                        src={DATASETS_ENDPOINTS.GET_FLAG(
                          headerQSlug.toLowerCase(),
                        )}
                        alt={headerQSlug}
                        className="h-10 w-auto rounded border border-border shrink-0"
                        onError={(e) => {
                          (e.target as HTMLImageElement).style.display = "none";
                        }}
                      />
                    )}
                    {firstAnswer?.question_title && (
                      <span className="text-xs text-muted-foreground truncate">
                        {firstAnswer.question_title}
                      </span>
                    )}
                  </div>
                  <div className="space-y-2 pl-2 border-l-2 border-muted">
                    {answers.map((answer) => {
                      const player = playerMap[answer.player_id];
                      const playerName =
                        player?.user?.username ?? answer.player_id.slice(0, 8);
                      const base = flagVariant
                        ? normalizeQType(flagVariant)
                        : normalizeQType(answer.question_type);
                      const rowData = answer.data as FlagAnswerData | undefined;
                      const dataCountrySlug = rowData?.question_slug || "";
                      const answerSlug =
                        rowData?.user_answer?.submitted_slug ||
                        rowData?.match_slug ||
                        answer.answer_slug ||
                        "";
                      const correctSlug =
                        answer.correct_answer_slug || dataCountrySlug;
                      const answerLabel =
                        countryNames[answerSlug.toLowerCase()] ||
                        answer.answer_label ||
                        answerSlug.toUpperCase();
                      const correctLabel =
                        countryNames[correctSlug.toLowerCase()] ||
                        answer.correct_answer_label ||
                        correctSlug.toUpperCase();
                      const showFlagAnswer =
                        ANSWER_IS_FLAG.has(base) ||
                        ANSWER_HAS_COUNTRY_SLUG.has(base);

                      return (
                        <div
                          key={answer.id}
                          className={`p-3 border rounded-lg border-l-2 ${answer.is_correct ? "border-l-foreground/40" : "border-l-muted-foreground/20"}`}
                        >
                          <div className="flex items-center justify-between gap-2 mb-2">
                            <div>
                              {player?.user ? (
                                <Link
                                  to={`/users?view=${player.user_public_id}`}
                                  className="text-sm font-medium hover:underline"
                                >
                                  {playerName}
                                </Link>
                              ) : (
                                <span className="text-sm font-medium">
                                  {playerName}
                                </span>
                              )}
                            </div>
                            <div className="flex items-center gap-2 shrink-0">
                              <span className="flex items-center gap-1 text-xs text-muted-foreground">
                                <IconClock className="h-3 w-3" />
                                {formatTimeSpent(answer.time_spent)}
                              </span>
                              {answer.is_correct ? (
                                <IconCircleCheck className="h-4 w-4 text-foreground/60" />
                              ) : (
                                <IconCircleX className="h-4 w-4 text-muted-foreground/60" />
                              )}
                              <span
                                className={`text-sm font-bold ${answer.points > 0 ? "text-foreground" : "text-muted-foreground"}`}
                              >
                                +{answer.points}
                              </span>
                            </div>
                          </div>
                          <div className="flex items-center gap-4 flex-wrap">
                            <div className="flex items-center gap-2">
                              {showFlagAnswer && answerSlug && (
                                <FlagImg
                                  code={answerSlug}
                                  className="h-8 w-auto rounded border border-border shrink-0"
                                />
                              )}
                              <span className="text-sm">
                                {answerLabel || "—"}
                              </span>
                            </div>
                            {!answer.is_correct &&
                              (correctSlug || correctLabel) && (
                                <div className="flex items-center gap-1.5 text-muted-foreground">
                                  <span className="text-xs">→ Correct:</span>
                                  {showFlagAnswer && correctSlug && (
                                    <FlagImg
                                      code={correctSlug}
                                      className="h-8 w-auto rounded border border-border shrink-0"
                                    />
                                  )}
                                  <span className="text-sm">
                                    {correctLabel}
                                  </span>
                                </div>
                              )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="space-y-3">
            {sortedAnswers.map((answer, idx) => {
              const qNumber = answer.question_id
                ? (questionOrderMap[answer.question_id] ?? idx + 1)
                : idx + 1;
              return renderAnswerCard(answer, qNumber, true);
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
