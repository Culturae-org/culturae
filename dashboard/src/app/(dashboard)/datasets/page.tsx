"use client";

import { DatasetsDataTable } from "@/components/datasets/datasets-data-table";
import { ImportDetailsDialog } from "@/components/datasets/datasets-import-details-dialog";
import { ImportDatasetsDataTable } from "@/components/datasets/datasets-imports-data-table";
import { DatasetsViewDialog } from "@/components/datasets/datasets-view-dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { ImportJob } from "@/lib/types/datasets.types";
import * as React from "react";
import { useSearchParams } from "react-router";

export default function DatasetsPage() {
  const [searchParams] = useSearchParams();
  const viewId = searchParams.get("view");
  const datasetViewId = searchParams.get("datasetView");
  const [_selectedJob, _setSelectedJob] = React.useState<ImportJob | null>(null);
  const [refreshKey, setRefreshKey] = React.useState(0);

  const handleRefresh = React.useCallback(() => {
    setRefreshKey((prev) => prev + 1);
  }, []);

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div>
          <h1 className="text-3xl font-bold">Datasets Management</h1>
          <p className="text-muted-foreground">
            Manage question and geography datasets from manifests
          </p>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Tabs defaultValue="datasets" className="w-full">
          <TabsList className="flex w-full overflow-x-auto gap-2 flex-nowrap">
            <TabsTrigger value="datasets">Datasets</TabsTrigger>
            <TabsTrigger value="history">History</TabsTrigger>
          </TabsList>

          <TabsContent value="datasets" className="mt-6">
            <DatasetsDataTable
              key={`datasets-${refreshKey}`}
              onRefresh={handleRefresh}
            />
          </TabsContent>

          <TabsContent value="history" className="mt-6">
            <ImportDatasetsDataTable key={`imports-${refreshKey}`} />
          </TabsContent>
        </Tabs>
      </div>

      <ImportDetailsDialog viewId={viewId} />

      <DatasetsViewDialog datasetId={datasetViewId} />
    </div>
  );
}
