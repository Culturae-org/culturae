"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { apiPatch } from "@/lib/api-client";
import { REPORTS_ENDPOINTS } from "@/lib/api/endpoints";
import type { Report } from "@/lib/types/reports.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { IconRefresh } from "@tabler/icons-react";
import * as React from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";

const formSchema = z.object({
  status: z.enum(["pending", "in_progress", "resolved"]),
  resolution_notes: z.string().optional(),
});

interface ReportActionDialogProps {
  report: Report | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function ReportActionDialog({
  report,
  open,
  onOpenChange,
  onSuccess,
}: ReportActionDialogProps) {
  const [loading, setLoading] = React.useState(false);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      status: report?.status || "pending",
      resolution_notes: report?.resolution_notes || "",
    },
  });

  React.useEffect(() => {
    if (report) {
      form.reset({
        status: report.status,
        resolution_notes: report.resolution_notes || "",
      });
    }
  }, [report, form]);

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    if (!report) return;

    setLoading(true);
    try {
      const response = await apiPatch(
        REPORTS_ENDPOINTS.UPDATE_STATUS(report.id),
        values,
      );

      if (!response.ok) throw new Error("Failed to update report");

      toast.success("Report updated successfully");
      onSuccess();
      onOpenChange(false);
    } catch (error) {
      toast.error("Failed to update report");
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  if (!report) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Update Report Status</DialogTitle>
          <DialogDescription>
            Update the status and add resolution notes for this report.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="status"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Status</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a status" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="pending">Pending</SelectItem>
                      <SelectItem value="in_progress">In Progress</SelectItem>
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
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button type="submit" disabled={loading}>
                {loading && (
                  <IconRefresh className="h-4 w-4 mr-2 animate-spin" />
                )}
                Save Changes
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
