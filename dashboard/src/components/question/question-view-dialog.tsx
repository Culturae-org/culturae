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
  DrawerTrigger,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useIsMobile } from "@/hooks/useMobile";
import { useViewDialog } from "@/hooks/useViewDialog";
import {
  IconClock,
  IconEdit,
  IconExternalLink,
  IconEye,
} from "@tabler/icons-react";
import * as React from "react";
import { Link } from "react-router";

import type { Question } from "@/lib/types/question.types";

// ─── Helpers ────────────────────────────────────────────────────────────────

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function getDifficultyLabel(difficulty: string) {
  const labels: Record<string, string> = {
    beginner: "Beginner",
    easy: "Easy",
    intermediate: "Intermediate",
    medium: "Medium",
    advanced: "Advanced",
    hard: "Hard",
  };
  return labels[difficulty] || difficulty;
}

function getQuestionTypeLabel(qtype: string) {
  const labels: Record<string, string> = {
    single_choice: "Single Choice",
    "true-false": "True / False",
    multiple_choice: "Multiple Choice",
  };
  return labels[qtype] || qtype.replace(/-|_/g, " ");
}

function getLanguageLabel(lang: string) {
  const labels: Record<string, string> = {
    en: "English",
    fr: "French",
    es: "Spanish",
  };
  return labels[lang] || lang.toUpperCase();
}

// ─── Main Component ─────────────────────────────────────────────────────────

interface QuestionViewDialogProps {
  question: Question;
  children?: React.ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onEditClick?: () => void;
}

export function QuestionViewDialog({
  question,
  children,
  open: controlledOpen,
  onOpenChange,
  onEditClick,
}: QuestionViewDialogProps) {
  const isMobile = useIsMobile();
  const viewDialog = useViewDialog("view");
  const isOpen =
    controlledOpen !== undefined ? controlledOpen : viewDialog.isOpen;
  const handleIsOpenChange = (open: boolean) => {
    if (onOpenChange) {
      onOpenChange(open);
    } else if (!open) {
      viewDialog.close();
    }
  };
  const languages = React.useMemo(
    () => (question.i18n ? Object.keys(question.i18n) : []),
    [question.i18n],
  );

  const initialLanguage = React.useMemo(
    () => (languages.includes("en") ? "en" : languages[0] || ""),
    [languages],
  );

  const [activeLanguage, setActiveLanguage] =
    React.useState<string>(initialLanguage);

  React.useEffect(() => {
    setActiveLanguage((current) => {
      if (languages.length > 0 && (!current || !languages.includes(current))) {
        return languages.includes("en") ? "en" : languages[0];
      }
      return current;
    });
  }, [languages]);

  return (
    <Drawer
      direction={isMobile ? "bottom" : "right"}
      open={isOpen}
      onOpenChange={handleIsOpenChange}
    >
      {controlledOpen === undefined && (
        <DrawerTrigger asChild>
          {children ? (
            <button
              type="button"
              onClick={() => handleIsOpenChange(true)}
              className="cursor-pointer text-left"
            >
              {children}
            </button>
          ) : (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => handleIsOpenChange(true)}
            >
              <IconEye className="h-4 w-4" />
            </Button>
          )}
        </DrawerTrigger>
      )}
      <DrawerContent>
        <DrawerHeader className="gap-3">
          <div className="flex items-center justify-between">
            <div>
              <DrawerTitle className="flex items-center">Question</DrawerTitle>
              <DrawerDescription className="font-mono text-xs mt-1">
                {question.slug}
              </DrawerDescription>
            </div>
            {onEditClick && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  handleIsOpenChange(false);
                  onEditClick();
                }}
              >
                <IconEdit className="h-4 w-4" />
              </Button>
            )}
          </div>

          <div className="flex flex-wrap gap-2 pt-2">
            <Badge variant="outline">
              {getQuestionTypeLabel(question.qtype)}
            </Badge>
            <Badge variant="outline">
              {getDifficultyLabel(question.difficulty)}
            </Badge>
            <Badge variant="outline" className="flex items-center gap-1">
              <IconClock className="h-3 w-3" />
              {question.estimated_seconds}s
            </Badge>
          </div>
        </DrawerHeader>

        <div className="flex flex-col gap-6 overflow-y-auto px-4 pb-4">
          {languages.length > 0 && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h4 className="font-medium flex items-center">Question</h4>
                {languages.length > 1 && (
                  <div className="flex gap-1 bg-muted p-1 rounded-md">
                    {languages.map((lang) => (
                      <button
                        type="button"
                        key={lang}
                        onClick={() => setActiveLanguage(lang)}
                        className={`px-2 py-1 rounded text-xs font-medium transition-all ${
                          activeLanguage === lang
                            ? "bg-background shadow-sm"
                            : "text-muted-foreground hover:text-foreground"
                        }`}
                      >
                        {getLanguageLabel(lang)}
                      </button>
                    ))}
                  </div>
                )}
              </div>

              <div className="bg-muted/50 rounded-lg p-4 space-y-4 pl-6">
                <div>
                  <Label className="text-xs text-muted-foreground uppercase tracking-wide">
                    Statement
                  </Label>
                  <p className="text-sm mt-2 leading-relaxed">
                    {question.i18n[activeLanguage]?.title || "N/A"}
                  </p>
                  <p className="text-base mt-3 font-medium">
                    {question.i18n[activeLanguage]?.stem || "N/A"}
                  </p>
                </div>
                {question.i18n[activeLanguage]?.explanation && (
                  <div className="pt-3 border-t border-muted">
                    <Label className="text-xs text-muted-foreground uppercase tracking-wide">
                      Explanation
                    </Label>
                    <p className="text-sm mt-2 text-muted-foreground leading-relaxed">
                      {question.i18n[activeLanguage].explanation}
                    </p>
                  </div>
                )}
              </div>
            </div>
          )}

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Categories</h4>
            <div className="flex flex-wrap gap-2 pl-6">
              {question.theme && (
                <Badge variant="default">{question.theme.slug}</Badge>
              )}
              {question.subthemes &&
                question.subthemes.length > 0 &&
                question.subthemes.map((subtheme) => (
                  <Badge key={subtheme.slug} variant="secondary">
                    {subtheme.slug}
                  </Badge>
                ))}
              {question.tags &&
                question.tags.length > 0 &&
                question.tags.map((tag) => (
                  <Badge key={tag.slug} variant="outline">
                    {tag.slug}
                  </Badge>
                ))}
            </div>
          </div>

          <Separator />

          {question.sources && question.sources.length > 0 && (
            <div className="space-y-4">
              <h4 className="font-medium flex items-center">
                Sources ({question.sources.length})
              </h4>
              <div className="space-y-2 pl-6">
                {question.sources.map((source) => {
                  let domain = source;
                  try {
                    domain = new URL(source).hostname.replace("www.", "");
                  } catch {}
                  return (
                    <a
                      key={source}
                      href={source}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground hover:underline transition-colors"
                    >
                      <span className="truncate">{domain}</span>
                    </a>
                  );
                })}
              </div>
            </div>
          )}

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Metadata</h4>
            <div className="grid grid-cols-2 gap-4 pl-6 text-sm">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">ID</Label>
                <div className="font-mono text-xs truncate">{question.id}</div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">Version</Label>
                <div className="font-mono text-xs">{question.version}</div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">Kind</Label>
                <div className="capitalize">{question.kind}</div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Shuffle Answers
                </Label>
                <div>{question.shuffle_answers ? "Yes" : "No"}</div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">Created</Label>
                <div className="text-xs">{formatDate(question.created_at)}</div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">Updated</Label>
                <div className="text-xs">{formatDate(question.updated_at)}</div>
              </div>
            </div>
          </div>
        </div>

        <DrawerFooter className="border-t flex flex-row gap-2 pt-4">
          <DrawerClose asChild>
            <Button variant="outline" className="flex-1">
              Close
            </Button>
          </DrawerClose>
          <Button variant="default" asChild className="flex-1">
            <Link to="/questions">
              <IconExternalLink className="h-4 w-4 mr-2" />
              All Questions
            </Link>
          </Button>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
