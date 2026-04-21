"use client";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
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
import { useIsMobile } from "@/hooks/useMobile";
import { useQuestions } from "@/hooks/useQuestions";
import { IconPlus, IconX } from "@tabler/icons-react";
import * as React from "react";
import { toast } from "sonner";

interface QuestionCreateDialogProps {
  datasetId?: string;
  onQuestionCreated?: () => void;
}

const getDefaultAnswers = (qtype: string) => {
  if (qtype === "true-false") {
    return [
      {
        slug: "true",
        is_correct: false,
        i18n: {
          en: { label: "True" },
          fr: { label: "Vrai" },
          es: { label: "Verdadero" },
        },
      },
      {
        slug: "false",
        is_correct: false,
        i18n: {
          en: { label: "False" },
          fr: { label: "Faux" },
          es: { label: "Falso" },
        },
      },
    ];
  }
  return [
    {
      slug: "a",
      is_correct: false,
      i18n: { en: { label: "" }, fr: { label: "" }, es: { label: "" } },
    },
    {
      slug: "b",
      is_correct: false,
      i18n: { en: { label: "" }, fr: { label: "" }, es: { label: "" } },
    },
    {
      slug: "c",
      is_correct: false,
      i18n: { en: { label: "" }, fr: { label: "" }, es: { label: "" } },
    },
    {
      slug: "d",
      is_correct: false,
      i18n: { en: { label: "" }, fr: { label: "" }, es: { label: "" } },
    },
  ];
};

export function QuestionCreateDialog({
  datasetId,
  onQuestionCreated,
}: QuestionCreateDialogProps) {
  const { createQuestion } = useQuestions(datasetId);
  const isMobile = useIsMobile();
  const [isOpen, setIsOpen] = React.useState(false);
  const [isLoading, setIsLoading] = React.useState(false);

  const [formData, setFormData] = React.useState({
    slug: "",
    version: "1.0.0",
    kind: "question",
    qtype: "single_choice",
    difficulty: "intermediate",
    estimated_seconds: 30,
    shuffle_answers: true,
    theme: { slug: "" },
    subthemes: [] as Array<{ slug: string }>,
    tags: [] as Array<{ slug: string }>,
    i18n: {
      en: { title: "", stem: "", explanation: "" },
      fr: { title: "", stem: "", explanation: "" },
      es: { title: "", stem: "", explanation: "" },
    },
    answers: getDefaultAnswers("single_choice"),
    sources: [] as string[],
  });

  const handleQtypeChange = (newQtype: string) => {
    setFormData((prev) => ({
      ...prev,
      qtype: newQtype,
      shuffle_answers: newQtype !== "true-false",
      answers: getDefaultAnswers(newQtype),
    }));
  };

  const handleSubmit = async () => {
    setIsLoading(true);

    if (!formData.slug.trim()) {
      toast.error("Slug is required");
      setIsLoading(false);
      return;
    }
    if (!formData.theme.slug.trim()) {
      toast.error("Theme is required");
      setIsLoading(false);
      return;
    }
    if (!formData.i18n.en.title.trim() || !formData.i18n.en.stem.trim()) {
      toast.error("English title and question text are required");
      setIsLoading(false);
      return;
    }
    if (!formData.i18n.fr.title.trim() || !formData.i18n.fr.stem.trim()) {
      toast.error("French title and question text are required");
      setIsLoading(false);
      return;
    }
    if (!formData.i18n.es.title.trim() || !formData.i18n.es.stem.trim()) {
      toast.error("Spanish title and question text are required");
      setIsLoading(false);
      return;
    }
    if (!formData.answers.some((a) => a.is_correct)) {
      toast.error("Exactly one answer must be marked as correct");
      setIsLoading(false);
      return;
    }
    if (formData.answers.filter((a) => a.is_correct).length > 1) {
      toast.error("Only one answer can be marked as correct");
      setIsLoading(false);
      return;
    }

    try {
      await createQuestion(formData);
      toast.success("Question created successfully");
      setIsOpen(false);
      onQuestionCreated?.();

      setFormData({
        slug: "",
        version: "1.0.0",
        kind: "question",
        qtype: "single_choice",
        difficulty: "intermediate",
        estimated_seconds: 30,
        shuffle_answers: true,
        theme: { slug: "" },
        subthemes: [],
        tags: [],
        i18n: {
          en: { title: "", stem: "", explanation: "" },
          fr: { title: "", stem: "", explanation: "" },
          es: { title: "", stem: "", explanation: "" },
        },
        answers: getDefaultAnswers("single_choice"),
        sources: [],
      });
    } catch (error) {
      toast.error("Failed to create question", {
        description:
          error instanceof Error ? error.message : "An error occurred",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Sheet modal={false} open={isOpen} onOpenChange={setIsOpen}>
      <SheetTrigger asChild>
        <Button size="sm">
          <IconPlus className="mr-2 h-4 w-4" />
          New question
        </Button>
      </SheetTrigger>
      <SheetContent
        side={isMobile ? "bottom" : "right"}
        className="flex flex-col p-0 sm:max-w-sm overflow-hidden"
        onPointerDownOutside={(e) => e.preventDefault()}
        onInteractOutside={(e) => e.preventDefault()}
      >
        <SheetHeader className="gap-1 px-4 pt-4">
          <SheetTitle>New question</SheetTitle>
        </SheetHeader>

        <div className="flex flex-col gap-5 overflow-y-auto flex-1 px-4 pb-4">
          {/* Basic Info */}
          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Identity
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>
                Slug <span className="text-destructive">*</span>
              </Label>
              <Input
                value={formData.slug}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, slug: e.target.value }))
                }
                placeholder="question-slug"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <Label>
                  Type <span className="text-destructive">*</span>
                </Label>
                <Select
                  value={formData.qtype}
                  onValueChange={handleQtypeChange}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="single_choice">Single choice</SelectItem>
                    <SelectItem value="true-false">True / False</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex flex-col gap-1.5">
                <Label>
                  Difficulty <span className="text-destructive">*</span>
                </Label>
                <Select
                  value={formData.difficulty}
                  onValueChange={(v) =>
                    setFormData((prev) => ({ ...prev, difficulty: v }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="beginner">Beginner</SelectItem>
                    <SelectItem value="intermediate">Intermediate</SelectItem>
                    <SelectItem value="advanced">Advanced</SelectItem>
                    <SelectItem value="expert">Pro</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <Label>
                  Theme <span className="text-destructive">*</span>
                </Label>
                <Input
                  value={formData.theme.slug}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      theme: { slug: e.target.value },
                    }))
                  }
                  placeholder="theme-slug"
                />
              </div>
            </div>
          </div>

          <Separator />

          {/* English */}
          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              English <span className="text-destructive">*</span>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Title</Label>
              <Input
                value={formData.i18n.en.title}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    i18n: {
                      ...prev.i18n,
                      en: { ...prev.i18n.en, title: e.target.value },
                    },
                  }))
                }
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Question text</Label>
              <Input
                value={formData.i18n.en.stem}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    i18n: {
                      ...prev.i18n,
                      en: { ...prev.i18n.en, stem: e.target.value },
                    },
                  }))
                }
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>
                Explanation{" "}
                <span className="text-muted-foreground text-xs">
                  (optional)
                </span>
              </Label>
              <Input
                value={formData.i18n.en.explanation}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    i18n: {
                      ...prev.i18n,
                      en: { ...prev.i18n.en, explanation: e.target.value },
                    },
                  }))
                }
              />
            </div>
          </div>

          <Separator />

          {/* French */}
          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Français <span className="text-destructive">*</span>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Titre</Label>
              <Input
                value={formData.i18n.fr.title}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    i18n: {
                      ...prev.i18n,
                      fr: { ...prev.i18n.fr, title: e.target.value },
                    },
                  }))
                }
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Texte de la question</Label>
              <Input
                value={formData.i18n.fr.stem}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    i18n: {
                      ...prev.i18n,
                      fr: { ...prev.i18n.fr, stem: e.target.value },
                    },
                  }))
                }
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>
                Explication{" "}
                <span className="text-muted-foreground text-xs">
                  (optionnel)
                </span>
              </Label>
              <Input
                value={formData.i18n.fr.explanation}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    i18n: {
                      ...prev.i18n,
                      fr: { ...prev.i18n.fr, explanation: e.target.value },
                    },
                  }))
                }
              />
            </div>
          </div>

          <Separator />

          {/* Spanish */}
          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Español <span className="text-destructive">*</span>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Título</Label>
              <Input
                value={formData.i18n.es.title}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    i18n: {
                      ...prev.i18n,
                      es: { ...prev.i18n.es, title: e.target.value },
                    },
                  }))
                }
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Texto de la pregunta</Label>
              <Input
                value={formData.i18n.es.stem}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    i18n: {
                      ...prev.i18n,
                      es: { ...prev.i18n.es, stem: e.target.value },
                    },
                  }))
                }
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>
                Explicación{" "}
                <span className="text-muted-foreground text-xs">
                  (opcional)
                </span>
              </Label>
              <Input
                value={formData.i18n.es.explanation}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    i18n: {
                      ...prev.i18n,
                      es: { ...prev.i18n.es, explanation: e.target.value },
                    },
                  }))
                }
              />
            </div>
          </div>

          <Separator />

          {/* Answers */}
          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              {formData.qtype === "true-false" ? "Correct answer" : "Answers"}{" "}
              <span className="text-destructive">*</span>
            </div>
            {formData.qtype === "true-false" ? (
              <div className="flex flex-col gap-2">
                {formData.answers.map((answer, index) => (
                  <div
                    key={answer.slug || index}
                    className="flex items-center gap-3 p-3 border rounded-lg"
                  >
                    <input
                      type="radio"
                      name="correct-answer-tf"
                      checked={answer.is_correct}
                      onChange={() =>
                        setFormData((prev) => ({
                          ...prev,
                          answers: prev.answers.map((a, i) => ({
                            ...a,
                            is_correct: i === index,
                          })),
                        }))
                      }
                      className="h-4 w-4"
                    />
                    <span className="text-sm font-medium">
                      {answer.slug === "true"
                        ? "True / Vrai / Verdadero"
                        : "False / Faux / Falso"}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {formData.answers.map((answer, index) => (
                  <div
                    key={answer.slug || `answer-${index}`}
                    className="p-3 border rounded-lg flex flex-col gap-2"
                  >
                    <div className="flex items-center gap-2">
                      <Checkbox
                        checked={answer.is_correct}
                        onCheckedChange={(checked) =>
                          setFormData((prev) => ({
                            ...prev,
                            answers: prev.answers.map((a, i) => ({
                              ...a,
                              is_correct: i === index ? !!checked : false,
                            })),
                          }))
                        }
                      />
                      <span className="text-sm font-medium">
                        {String.fromCharCode(65 + index)}
                        {answer.is_correct ? " — Correct" : ""}
                      </span>
                    </div>
                    <div className="grid grid-cols-3 gap-2">
                      {(["en", "fr", "es"] as const).map((lang) => (
                        <div key={lang} className="flex flex-col gap-1">
                          <Label className="text-xs text-muted-foreground">
                            {lang === "en"
                              ? "English"
                              : lang === "fr"
                                ? "Français"
                                : "Español"}
                          </Label>
                          <Input
                            value={answer.i18n[lang].label}
                            onChange={(e) =>
                              setFormData((prev) => ({
                                ...prev,
                                answers: prev.answers.map((a, i) =>
                                  i === index
                                    ? {
                                        ...a,
                                        i18n: {
                                          ...a.i18n,
                                          [lang]: { label: e.target.value },
                                        },
                                      }
                                    : a,
                                ),
                              }))
                            }
                            placeholder={
                              lang === "en"
                                ? "English"
                                : lang === "fr"
                                  ? "Français"
                                  : "Español"
                            }
                          />
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <Separator />

          {/* Sources */}
          <div className="flex flex-col gap-3">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Sources{" "}
              <span className="text-muted-foreground normal-case">
                (optional)
              </span>
            </div>
            <div className="flex flex-col gap-2">
              {formData.sources.map((source, index) => (
                <div
                  key={source || `source-${index}`}
                  className="flex items-center gap-2"
                >
                  <Input
                    value={source}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        sources: prev.sources.map((s, i) =>
                          i === index ? e.target.value : s,
                        ),
                      }))
                    }
                    placeholder="https://…"
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 shrink-0"
                    onClick={() =>
                      setFormData((prev) => ({
                        ...prev,
                        sources: prev.sources.filter((_, i) => i !== index),
                      }))
                    }
                  >
                    <IconX className="h-4 w-4" />
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="self-start"
                onClick={() =>
                  setFormData((prev) => ({
                    ...prev,
                    sources: [...prev.sources, ""],
                  }))
                }
              >
                <IconPlus className="mr-1 h-3.5 w-3.5" />
                Add source
              </Button>
            </div>
          </div>
        </div>

        <SheetFooter className="flex flex-row gap-2 px-4 pb-4">
          <SheetClose asChild>
            <Button variant="outline" className="flex-1" disabled={isLoading}>
              Cancel
            </Button>
          </SheetClose>
          <Button
            className="flex-1"
            onClick={handleSubmit}
            disabled={isLoading}
          >
            {isLoading ? "Creating…" : "Create question"}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
