import "./app/globals.css";
import AnalyticsPage from "@/app/(dashboard)/analytics/page";
import ApiExplorerPage from "@/app/(dashboard)/api-explorer/page";
// import WsExplorerPage from "@/app/(dashboard)/ws-explorer/page";
import DatasetsPage from "@/app/(dashboard)/datasets/page";
import GameTemplatesPage from "@/app/(dashboard)/game-templates/page";
import GameDetailPage from "@/app/(dashboard)/games/[id]/page";
import GamesPage from "@/app/(dashboard)/games/page";
import GeographyPage from "@/app/(dashboard)/geography/page";
import LogsPage from "@/app/(dashboard)/logs/page";
import HomePage from "@/app/(dashboard)/page";
import QuestionsPage from "@/app/(dashboard)/questions/page";
import PodsPage from "@/app/(dashboard)/pods/page";
import ReportsPage from "@/app/(dashboard)/reports/page";
import SettingsPage from "@/app/(dashboard)/settings/page";
import UserDetailPage from "@/app/(dashboard)/users/[id]/page";
import UsersPage from "@/app/(dashboard)/users/page";
import { AuthGuard } from "@/components/auth-guard";
import DashboardLayout from "@/components/dashboard-layout";
import ScrollToTop from "@/components/scroll-to-top";
import { Toaster } from "@/components/ui/sonner";
import { AppProviders } from "@/lib/stores";
import LoginPage from "@/pages/login";
import NotFoundPage from "@/pages/not-found";
import { ThemeProvider } from "next-themes";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router";

const root = document.getElementById("root");
if (!root) throw new Error("Root element not found");

createRoot(root).render(
  <StrictMode>
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      disableTransitionOnChange
    >
      <AppProviders>
        <BrowserRouter basename="/console">
          <ScrollToTop />
          <Routes>
            <Route
              path="/login"
              element={
                <AuthGuard requireAuth={false}>
                  <LoginPage />
                </AuthGuard>
              }
            />
            <Route element={<DashboardLayout />}>
              <Route index element={<HomePage />} />
              <Route path="/users" element={<UsersPage />} />
              <Route path="/users/:id" element={<UserDetailPage />} />
              <Route path="/games" element={<GamesPage />} />
              <Route path="/games/:id" element={<GameDetailPage />} />
              <Route path="/datasets" element={<DatasetsPage />} />
              <Route path="/questions" element={<QuestionsPage />} />
              <Route path="/analytics" element={<AnalyticsPage />} />
              <Route path="/geography" element={<GeographyPage />} />
              <Route path="/game-templates" element={<GameTemplatesPage />} />
              <Route path="/logs" element={<LogsPage />} />
              <Route path="/pods" element={<PodsPage />} />
              <Route path="/reports" element={<ReportsPage />} />
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="/api-explorer" element={<ApiExplorerPage />} />
              {/* <Route path="/ws-explorer" element={<WsExplorerPage />} /> */}
              <Route path="*" element={<NotFoundPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </AppProviders>
      <Toaster position="bottom-left" />
    </ThemeProvider>
  </StrictMode>,
);
