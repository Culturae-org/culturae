"use client";

import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useIsMobile } from "@/hooks/useMobile";
import { IconRefresh, IconX } from "@tabler/icons-react";
import * as React from "react";
import slugify from "slugify";
import { toast } from "sonner";

interface CreateDatasetForm {
  name: string;
  slug: string;
  description: string;
  version: string;
}

interface CreateDatasetDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreate: (data: {
    name: string;
    slug: string;
    description?: string;
    version: string;
    source?: string;
  }) => Promise<unknown>;
  creating: boolean;
}

const defaultForm: CreateDatasetForm = {
  name: "",
  slug: "",
  description: "",
  version: "1.0.0",
};

export function CreateDatasetDrawer({
  open,
  onOpenChange,
  onCreate,
  creating,
}: CreateDatasetDrawerProps) {
  const isMobile = useIsMobile();
  const [form, setForm] = React.useState<CreateDatasetForm>(defaultForm);

  React.useEffect(() => {
    if (open) {
      setForm(defaultForm);
    }
  }, [open]);

  const handleCreate = async () => {
    if (!form.name || !form.slug) {
      toast.error("Name and slug are required");
      return;
    }
    try {
      await onCreate({
        name: form.name,
        slug: form.slug,
        description: form.description,
        version: form.version,
        source: "custom",
      });
      onOpenChange(false);
      setForm(defaultForm);
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to create dataset",
      );
    }
  };

  return (
    <Drawer
      open={open}
      onOpenChange={onOpenChange}
      direction={isMobile ? "bottom" : "right"}
    >
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>New Dataset</DrawerTitle>
        </DrawerHeader>
        <div className="flex-1 overflow-y-auto p-4 space-y-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="ds-name">
                Name <span className="text-destructive">*</span>
              </Label>
              <Input
                id="ds-name"
                value={form.name}
                onChange={(e) => {
                  const name = e.target.value;
                  const slug = slugify(name, {
                    lower: true,
                    strict: true,
                    remove: /[*+~.()'"!:@]/g,
                  });
                  setForm((prev) => ({ ...prev, name, slug }));
                }}
                placeholder="My Custom Dataset"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ds-slug">
                Slug <span className="text-destructive">*</span>
              </Label>
              <div className="flex gap-2">
                <Input
                  id="ds-slug"
                  value={form.slug}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, slug: e.target.value }))
                  }
                  placeholder="my-custom-dataset"
                />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={() => {
                    const slug = slugify(form.name, {
                      lower: true,
                      strict: true,
                      remove: /[*+~.()'"!:@]/g,
                    });
                    setForm((prev) => ({ ...prev, slug }));
                  }}
                >
                  <IconX className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="ds-description">Description</Label>
              <Textarea
                id="ds-description"
                value={form.description}
                onChange={(e) =>
                  setForm((prev) => ({
                    ...prev,
                    description: e.target.value,
                  }))
                }
                placeholder="A brief description of this dataset"
                rows={3}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ds-version">Version</Label>
              <Input
                id="ds-version"
                value={form.version}
                onChange={(e) =>
                  setForm((prev) => ({
                    ...prev,
                    version: e.target.value,
                  }))
                }
                placeholder="1.0.0"
              />
            </div>
          </div>
        </div>
        <DrawerFooter>
          <Button
            onClick={handleCreate}
            disabled={creating || !form.name || !form.slug}
          >
            {creating && <IconRefresh className="h-4 w-4 mr-2 animate-spin" />}
            Create Dataset
          </Button>
          <DrawerClose asChild>
            <Button variant="outline">Cancel</Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
