"use client";

import { ReportViewDrawer } from "@/components/reports/report-view-drawer";
import { ReportsDataTable } from "@/components/reports/reports-data-table";
import { apiGet } from "@/lib/api-client";
import { REPORTS_ENDPOINTS } from "@/lib/api/endpoints";
import type { Report } from "@/lib/types/reports.types";
import * as React from "react";
import { useNavigate, useSearchParams } from "react-router";

export default function ReportsPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const viewId = searchParams.get("view");

  const [viewReport, setViewReport] = React.useState<Report | null>(null);

  React.useEffect(() => {
    if (!viewId) {
      setViewReport(null);
      return;
    }
    apiGet(REPORTS_ENDPOINTS.GET(viewId))
      .then((r) => (r.ok ? r.json() : null))
      .then((json) => {
        if (json) setViewReport(json.data ?? json);
      })
      .catch(() => {});
  }, [viewId]);

  const handleViewClose = (open: boolean) => {
    if (!open) {
      const params = new URLSearchParams(searchParams.toString());
      params.delete("view");
      const qs = params.toString();
      navigate(qs ? `/reports?${qs}` : "/reports");
      setViewReport(null);
    }
  };

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <h1 className="text-3xl font-bold tracking-tight">Report Management</h1>
        <p className="text-muted-foreground mt-2">
          Review and resolve user reports.
        </p>
      </div>
      <div className="px-4 lg:px-6">
        <ReportsDataTable onViewRequest={(report) => setViewReport(report)} />
      </div>

      <ReportViewDrawer
        report={viewReport}
        open={!!viewReport}
        onOpenChange={handleViewClose}
        onSuccess={() => setViewReport(null)}
      />
    </div>
  );
}
