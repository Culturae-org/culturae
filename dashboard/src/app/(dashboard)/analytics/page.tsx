import { DashboardCharts } from "@/components/dashboard/dashboard-charts";
import { SectionCardsTwo } from "@/components/dashboard/dashboard-section-cards-two";

export default function AnalyticsPage() {
  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <SectionCardsTwo />
      <DashboardCharts />
    </div>
  );
}
