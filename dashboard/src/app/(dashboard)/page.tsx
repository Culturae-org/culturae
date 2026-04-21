import { DashboardCharts } from "@/components/dashboard/dashboard-charts";
import { SectionCardsOne } from "@/components/dashboard/dashboard-section-cards-one";
import { SectionCardsTwo } from "@/components/dashboard/dashboard-section-cards-two";
import UsersPage from "@/components/users/users-page";

export default function Page() {
  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <SectionCardsOne />
      <SectionCardsTwo />
      <DashboardCharts />
      <UsersPage />
    </div>
  );
}
