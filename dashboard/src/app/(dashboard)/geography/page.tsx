"use client";

import { ContinentsDataTable } from "@/components/geography/continent/continents-data-table";
import { CountriesDataTable } from "@/components/geography/country/countries-data-table";
import { GeographyDatasetSelector } from "@/components/geography/geography-dataset-selector";
import { GeographyStatsCards } from "@/components/geography/geography-stats-cards";
import { RegionsDataTable } from "@/components/geography/region/region-data-table";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useGeographyDatasets } from "@/hooks/useDatasets";
import { apiGet } from "@/lib/api-client";
import { DATASETS_ENDPOINTS } from "@/lib/api/endpoints";
import { Separator } from "@radix-ui/react-dropdown-menu";
import { IconFlag, IconWorld } from "@tabler/icons-react";
import * as React from "react";
import { useNavigate } from "react-router";

interface GeographyStats {
  country_count: number;
  continent_count: number;
  region_count: number;
  flag_count: number;
  flag_png512_count?: number;
  flag_png1024_count?: number;
}

export default function GeographyPage() {
  const navigate = useNavigate();
  const { datasets, loading } = useGeographyDatasets();
  const [selectedDatasetId, setSelectedDatasetId] = React.useState<string>("");
  const [stats, setStats] = React.useState<GeographyStats | null>(null);

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

  React.useEffect(() => {
    const fetchStats = async () => {
      if (!selectedDatasetId) return;
      try {
        const response = await apiGet(
          DATASETS_ENDPOINTS.GET_GEO_STATS(selectedDatasetId),
        );
        if (response.ok) {
          const data = await response.json();
          setStats(data.data ?? data);
        }
      } catch (error) {
        console.error("Failed to fetch stats:", error);
      }
    };
    fetchStats();
  }, [selectedDatasetId]);

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
          <h1 className="text-3xl font-bold">Geography</h1>
          <p className="text-muted-foreground">
            Manage countries, continents, and regions
          </p>
        </div>
        <div className="px-4 lg:px-6">
          <Card className="border-0 dark:border">
            <CardContent className="flex flex-col items-center justify-center py-12">
              <IconWorld className="h-16 w-16 text-muted-foreground mb-4" />
              <h3 className="text-xl font-semibold">No Geography Data</h3>
              <p className="text-muted-foreground text-center max-w-md mt-2 mb-6">
                Import a geography dataset from Cultpedia to get started.
              </p>
              <Button onClick={() => navigate("/datasets")}>
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
              <h1 className="text-3xl font-bold">Geography</h1>
              <p className="text-muted-foreground">
                Manage countries, continents, and regions
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" onClick={() => navigate("/datasets")}>
                <IconWorld className="mr-2 h-4 w-4" />
                Import Geography
              </Button>
              <GeographyDatasetSelector
                datasets={datasets}
                selectedDatasetId={selectedDatasetId}
                onDatasetChange={setSelectedDatasetId}
              />
            </div>
          </div>
        </div>

        <div className="px-4 lg:px-6">
          <Tabs defaultValue="countries" className="w-full">
            <div className="flex items-center justify-between">
              <TabsList>
                <TabsTrigger
                  value="countries"
                  className="flex items-center gap-2"
                >
                  <IconFlag className="h-4 w-4" />
                  Countries
                </TabsTrigger>
                <TabsTrigger
                  value="continents-regions"
                  className="flex items-center gap-2"
                >
                  <IconWorld className="h-4 w-4" />
                  Continents & Regions
                </TabsTrigger>
              </TabsList>
              <div>
                <GeographyStatsCards stats={stats} />
              </div>
            </div>

            <TabsContent value="countries" className="mt-4">
              {selectedDatasetId && (
                <CountriesDataTable datasetId={selectedDatasetId} />
              )}
            </TabsContent>

            <TabsContent value="continents-regions" className="mt-4">
              {selectedDatasetId && (
                <ContinentsDataTable datasetId={selectedDatasetId} />
              )}

              <Separator className="my-6" />

              {selectedDatasetId && (
                <RegionsDataTable datasetId={selectedDatasetId} />
              )}
            </TabsContent>
          </Tabs>
        </div>
      </div>
  );
}
