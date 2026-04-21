"use client";

import { AdminActionLogsDataTable } from "@/components/admin/admin-action-logs-data-table";
import { AdminActionViewDialog } from "@/components/admin/admin-action-view-dialog";
import { AdminActionsByResourceChart } from "@/components/admin/admin-actions-by-resource-chart";
import { AdminActionsByTypeChart } from "@/components/admin/admin-actions-by-type-chart";
import { ApiRequestsByMethodChart } from "@/components/api/api-requests-by-method-chart";

import { ApiRequestsChart } from "@/components/api/api-requests-chart";
import { SystemMetricsCards } from "@/components/logs/system-metrics-cards";
import { ServiceDetailsDialog } from "@/components/service/service-details-dialog";
import { ServiceHealthCard } from "@/components/service/service-health-card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { apiGet } from "@/lib/api-client";
import { LOGS_ENDPOINTS, USERS_ENDPOINTS } from "@/lib/api/endpoints";
import type { AdminActionLog } from "@/lib/types/logs.types";
import type { SystemMetrics } from "@/lib/types/metrics.types";
import type { ServiceStatus } from "@/lib/types/stats.types";
import * as React from "react";
import { useNavigate, useSearchParams } from "react-router";

export default function LogsPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const adminViewId = searchParams.get("admin-view");
  const serviceParam = searchParams.get("service");

  const [viewAdminLog, setViewAdminLog] = React.useState<AdminActionLog | null>(
    null,
  );

  const [activeTab, setActiveTab] = React.useState("overview");
  const [serviceStatus, setServiceStatus] = React.useState<ServiceStatus[]>([]);
  const [systemMetrics, setSystemMetrics] =
    React.useState<SystemMetrics | null>(null);
  const [onlineUsers, setOnlineUsers] = React.useState<number>(0);
  const [_loading, setLoading] = React.useState(true);
  const [selectedService, setSelectedService] =
    React.useState<ServiceStatus | null>(null);
  const [serviceDialogOpen, setServiceDialogOpen] = React.useState(false);

  React.useEffect(() => {
    if (adminViewId) setActiveTab("admin");
  }, [adminViewId]);

  const handleAdminViewClose = (open: boolean) => {
    if (!open) {
      const params = new URLSearchParams(searchParams.toString());
      params.delete("admin-view");
      const qs = params.toString();
      navigate(qs ? `/logs?${qs}` : "/logs");
      setViewAdminLog(null);
    }
  };

  const selectedServiceRef = React.useRef<ServiceStatus | null>(null);
  selectedServiceRef.current = selectedService;

  const loadServiceStatus = React.useCallback(async () => {
    try {
      const response = await apiGet(LOGS_ENDPOINTS.SERVICE_STATUS);
      if (response.ok) {
        const data = await response.json();
        const services = data.data?.services ?? data.services ?? [];
        setServiceStatus(services);
        if (selectedServiceRef.current) {
          const updatedService = services.find(
            (s: ServiceStatus) =>
              s.service_name === selectedServiceRef.current?.service_name,
          );
          if (updatedService) setSelectedService(updatedService);
        }
      }
    } catch (error) {
      console.error("Error loading service status:", error);
    }
  }, []);

  React.useEffect(() => {
    if (serviceParam && serviceStatus.length > 0 && !serviceDialogOpen) {
      const service = serviceStatus.find(
        (s) => s.service_name === serviceParam,
      );
      if (service) {
        setSelectedService(service);
        setServiceDialogOpen(true);
      }
    }
  }, [serviceParam, serviceStatus, serviceDialogOpen]);

  const loadSystemMetrics = React.useCallback(async () => {
    try {
      const [metricsResponse, onlineResponse] = await Promise.all([
        apiGet(LOGS_ENDPOINTS.SYSTEM_METRICS),
        apiGet(USERS_ENDPOINTS.COUNT_ONLINE),
      ]);

      if (metricsResponse.ok) {
        const data = await metricsResponse.json();
        const metrics = data.data ?? data;
        setSystemMetrics({
          total_users: metrics.total_users ?? 0,
          active_users: metrics.active_users ?? 0,
          total_sessions: metrics.total_sessions ?? 0,
          active_sessions: metrics.active_sessions ?? 0,
          total_api_requests: metrics.total_api_requests ?? 0,
          error_rate: metrics.error_rate ?? 0,
          avg_response_time_ms: metrics.avg_response_time_ms ?? 0,
        });
      }

      if (onlineResponse.ok) {
        const data = await onlineResponse.json();
        setOnlineUsers(
          typeof data === "number"
            ? data
            : typeof data.data === "number"
              ? data.data
              : data.count || 0,
        );
      }
    } catch (error) {
      console.error("Error loading system metrics:", error);
    }
  }, []);

  const loadAllData = React.useCallback(async () => {
    try {
      await Promise.all([loadServiceStatus(), loadSystemMetrics()]);
    } catch (error) {
      console.error("Error loading logs data:", error);
    } finally {
      setLoading(false);
    }
  }, [loadServiceStatus, loadSystemMetrics]);

  React.useEffect(() => {
    loadAllData();
    const interval = setInterval(loadAllData, 30000);
    return () => clearInterval(interval);
  }, [loadAllData]);

  const handleServiceClick = (service: ServiceStatus) => {
    setSelectedService(service);
    setServiceDialogOpen(true);
  };

  const handleServiceDialogClose = (open: boolean) => {
    setServiceDialogOpen(open);
  };

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">System Logs & Monitoring</h1>
            <p className="text-muted-foreground">
              Monitor system activity, API requests, and service health
            </p>
          </div>
        </div>
      </div>

      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="px-4 lg:px-6"
      >
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="admin">Admin Actions</TabsTrigger>
          <TabsTrigger value="api">API Requests</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          <SystemMetricsCards
            systemMetrics={systemMetrics}
            onlineUsers={onlineUsers}
          />
          <ServiceHealthCard
            services={serviceStatus}
            onServiceClick={handleServiceClick}
          />
        </TabsContent>

        <TabsContent value="admin" className="space-y-4">
          <AdminActionLogsDataTable
            onViewRequest={(log) => setViewAdminLog(log)}
          />
          <div className="flex flex-col lg:flex-row gap-4">
            <AdminActionsByResourceChart />
            <AdminActionsByTypeChart />
          </div>
        </TabsContent>

        <TabsContent value="api" className="space-y-4">
          <ApiRequestsChart />
          <div className="flex flex-col lg:flex-row gap-4">
            <ApiRequestsByMethodChart />
          </div>
        </TabsContent>
      </Tabs>

      <ServiceDetailsDialog
        service={selectedService}
        open={serviceDialogOpen}
        onOpenChange={handleServiceDialogClose}
        onRefresh={loadServiceStatus}
      />

      {viewAdminLog && (
        <AdminActionViewDialog
          action={viewAdminLog}
          open={true}
          onOpenChange={handleAdminViewClose}
        />
      )}
    </div>
  );
}
