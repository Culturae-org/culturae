"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import type { AdminUser } from "@/lib/types/user.types";

interface UserPrivacyCardProps {
  user: AdminUser;
}

export function UserPrivacyCard({ user }: UserPrivacyCardProps) {
  const privacySettings = [
    {
      label: "Profile Public",
      desc: "Profile visibility",
      val: user.is_profile_public,
    },
    {
      label: "Show Online Status",
      desc: "Visible when online",
      val: user.show_online_status,
    },
    {
      label: "Allow Friend Requests",
      desc: "Others can send requests",
      val: user.allow_friend_requests,
    },
    {
      label: "Allow Party Invites",
      desc: "Others can invite to party",
      val: user.allow_party_invites,
    },
  ];

  return (
    <Card className="border-0 dark:border">
      <CardHeader>
        <CardTitle className="flex items-center">Privacy Settings</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {privacySettings.map(({ label, desc, val }) => (
            <div
              key={label}
              className="flex items-center justify-between p-3 border rounded-lg"
            >
              <div>
                <Label className="text-sm font-medium">{label}</Label>
                <p className="text-xs text-muted-foreground">{desc}</p>
              </div>
              <Badge variant={val ? "default" : "secondary"}>
                {val ? "Enabled" : "Disabled"}
              </Badge>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
