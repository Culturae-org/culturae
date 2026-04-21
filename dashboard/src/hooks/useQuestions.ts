"use client";

import { questionsService } from "@/lib/services/questions.service";
import type {
  Question,
  QuestionCreateData,
  QuestionUpdateData,
  QuestionsQueryParams,
} from "@/lib/types/question.types";
import { useCallback, useState } from "react";
import { toast } from "sonner";

export function useQuestions(defaultDatasetId?: string) {
  const [questions, setQuestions] = useState<Question[]>([]);
  const [currentQuestion, setCurrentQuestion] = useState<Question | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);

  const fetchQuestions = useCallback(
    async (params?: QuestionsQueryParams) => {
      setLoading(true);
      setError(null);

      try {
        const queryParams = { ...params };
        if (defaultDatasetId && !queryParams.dataset_id) {
          queryParams.dataset_id = defaultDatasetId;
        }

        const result = await questionsService.getQuestions(queryParams);
        setQuestions(result.data);
        setCurrentPage(result.page);
        setTotalPages(result.total_pages);
        setTotalCount(result.total);
        return result;
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to fetch questions";
        setError(message);
        toast.error(message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [defaultDatasetId],
  );

  const fetchQuestionById = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);

    try {
      const question = await questionsService.getQuestionById(id);
      setCurrentQuestion(question);
      return question;
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch question";
      setError(message);
      toast.error(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchQuestionBySlug = useCallback(
    async (slug: string, datasetId?: string) => {
      setLoading(true);
      setError(null);

      try {
        const targetDatasetId = datasetId || defaultDatasetId;
        const question = await questionsService.getQuestionBySlug(
          slug,
          targetDatasetId,
        );
        setCurrentQuestion(question);
        return question;
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to fetch question";
        setError(message);
        toast.error(message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [defaultDatasetId],
  );

  const createQuestion = useCallback(
    async (questionData: QuestionCreateData) => {
      setLoading(true);
      setError(null);

      try {
        const payload: QuestionCreateData = {
          ...questionData,
          dataset_id: questionData.dataset_id ?? defaultDatasetId,
        };
        const newQuestion = await questionsService.createQuestion(payload);
        toast.success("Question created successfully");

        setQuestions((prev) => [newQuestion, ...prev]);
        setTotalCount((prev) => prev + 1);

        return newQuestion;
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to create question";
        setError(message);
        toast.error(message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [defaultDatasetId],
  );

  const updateQuestion = useCallback(
    async (id: string, updates: QuestionUpdateData) => {
      setLoading(true);
      setError(null);

      try {
        const updatedQuestion = await questionsService.updateQuestion(
          id,
          updates,
        );
        toast.success("Question updated successfully");

        setQuestions((prev) =>
          prev.map((q) => (q.id === id ? updatedQuestion : q)),
        );

        if (currentQuestion?.id === id) {
          setCurrentQuestion(updatedQuestion);
        }

        return updatedQuestion;
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to update question";
        setError(message);
        toast.error(message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [currentQuestion],
  );

  const deleteQuestion = useCallback(
    async (id: string) => {
      setLoading(true);
      setError(null);

      try {
        await questionsService.deleteQuestion(id);
        toast.success("Question deleted successfully");

        setQuestions((prev) => prev.filter((q) => q.id !== id));
        setTotalCount((prev) => Math.max(0, prev - 1));

        if (currentQuestion?.id === id) {
          setCurrentQuestion(null);
        }
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to delete question";
        setError(message);
        toast.error(message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [currentQuestion],
  );

  const exportQuestions = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const blob = await questionsService.exportQuestions();

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `questions-export-${new Date().toISOString()}.json`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      toast.success("Questions export downloaded");
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to export questions";
      setError(message);
      toast.error(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    questions,
    currentQuestion,
    loading,
    error,
    currentPage,
    totalPages,
    totalCount,

    fetchQuestions,
    fetchQuestionById,
    fetchQuestionBySlug,
    createQuestion,
    updateQuestion,
    deleteQuestion,
    exportQuestions,

    setQuestions,
    setCurrentQuestion,
    clearError: () => setError(null),
  };
}
