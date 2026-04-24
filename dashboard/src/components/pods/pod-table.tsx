import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { PodInfo } from "@/lib/types/metrics.types";
import { IconServerOff } from "@tabler/icons-react";
import { useMemo, useState } from "react";
import { PodRow } from "./pod-row";

interface PodTableProps {
  pods: PodInfo[];
  mainPods: PodInfo[];
  headlessPods: PodInfo[];
  loading: boolean;
}

export function PodTable({ pods, mainPods, headlessPods, loading }: PodTableProps) {
  const [activeTab, setActiveTab] = useState("all");

  const filteredPods = useMemo(() => {
    if (activeTab === "main") return mainPods;
    if (activeTab === "headless") return headlessPods;
    return pods;
  }, [activeTab, pods, mainPods, headlessPods]);

  return (
    <div className="px-4 lg:px-6">
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="all">All ({pods.length})</TabsTrigger>
          <TabsTrigger value="main">Main ({mainPods.length})</TabsTrigger>
          <TabsTrigger value="headless">Headless ({headlessPods.length})</TabsTrigger>
        </TabsList>

        {(["all", "main", "headless"] as const).map((tab) => (
          <TabsContent key={tab} value={tab} className="mt-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-base font-medium">
                  {tab === "all" ? "All pods" : tab === "main" ? "Main pods" : "Headless pods"}
                </CardTitle>
                <CardDescription>
                  {filteredPods.length} pod{filteredPods.length !== 1 ? "s" : ""} active
                </CardDescription>
              </CardHeader>
              <CardContent className="p-0">
                {loading ? (
                  <PodTableSkeleton />
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Status</TableHead>
                        <TableHead>Pod ID</TableHead>
                        <TableHead>Type</TableHead>
                        <TableHead className="text-right">Clients</TableHead>
                        <TableHead className="text-right">Users</TableHead>
                        <TableHead className="text-right">Games</TableHead>
                        <TableHead className="text-right">Load</TableHead>
                        <TableHead>Uptime</TableHead>
                        <TableHead>Heartbeat</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredPods.map((pod) => (
                        <PodRow key={pod.pod_id} pod={pod} />
                      ))}
                      {filteredPods.length === 0 && (
                        <TableRow>
                          <TableCell colSpan={9} className="text-center py-12 text-muted-foreground">
                            <div className="flex flex-col items-center gap-2">
                              <IconServerOff className="h-6 w-6" />
                              <p className="text-sm">No pods found</p>
                            </div>
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        ))}
      </Tabs>
    </div>
  );
}

function PodTableSkeleton() {
  return (
    <div className="p-6 space-y-3">
      {[1, 2, 3].map((i) => (
        <Skeleton key={i} className="h-10 w-full" />
      ))}
    </div>
  );
}
