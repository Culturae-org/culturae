"use client";

import { NavDocuments } from "@/components/nav-documents";
import { NavMain } from "@/components/nav-main";
import { NavSecondary } from "@/components/nav-secondary";
import { NavUser } from "@/components/nav-user";
import { Separator } from "@/components/ui/separator";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { useUser } from "@/lib/stores";
import {
  IconApi,
  IconBrandGithub,
  IconChartBar,
  IconDashboard,
  IconDatabase,
  IconDeviceGamepad2,
  IconFileSad,
  IconQuestionMark,
  IconReport,
  IconServer,
  IconSettings,
  IconTemplate,
  IconUsers,
  IconWorld,
} from "@tabler/icons-react";
import type * as React from "react";
import { useEffect, useState } from "react";
import { FcGlobe } from "react-icons/fc";
import { Link } from "react-router";

const NAV_MAIN = [
  {
    title: "Dashboard",
    url: "/",
    icon: IconDashboard,
  },
  {
    title: "Users",
    url: "/users",
    icon: IconUsers,
  },
  {
    title: "Datasets",
    url: "/datasets",
    icon: IconDatabase,
  },
  {
    title: "Questions",
    url: "/questions",
    icon: IconQuestionMark,
  },
  {
    title: "Geography",
    url: "/geography",
    icon: IconWorld,
  },
  {
    title: "Games",
    url: "/games",
    icon: IconDeviceGamepad2,
  },
  {
    title: "Game Templates",
    url: "/game-templates",
    icon: IconTemplate,
  },
  {
    title: "Analytics",
    url: "/analytics",
    icon: IconChartBar,
  },
  {
    title: "Logs",
    url: "/logs",
    icon: IconReport,
  },
  {
    title: "Pods",
    url: "/pods",
    icon: IconServer,
  },
  {
    title: "Reports",
    url: "/reports",
    icon: IconReport,
  },
  {
    title: "Documentation",
    url: "https://docs.culturae.me",
    icon: IconFileSad,
  },
  {
    title: "Api Explorer",
    url: "/api-explorer",
    icon: IconApi,
  },
  // STEEL IN DEV
  // {
  //   title: "WS Explorer",
  //   url: "/ws-explorer",
  //   icon: IconWebhook,
  // },
  {
    title: "Settings",
    url: "/settings",
    icon: IconSettings,
  },
];

const NAV_SECONDARY = [
  {
    title: "Github",
    url: "https://github.com/Culturae-org/culturae",
    icon: IconBrandGithub,
  },
  {
    title: "Cultpedia",
    url: "https://github.com/Culturae-org/cultpedia",
    icon: IconWorld,
  },
  {
    title: "Issues & Problems",
    url: "https://github.com/Culturae-org/culturae/issues",
    icon: IconReport,
  },
];

const DOCUMENTS: { name: string; url: string; icon: typeof IconDatabase }[] =
  [];

const fallbackData = {
  user: {
    id: "",
    name: "culturae",
    email: "culturae@example.com",
    has_avatar: false,
  },
  navMain: NAV_MAIN,
  navSecondary: NAV_SECONDARY,
  documents: DOCUMENTS,
};

type SidebarData = typeof fallbackData;

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { user, userId } = useUser();
  const [data, setData] = useState<SidebarData>(fallbackData);

  useEffect(() => {
    if (user) {
      setData({
        ...fallbackData,
        user: {
          id: userId ?? "",
          name: user.username ?? user.email,
          email: user.email ?? user.username,
          has_avatar: user.has_avatar ?? false,
        },
      });
    }
  }, [user, userId]);

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:p-1.5!"
            >
              <Link to="/">
                <FcGlobe strokeWidth={1} className="size-5!" />
                <span className="text-base font-semibold">Culturae.</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <Separator />
      <div className="py-3" />
      <SidebarContent>
        <NavMain items={data.navMain} />
        {data.documents.length > 0 && <NavDocuments items={data.documents} />}
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
    </Sidebar>
  );
}
