"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";
import { useIsMobile } from "@/hooks/useMobile";
import { apiPatch } from "@/lib/api-client";
import { REPORTS_ENDPOINTS } from "@/lib/api/endpoints";
import type { Report, ReportStatus } from "@/lib/types/reports.types";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  IconCheck,
  IconCopy,
  IconExternalLink,
  IconLoader2,
} from "@tabler/icons-react";
import { format } from "date-fns";
import * as React from "react";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import * as z from "zod";

const formSchema = z.object({
  status: z.enum(["pending", "in_progress", "resolved"]),
  resolution_notes: z.string().optional(),
});

interface ReportViewDrawerProps {
  report: Report | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

const STATUS_COLORS: Record<ReportStatus, string> = {
  pending:
    "text-yellow-600 bg-yellow-100 dark:text-yellow-400 dark:bg-yellow-900/40",
  in_progress:
    "text-blue-600 bg-blue-100 dark:text-blue-400 dark:bg-blue-900/40",
  resolved:
    "text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-900/40",
};

const REASON_LABELS: Record<string, string> = {
  wrong_answer: "Wrong answer",
  typo: "Typo / Spelling error",
  offensive: "Inappropriate content",
  other: "Other",
};

export function ReportViewDrawer({
  report,
  open,
  onOpenChange,
  onSuccess,
}: ReportViewDrawerProps) {
  const isMobile = useIsMobile();
  const navigate = useNavigate();
  const [loading, setLoading] = React.useState(false);
  const [copiedField, setCopiedField] = React.useState<string | null>(null);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      status: report?.status ?? "pending",
      resolution_notes: report?.resolution_notes ?? "",
    },
  });

  React.useEffect(() => {
    if (report) {
      form.reset({
        status: report.status,
        resolution_notes: report.resolution_notes ?? "",
      });
    }
  }, [report, form]);

  const copyToClipboard = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch {}
  };

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    if (!report) return;
    setLoading(true);
    try {
      const response = await apiPatch(
        REPORTS_ENDPOINTS.UPDATE_STATUS(report.id),
        values,
      );
      if (!response.ok) throw new Error();
      toast.success("Report updated successfully");
      onSuccess();
      onOpenChange(false);
    } catch {
      toast.error("Failed to update report");
    } finally {
      setLoading(false);
    }
  };

  if (!report) return null;

  const questionRef = report.question_id || report.game_question_id;
  const isGameQuestion = !report.question_id && !!report.game_question_id;

  return (
    <Sheet modal={false} open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side={isMobile ? "bottom" : "right"}
        className="flex flex-col p-0 sm:max-w-sm overflow-hidden"
        onPointerDownOutside={(e) => e.preventDefault()}
        onInteractOutside={(e) => e.preventDefault()}
      >
        <SheetHeader className="gap-1 px-4 pt-4">
          <SheetTitle className="flex items-center">
            Report Details
          </SheetTitle>
          <SheetDescription>
            Submitted{" "}
            {format(new Date(report.created_at), "MMM d, yyyy 'at' HH:mm")}
          </SheetDescription>
        </SheetHeader>

        <div className="flex flex-col gap-6 overflow-y-auto flex-1 px-4 pb-4">
          <div
            className={`flex items-center justify-between rounded-lg px-4 py-3 ${STATUS_COLORS[report.status]}`}
          >
            <span className="text-sm font-semibold capitalize">
              {report.status.replace("_", " ")}
            </span>
            <Badge
              variant="outline"
              className="capitalize border-current text-current"
            >
              {REASON_LABELS[report.reason] ?? report.reason.replace(/_/g, " ")}
            </Badge>
          </div>

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Report Content</h4>
            <div className="space-y-3 pl-6">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">Reason</Label>
                <div className="text-sm font-medium">
                  {REASON_LABELS[report.reason] ??
                    report.reason.replace(/_/g, " ")}
                </div>
              </div>
              {report.message && (
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Message
                  </Label>
                  <div className="text-sm bg-muted/50 rounded-md px-3 py-2 italic">
                    "{report.message}"
                  </div>
                </div>
              )}
              {report.resolution_notes && (
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Resolution Notes
                  </Label>
                  <div className="text-sm bg-muted/50 rounded-md px-3 py-2">
                    {report.resolution_notes}
                  </div>
                </div>
              )}
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Reporter</h4>
            <div className="space-y-3 pl-6">
              {report.user?.username && (
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Username
                  </Label>
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">
                      {report.user.username}
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 px-2 text-xs gap-1"
                      onClick={() => {
                        onOpenChange(false);
                        navigate(`/users?view=${report.user_id}`);
                      }}
                    >
                      <IconExternalLink className="h-3 w-3" />
                      View
                    </Button>
                  </div>
                </div>
              )}
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">User ID</Label>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-mono bg-muted px-2 py-1 rounded flex-1 truncate">
                    {report.user_id}
                  </span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0 shrink-0"
                    onClick={() => copyToClipboard(report.user_id, "userId")}
                  >
                    {copiedField === "userId" ? (
                      <IconCheck className="h-3 w-3 text-green-500" />
                    ) : (
                      <IconCopy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Question</h4>
            <div className="space-y-3 pl-6">
              {report.question?.question?.en && (
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Question text
                  </Label>
                  <div className="text-sm bg-muted/50 rounded-md px-3 py-2">
                    {report.question.question.en}
                  </div>
                </div>
              )}
              {report.question?.category && (
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Category
                  </Label>
                  <div className="text-sm capitalize">
                    {report.question.category}
                  </div>
                </div>
              )}
              {questionRef && (
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    {isGameQuestion ? "Game Question ID" : "Question ID"}
                  </Label>
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-mono bg-muted px-2 py-1 rounded flex-1 truncate">
                      {questionRef}
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0 shrink-0"
                      onClick={() => copyToClipboard(questionRef, "questionId")}
                    >
                      {copiedField === "questionId" ? (
                        <IconCheck className="h-3 w-3 text-green-500" />
                      ) : (
                        <IconCopy className="h-3 w-3" />
                      )}
                    </Button>
                  </div>
                </div>
              )}
              {report.game_question && (
                <div className="flex flex-col gap-1">
                  <Label className="text-xs text-muted-foreground">
                    Question #{report.game_question.order_number} in game
                  </Label>
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full justify-start gap-2"
                    onClick={() => {
                      onOpenChange(false);
                      navigate(`/games/${report.game_question?.game_id}`);
                    }}
                  >
                    <IconExternalLink className="h-3.5 w-3.5" />
                    View Game
                    <span className="font-mono text-xs text-muted-foreground ml-auto">
                      {report.game_question.game_id.slice(0, 8)}…
                    </span>
                  </Button>
                </div>
              )}
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Timeline</h4>
            <div className="space-y-3 pl-6">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Report ID
                </Label>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-mono bg-muted px-2 py-1 rounded flex-1 truncate">
                    {report.id}
                  </span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0 shrink-0"
                    onClick={() => copyToClipboard(report.id, "reportId")}
                  >
                    {copiedField === "reportId" ? (
                      <IconCheck className="h-3 w-3 text-green-500" />
                    ) : (
                      <IconCopy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">Created</Label>
                <div className="text-sm">
                  {format(new Date(report.created_at), "PPP 'at' HH:mm")}
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-muted-foreground">
                  Last Updated
                </Label>
                <div className="text-sm">
                  {format(new Date(report.updated_at), "PPP 'at' HH:mm")}
                </div>
              </div>
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h4 className="font-medium flex items-center">Update Status</h4>
            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(onSubmit)}
                className="space-y-4 pl-6"
              >
                <FormField
                  control={form.control}
                  name="status"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Status</FormLabel>
                      <Select
                        onValueChange={field.onChange}
                        value={field.value}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="Select a status" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="pending">Pending</SelectItem>
                          <SelectItem value="in_progress">
                            In Progress
                          </SelectItem>
                          <SelectItem value="resolved">Resolved</SelectItem>
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="resolution_notes"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Resolution Notes</FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder="Explain what action was taken..."
                          className="resize-none"
                          rows={3}
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button type="submit" disabled={loading} className="w-full">
                  {loading && (
                    <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                  )}
                  Save Changes
                </Button>
              </form>
            </Form>
          </div>
        </div>

        <SheetFooter className="px-4 pb-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
