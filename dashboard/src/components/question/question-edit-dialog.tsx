"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
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
import { IconEdit, IconPlus, IconTrash } from "@tabler/icons-react";
import * as React from "react";
import { toast } from "sonner";

import type { Question, QuestionUpdateData } from "@/lib/types/question.types";

interface QuestionEditDialogProps {
  question: Question;
  onQuestionUpdated?: (updatedQuestion: Question) => void;
  children?: React.ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function QuestionEditDialog({
  question,
  onQuestionUpdated,
  children,
  open: controlledOpen,
  onOpenChange,
}: QuestionEditDialogProps) {
  const isMobile = useIsMobile();
  const { updateQuestion } = useQuestions();
  const [internalOpen, setInternalOpen] = React.useState(false);
  const isOpen = controlledOpen ?? internalOpen;
  const setIsOpen = onOpenChange ?? setInternalOpen;
  const [isLoading, setIsLoading] = React.useState(false);

  const [formData, setFormData] = React.useState<QuestionUpdateData>({
    slug: question.slug,
    version: question.version,
    qtype: question.qtype,
    difficulty: question.difficulty,
    estimated_seconds: question.estimated_seconds,
    shuffle_answers: question.shuffle_answers,
    theme: { slug: question.theme.slug },
    subthemes: [...(question.subthemes || [])],
    tags: [...(question.tags || [])],
    i18n: { ...(question.i18n || {}) },
    answers: (question.answers || []).map((answer) => ({
      slug: answer.slug,
      is_correct: answer.is_correct,
      i18n: { ...answer.i18n },
    })),
    sources: [...(question.sources || [])],
  });

  const isTrueFalse = formData.qtype === "true-false";

  const handleInputChange = (
    field: keyof QuestionUpdateData,
    value: string | number | boolean | object | undefined,
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleI18nChange = (lang: string, field: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      i18n: {
        ...prev.i18n,
        [lang]: {
          ...prev.i18n?.[lang],
          [field]: value,
        },
      },
    }));
  };

  const handleAnswerI18nChange = (
    answerIndex: number,
    lang: string,
    value: string,
  ) => {
    setFormData((prev) => ({
      ...prev,
      answers: prev.answers?.map((answer, i) =>
        i === answerIndex
          ? {
              ...answer,
              i18n: {
                ...answer.i18n,
                [lang]: { label: value },
              },
            }
          : answer,
      ),
    }));
  };

  const addSubtheme = () => {
    setFormData((prev) => ({
      ...prev,
      subthemes: [...(prev.subthemes || []), { slug: "" }],
    }));
  };

  const removeSubtheme = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      subthemes: prev.subthemes?.filter((_, i) => i !== index),
    }));
  };

  const updateSubtheme = (index: number, slug: string) => {
    setFormData((prev) => ({
      ...prev,
      subthemes: prev.subthemes?.map((subtheme, i) =>
        i === index ? { slug } : subtheme,
      ),
    }));
  };

  const addTag = () => {
    setFormData((prev) => ({
      ...prev,
      tags: [...(prev.tags || []), { slug: "" }],
    }));
  };

  const removeTag = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      tags: prev.tags?.filter((_, i) => i !== index),
    }));
  };

  const updateTag = (index: number, slug: string) => {
    setFormData((prev) => ({
      ...prev,
      tags: prev.tags?.map((tag, i) => (i === index ? { slug } : tag)),
    }));
  };

  const addSource = () => {
    setFormData((prev) => ({
      ...prev,
      sources: [...(prev.sources || []), ""],
    }));
  };

  const removeSource = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      sources: prev.sources?.filter((_, i) => i !== index),
    }));
  };

  const updateSource = (index: number, url: string) => {
    setFormData((prev) => ({
      ...prev,
      sources: prev.sources?.map((source, i) => (i === index ? url : source)),
    }));
  };

  const validateForm = (): string | null => {
    if (!formData.slug?.trim()) return "Slug is required";
    if (!formData.version?.trim()) return "Version is required";
    if (!formData.difficulty) return "Difficulty is required";
    if (!formData.theme?.slug?.trim()) return "Theme is required";
    if (!formData.i18n) return "Question content is required";

    const hasValidI18n = Object.values(formData.i18n).some(
      (lang) => lang.title?.trim() && lang.stem?.trim(),
    );
    if (!hasValidI18n)
      return "At least one language must have title and question text";

    if (!formData.answers || formData.answers.length < 2)
      return "At least 2 answers are required";
    if (!formData.answers.some((a) => a.is_correct))
      return "Exactly one answer must be marked as correct";
    if (formData.answers.filter((a) => a.is_correct).length > 1)
      return "Only one answer can be marked as correct";

    return null;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    const validationError = validateForm();
    if (validationError) {
      toast.error(validationError);
      setIsLoading(false);
      return;
    }

    try {
      const updatedQuestion = await updateQuestion(question.id, formData);

      setIsOpen(false);

      if (onQuestionUpdated) {
        onQuestionUpdated(updatedQuestion);
      }
    } catch (error) {
      console.error("Error updating question:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const languages = ["fr", "en", "es"];

  return (
    <Sheet modal={false} open={isOpen} onOpenChange={setIsOpen}>
      {controlledOpen === undefined && (
        <SheetTrigger asChild>
          {children ? (
            <button
              type="button"
              onClick={() => setIsOpen(true)}
              className="cursor-pointer text-left"
            >
              {children}
            </button>
          ) : (
            <Button variant="ghost" size="sm" onClick={() => setIsOpen(true)}>
              <IconEdit className="h-4 w-4 mr-2" />
              Edit
            </Button>
          )}
        </SheetTrigger>
      )}
      <SheetContent
        side={isMobile ? "bottom" : "right"}
        className="flex flex-col p-0 sm:max-w-4xl overflow-hidden"
        onPointerDownOutside={(e) => e.preventDefault()}
        onInteractOutside={(e) => e.preventDefault()}
      >
        <SheetHeader className="gap-1 px-4 pt-4">
          <SheetTitle>Edit Question</SheetTitle>
          <SheetDescription>
            Modify question information for <strong>{question.slug}</strong>
          </SheetDescription>
        </SheetHeader>

        <form
          onSubmit={handleSubmit}
          className="flex flex-col gap-6 overflow-y-auto flex-1 px-4 pb-4"
        >
          <div className="space-y-4">
            <h4 className="font-medium">Basic Information</h4>
            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-2">
                <Label htmlFor="slug">Slug</Label>
                <Input
                  id="slug"
                  value={formData.slug || ""}
                  onChange={(e) => handleInputChange("slug", e.target.value)}
                  placeholder="question-slug"
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="version">Version</Label>
                <Input
                  id="version"
                  value={formData.version || ""}
                  onChange={(e) => handleInputChange("version", e.target.value)}
                  placeholder="1.0"
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="difficulty">Difficulty</Label>
                <Select
                  value={formData.difficulty || ""}
                  onValueChange={(value) =>
                    handleInputChange("difficulty", value)
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select difficulty" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="beginner">Beginner</SelectItem>
                    <SelectItem value="intermediate">Intermediate</SelectItem>
                    <SelectItem value="advanced">Advanced</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="estimated_seconds">
                  Time Estimate (seconds)
                </Label>
                <Input
                  id="estimated_seconds"
                  type="number"
                  min="1"
                  value={formData.estimated_seconds || 0}
                  onChange={(e) =>
                    handleInputChange(
                      "estimated_seconds",
                      Number.parseInt(e.target.value, 10),
                    )
                  }
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="theme">Theme</Label>
                <Input
                  id="theme"
                  value={formData.theme?.slug || ""}
                  onChange={(e) =>
                    handleInputChange("theme", { slug: e.target.value })
                  }
                  placeholder="geography"
                />
              </div>
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h4 className="font-medium">Subthemes</h4>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addSubtheme}
              >
                <IconPlus className="h-4 w-4 mr-2" />
                Add Subtheme
              </Button>
            </div>
            <div className="space-y-2">
              {formData.subthemes?.map((subtheme, index) => (
                <div
                  key={subtheme.slug || `subtheme-${index}`}
                  className="flex items-center gap-2"
                >
                  <Input
                    value={subtheme.slug}
                    onChange={(e) => updateSubtheme(index, e.target.value)}
                    placeholder="subtheme-slug"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => removeSubtheme(index)}
                  >
                    <IconTrash className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h4 className="font-medium">Tags</h4>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addTag}
              >
                <IconPlus className="h-4 w-4 mr-2" />
                Add Tag
              </Button>
            </div>
            <div className="space-y-2">
              {formData.tags?.map((tag, index) => (
                <div
                  key={tag.slug || `tag-${index}`}
                  className="flex items-center gap-2"
                >
                  <Input
                    value={tag.slug}
                    onChange={(e) => updateTag(index, e.target.value)}
                    placeholder="tag-slug"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => removeTag(index)}
                  >
                    <IconTrash className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium">Question Content</h4>
            {languages.map((lang) => (
              <div key={lang} className="space-y-3 p-4 border rounded-lg">
                <h5 className="font-medium uppercase text-sm">
                  {lang} Version
                </h5>
                <div className="space-y-3">
                  <div>
                    <Label>Title</Label>
                    <Input
                      value={formData.i18n?.[lang]?.title || ""}
                      onChange={(e) =>
                        handleI18nChange(lang, "title", e.target.value)
                      }
                      placeholder={`Title in ${lang.toUpperCase()}`}
                    />
                  </div>
                  <div>
                    <Label>Question</Label>
                    <Input
                      value={formData.i18n?.[lang]?.stem || ""}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                        handleI18nChange(lang, "stem", e.target.value)
                      }
                      placeholder={`Question text in ${lang.toUpperCase()}`}
                    />
                  </div>
                  <div>
                    <Label>Explanation</Label>
                    <Input
                      value={formData.i18n?.[lang]?.explanation || ""}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                        handleI18nChange(lang, "explanation", e.target.value)
                      }
                      placeholder={`Explanation in ${lang.toUpperCase()}`}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium">
              {isTrueFalse ? "Select the correct answer" : "Answers"}
            </h4>
            <div className="space-y-3">
              {isTrueFalse
                ? formData.answers?.map((answer, index) => (
                    <div
                      key={answer.slug || `tf-${index}`}
                      className="flex items-center gap-3 p-4 border rounded-lg"
                    >
                      <input
                        type="radio"
                        name="correct-answer-tf"
                        checked={answer.is_correct}
                        onChange={() => {
                          setFormData((prev) => ({
                            ...prev,
                            answers: prev.answers?.map((a, i) => ({
                              ...a,
                              is_correct: i === index,
                            })),
                          }));
                        }}
                        className="h-4 w-4"
                      />
                      <div className="flex-1">
                        <span className="font-medium">
                          {answer.slug === "true" ? "True" : "False"}
                        </span>
                        <span className="text-muted-foreground ml-2">
                          / {answer.slug === "true" ? "Vrai" : "Faux"}
                          {" / "}
                          {answer.slug === "true" ? "Verdadero" : "Falso"}
                        </span>
                      </div>
                      <Badge
                        variant={answer.is_correct ? "default" : "outline"}
                      >
                        {answer.is_correct ? "Correct" : "Incorrect"}
                      </Badge>
                    </div>
                  ))
                : formData.answers?.map((answer, index) => (
                    <div
                      key={answer.slug || `answer-${index}`}
                      className="p-4 border rounded-lg space-y-3"
                    >
                      <div className="flex items-center gap-2">
                        <input
                          type="radio"
                          name="correct-answer"
                          checked={answer.is_correct}
                          onChange={() => {
                            setFormData((prev) => ({
                              ...prev,
                              answers: prev.answers?.map((a, i) => ({
                                ...a,
                                is_correct: i === index,
                              })),
                            }));
                          }}
                        />
                        <Label className="font-medium">Correct Answer</Label>
                        <Badge variant="outline" className="ml-auto">
                          Answer {index + 1}
                        </Badge>
                      </div>
                      {languages.map((lang) => (
                        <div key={lang}>
                          <Label className="text-sm">
                            {lang.toUpperCase()} Label
                          </Label>
                          <Input
                            value={answer.i18n[lang]?.label || ""}
                            onChange={(e) =>
                              handleAnswerI18nChange(
                                index,
                                lang,
                                e.target.value,
                              )
                            }
                            placeholder={`Answer text in ${lang.toUpperCase()}`}
                          />
                        </div>
                      ))}
                    </div>
                  ))}
            </div>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h4 className="font-medium">Sources</h4>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addSource}
              >
                <IconPlus className="h-4 w-4 mr-2" />
                Add Source
              </Button>
            </div>
            <div className="space-y-2">
              {formData.sources?.map((source, index) => (
                <div
                  key={source || `source-${index}`}
                  className="flex items-center gap-2"
                >
                  <Input
                    value={source}
                    onChange={(e) => updateSource(index, e.target.value)}
                    placeholder="https://example.com"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => removeSource(index)}
                  >
                    <IconTrash className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          </div>
        </form>

        <SheetFooter className="px-4 pb-4">
          <div className="flex gap-2">
            <Button
              onClick={handleSubmit}
              disabled={isLoading}
              className="flex-1"
            >
              {isLoading ? "Updating..." : "Update Question"}
            </Button>
          </div>
          <SheetClose asChild>
            <Button variant="outline" disabled={isLoading}>
              Cancel
            </Button>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
