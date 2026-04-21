"use client";

import { QuestionEditDialog } from "@/components/question/question-edit-dialog";
import { QuestionViewDialog } from "@/components/question/question-view-dialog";
import { QuestionsDataTable } from "@/components/question/questions-data-table";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useDatasets } from "@/hooks/useDatasets";
import { apiGet } from "@/lib/api-client";
import { QUESTIONS_ENDPOINTS } from "@/lib/api/endpoints";
import type { Question } from "@/lib/types/question.types";
import { IconDatabase, IconHelpCircle } from "@tabler/icons-react";
import * as React from "react";
import { useNavigate, useSearchParams } from "react-router";

export default function QuestionsPage() {
  const [totalQuestions, setTotalQuestions] = React.useState<number>(0);
  const [selectedDatasetId, setSelectedDatasetId] = React.useState<
    string | undefined
  >(undefined);

  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const viewId = searchParams.get("view");
  const [viewQuestion, setViewQuestion] = React.useState<Question | null>(null);
  const [editQuestion, setEditQuestion] = React.useState<Question | null>(null);
  const [editOpen, setEditOpen] = React.useState(false);

  React.useEffect(() => {
    if (!viewId) {
      setViewQuestion(null);
      return;
    }
    apiGet(QUESTIONS_ENDPOINTS.GET(viewId))
      .then((r) => (r.ok ? r.json() : null))
      .then((json) => {
        if (json) setViewQuestion(json.data ?? json);
      })
      .catch(() => {});
  }, [viewId]);

  const handleViewClose = (open: boolean) => {
    if (!open) {
      const params = new URLSearchParams(searchParams.toString());
      params.delete("view");
      const qs = params.toString();
      navigate(qs ? `/questions?${qs}` : "/questions");
      setViewQuestion(null);
    }
  };

  const handleEditQuestion = () => {
    if (!viewQuestion) return;
    setEditQuestion(viewQuestion);
    handleViewClose(false);
    setEditOpen(true);
  };

  const handleQuestionUpdated = (_updatedQuestion: Question) => {
    setEditOpen(false);
    setEditQuestion(null);
  };

  const { datasets, loading } = useDatasets();

  React.useEffect(() => {
    if (datasets.length > 0 && !selectedDatasetId) {
      const defaultDataset = datasets.find((d) => d.is_default);
      if (defaultDataset) {
        setSelectedDatasetId(defaultDataset.id);
      } else if (datasets[0]) {
        setSelectedDatasetId(datasets[0].id);
      }
    }
  }, [datasets, selectedDatasetId]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (datasets.length === 0) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <h1 className="text-3xl font-bold">Questions</h1>
          <p className="text-muted-foreground">
            Browse and search questions from your datasets
          </p>
        </div>
        <div className="px-4 lg:px-6">
          <Card className="border-0 dark:border">
            <CardContent className="flex flex-col items-center justify-center py-12">
              <IconHelpCircle className="h-16 w-16 text-muted-foreground mb-4" />
              <h3 className="text-xl font-semibold">No Question Datasets</h3>
              <p className="text-muted-foreground text-center max-w-md mt-2">
                Import a question dataset from Cultpedia to browse and manage
                questions.
              </p>
              <Button className="mt-4" onClick={() => navigate("/datasets")}>
                Go to Datasets
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Questions</h1>
            <p className="text-muted-foreground">
              Browse and search questions • {totalQuestions} questions
            </p>
          </div>
          <div className="flex items-center gap-2">
            <IconDatabase className="h-4 w-4 text-muted-foreground" />
            <Select
              value={selectedDatasetId}
              onValueChange={setSelectedDatasetId}
            >
              <SelectTrigger className="w-[300px]">
                <SelectValue placeholder="Select dataset" />
              </SelectTrigger>
              <SelectContent>
                {datasets
                  .filter((d) => d.is_active)
                  .map((dataset) => (
                    <SelectItem key={dataset.id} value={dataset.id}>
                      {dataset.name} {dataset.is_default && "(Default)"}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <QuestionsDataTable
          key={`questions-${selectedDatasetId}`}
          onTotalCountChange={setTotalQuestions}
          datasetId={selectedDatasetId}
        />
      </div>

      {viewQuestion && (
        <QuestionViewDialog
          question={viewQuestion}
          open={true}
          onOpenChange={handleViewClose}
          onEditClick={handleEditQuestion}
        />
      )}

      {editQuestion && (
        <QuestionEditDialog
          question={editQuestion}
          open={editOpen}
          onOpenChange={(open) => {
            setEditOpen(open);
            if (!open) setEditQuestion(null);
          }}
          onQuestionUpdated={handleQuestionUpdated}
        />
      )}
    </div>
  );
}
