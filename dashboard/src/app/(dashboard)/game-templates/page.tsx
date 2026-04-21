"use client";

import { GameTemplateEditDialog } from "@/components/game-templates/game-template-edit-dialog";
import { GameTemplateViewDialog } from "@/components/game-templates/game-template-view-dialog";
import { GameTemplatesDataTable } from "@/components/game-templates/game-templates-data-table";
import { apiGet } from "@/lib/api-client";
import { GAME_TEMPLATES_ENDPOINTS } from "@/lib/api/endpoints";
import { gameTemplatesService } from "@/lib/services/game-templates.service";
import type {
  GameTemplate,
  UpdateGameTemplateRequest,
} from "@/lib/types/game-template.types";
import * as React from "react";
import { useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

export default function GameTemplatesPage() {
  const [_total, setTotal] = useState(0);
  const [refreshKey, setRefreshKey] = useState(0);

  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const viewId = searchParams.get("view");

  const [viewTemplate, setViewTemplate] = React.useState<GameTemplate | null>(
    null,
  );
  const [editTemplate, setEditTemplate] = React.useState<GameTemplate | null>(
    null,
  );
  const [editOpen, setEditOpen] = React.useState(false);

  React.useEffect(() => {
    if (!viewId) {
      setViewTemplate(null);
      return;
    }
    apiGet(GAME_TEMPLATES_ENDPOINTS.GET(viewId))
      .then((r) => (r.ok ? r.json() : null))
      .then((json) => {
        if (json) setViewTemplate(json.data ?? json);
      })
      .catch(() => {});
  }, [viewId]);

  const handleViewClose = (open: boolean) => {
    if (!open) {
      const params = new URLSearchParams(searchParams.toString());
      params.delete("view");
      const qs = params.toString();
      navigate(qs ? `/game-templates?${qs}` : "/game-templates");
      setViewTemplate(null);
    }
  };

  const handleEditClick = () => {
    if (!viewTemplate) return;
    setEditTemplate(viewTemplate);
    handleViewClose(false);
    setEditOpen(true);
  };

  const handleUpdate = async (id: string, data: UpdateGameTemplateRequest) => {
    const updated = await gameTemplatesService.updateTemplate(id, data);
    setRefreshKey((k) => k + 1);
    return updated;
  };

  return (
    <div className="flex flex-col gap-6 py-4 md:gap-8 md:py-6">
      <div className="px-4 lg:px-6">
        <h1 className="text-3xl font-bold">Game Templates</h1>
        <p className="text-muted-foreground">
          Manage all game modes — configure dataset, scoring, player limits, and
          activate or suspend them.
        </p>
      </div>

      <GameTemplatesDataTable
        onTotalCountChange={setTotal}
        refreshKey={refreshKey}
      />

      {viewTemplate && (
        <GameTemplateViewDialog
          template={viewTemplate}
          open={true}
          onOpenChange={handleViewClose}
          onEditClick={handleEditClick}
        />
      )}

      {editTemplate && (
        <GameTemplateEditDialog
          template={editTemplate}
          open={editOpen}
          onOpenChange={(open) => {
            setEditOpen(open);
            if (!open) setEditTemplate(null);
          }}
          onUpdated={handleUpdate}
        />
      )}
    </div>
  );
}
