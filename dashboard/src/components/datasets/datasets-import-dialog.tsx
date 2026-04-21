"use client";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiPost } from "@/lib/api-client";
import { DATASETS_ENDPOINTS } from "@/lib/api/endpoints";
import type { ImportResult } from "@/lib/types/datasets.types";
import type { GeographyImportResult } from "@/lib/types/geography.types";
import {
  IconCircleCheck,
  IconCircleX,
  IconDownload,
  IconExternalLink,
  IconFlag,
  IconMapPin,
  IconRefresh,
  IconWorld,
} from "@tabler/icons-react";
import * as React from "react";
import { toast } from "sonner";

const DEFAULT_QUESTIONS_MANIFEST_URL =
  "https://raw.githubusercontent.com/Culturae-org/cultpedia/refs/heads/main/datasets/general-knowledge/manifest.json";
const DEFAULT_GEOGRAPHY_MANIFEST_URL =
  "https://raw.githubusercontent.com/Culturae-org/cultpedia/refs/heads/main/datasets/geography/manifest.json";

const CULTPEDIA_TEMPLATE_REPO =
  "https://github.com/Culturae-org/cultpedia-dataset-template";

type ImportType = "questions" | "geography";

export function ImportDatasetDialog({
  onImportComplete,
  open,
  onOpenChange,
  defaultManifestUrl,
  defaultType = "questions",
}: {
  onImportComplete?: () => void;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  defaultManifestUrl?: string;
  defaultType?: ImportType;
}) {
  const [manifestUrl, setManifestUrl] = React.useState("");
  const [importType, setImportType] = React.useState<ImportType>(defaultType);
  const [setAsDefault, setSetAsDefault] = React.useState(false);
  const [importing, setImporting] = React.useState(false);
  const [result, setResult] = React.useState<
    ImportResult | GeographyImportResult | null
  >(null);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (open) {
      setImportType(defaultType);
      setManifestUrl("");
      setError(null);
      setResult(null);
    }
  }, [open, defaultType]);

  const useCultpedia = () => {
    const url =
      importType === "geography"
        ? DEFAULT_GEOGRAPHY_MANIFEST_URL
        : DEFAULT_QUESTIONS_MANIFEST_URL;
    setManifestUrl(url);
  };

  const handleImport = async (e: React.FormEvent) => {
    e.preventDefault();
    setImporting(true);
    setError(null);
    setResult(null);

    try {
      const response = await apiPost(DATASETS_ENDPOINTS.IMPORT, {
        dataset_type: importType,
        manifest_url: manifestUrl,
        set_as_default: setAsDefault,
      });

      if (!response.ok) {
        const errorData = await response.json();
        const message =
          errorData.error?.message || errorData.error || "Import failed";
        throw new Error(message);
      }

      const rawData = await response.json();
      const importResult =
        "data" in rawData
          ? (rawData as { data: ImportResult | GeographyImportResult }).data
          : (rawData as ImportResult | GeographyImportResult);
      setResult(importResult);

      if (importResult.success) {
        if (importType === "geography") {
          const geoResult = importResult as GeographyImportResult;
          toast.success("Geography dataset imported successfully", {
            description: `${geoResult.countries_added} countries, ${geoResult.continents_added} continents, ${geoResult.regions_added} regions added`,
          });
        } else {
          const qResult = importResult as ImportResult;
          toast.success("Dataset imported successfully", {
            description: `${qResult.questions_added} questions added`,
          });
        }
        onImportComplete?.();
      }
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to import dataset";
      setError(errorMessage);
    } finally {
      setImporting(false);
    }
  };

  const isGeography = importType === "geography";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            Import {isGeography ? "Geography" : "Questions"} Dataset
          </DialogTitle>
          <DialogDescription>
            {isGeography
              ? "Import countries, continents, and regions from a Cultpedia geography manifest"
              : "Import questions from a manifest URL (NDJSON format)"}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleImport}>
          <div className="grid gap-4 py-4">
            <div className="flex gap-2 mb-2">
              <Button
                type="button"
                variant={importType === "geography" ? "default" : "outline"}
                size="sm"
                onClick={() => setImportType("geography")}
                disabled={importing}
              >
                <IconWorld className="mr-2 h-4 w-4" />
                Geography
              </Button>
              <Button
                type="button"
                variant={importType === "questions" ? "default" : "outline"}
                size="sm"
                onClick={() => setImportType("questions")}
                disabled={importing}
              >
                <IconDownload className="mr-2 h-4 w-4" />
                Questions
              </Button>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="manifest-url">Manifest URL</Label>
              <div className="flex gap-2">
                <Input
                  id="manifest-url"
                  placeholder="https://example.com/manifest.json"
                  value={manifestUrl}
                  onChange={(e) => setManifestUrl(e.target.value)}
                  disabled={importing}
                  required
                  className="flex-1"
                />
                <Button
                  type="button"
                  variant="secondary"
                  onClick={useCultpedia}
                  disabled={importing}
                >
                  Use Cultpedia
                </Button>
              </div>
              <p className="text-sm text-muted-foreground">
                Enter the URL of a manifest.json file.
              </p>
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox
                id="set-default"
                checked={setAsDefault}
                onCheckedChange={(checked) => setSetAsDefault(checked === true)}
                disabled={importing}
              />
              <Label
                htmlFor="set-default"
                className="text-sm font-normal cursor-pointer"
              >
                Set this dataset as default after import
              </Label>
            </div>

            <div className="rounded-lg border bg-muted/50 p-4">
              <div className="flex items-start">
                <div className="flex-1">
                  <div className="font-medium text-sm mb-1">
                    Want to create your own dataset?
                  </div>
                  <p className="text-sm text-muted-foreground mb-2">
                    Use our template to create a custom manifest with your own
                    questions.
                  </p>
                  <a
                    href={CULTPEDIA_TEMPLATE_REPO}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm hover:underline inline-flex items-center gap-1"
                  >
                    View template repository
                    <IconExternalLink className="h-3 w-3" />
                  </a>
                </div>
              </div>
            </div>

            {error && (
              <div className="rounded-lg border border-destructive bg-destructive/10 p-4">
                <div className="flex items-start gap-3">
                  <IconCircleX className="h-5 w-5 text-destructive mt-0.5" />
                  <div className="flex-1 text-sm text-destructive">{error}</div>
                </div>
              </div>
            )}

            {result && (
              <div
                className={`rounded-lg border p-4 ${
                  result.success
                    ? "border-green-500/50 bg-green-500/10"
                    : "border-destructive bg-destructive/10"
                }`}
              >
                <div className="flex items-start gap-3">
                  {result.success ? (
                    <IconCircleCheck className="h-5 w-5 text-green-500 mt-0.5" />
                  ) : (
                    <IconCircleX className="h-5 w-5 text-destructive mt-0.5" />
                  )}
                  <div className="flex-1">
                    <div className="font-medium mb-2">{result.message}</div>
                    {result.success && (
                      <div className="text-sm space-y-1">
                        {isGeography ? (
                          (() => {
                              const geoResult = result as GeographyImportResult;
                              return (
                                <>
                                  <div className="flex items-center gap-2">
                                    <IconFlag className="h-4 w-4 text-muted-foreground" />
                                    <span>
                                      Countries: {geoResult.countries_added}
                                    </span>
                                  </div>
                                  <div className="flex items-center gap-2">
                                    <IconWorld className="h-4 w-4 text-muted-foreground" />
                                    <span>
                                      Continents: {geoResult.continents_added}
                                    </span>
                                  </div>
                                  <div className="flex items-center gap-2">
                                    <IconMapPin className="h-4 w-4 text-muted-foreground" />
                                    <span>
                                      Regions: {geoResult.regions_added}
                                    </span>
                                  </div>
                                  {geoResult.flags_added > 0 && (
                                    <div className="flex items-center gap-2">
                                      <IconFlag className="h-4 w-4 text-muted-foreground" />
                                      <span>
                                        Flags SVG: {geoResult.flags_added}
                                      </span>
                                    </div>
                                  )}
                                  {geoResult.flags_png512_added > 0 && (
                                    <div className="flex items-center gap-2">
                                      <IconFlag className="h-4 w-4 text-muted-foreground" />
                                      <span>
                                        Flags PNG 512px:{" "}
                                        {geoResult.flags_png512_added}
                                      </span>
                                    </div>
                                  )}
                                  {geoResult.flags_png1024_added > 0 && (
                                    <div className="flex items-center gap-2">
                                      <IconFlag className="h-4 w-4 text-muted-foreground" />
                                      <span>
                                        Flags PNG 1024px:{" "}
                                        {geoResult.flags_png1024_added}
                                      </span>
                                    </div>
                                  )}
                                  {!geoResult.flags_added &&
                                    !geoResult.flags_png512_added &&
                                    !geoResult.flags_png1024_added && (
                                      <div className="flex items-center gap-2 text-muted-foreground">
                                        <IconFlag className="h-4 w-4" />
                                        <span className="text-xs">
                                          Flags importing in background...
                                        </span>
                                      </div>
                                    )}
                                </>
                              );
                            })()
                        ) : (
                          (() => {
                              const qResult = result as ImportResult;
                              return (
                                <>
                                  <div>
                                    Questions added: {qResult.questions_added}
                                  </div>
                                  {qResult.themes_added > 0 && (
                                    <div>
                                      Themes added: {qResult.themes_added}
                                    </div>
                                  )}
                                  {qResult.subthemes_added > 0 && (
                                    <div>
                                      Subthemes added: {qResult.subthemes_added}
                                    </div>
                                  )}
                                  {qResult.tags_added > 0 && (
                                    <div>Tags added: {qResult.tags_added}</div>
                                  )}
                                </>
                              );
                            })()
                        )}
                      </div>
                    )}
                    {result.errors && result.errors.length > 0 && (
                      <div className="mt-2 text-sm">
                        <div className="font-medium">
                          Errors ({result.errors.length}):
                        </div>
                        <ul className="list-disc list-inside max-h-32 overflow-y-auto">
                          {result.errors.slice(0, 10).map((err) => (
                            <li key={err} className="text-xs">
                              {err}
                            </li>
                          ))}
                          {result.errors.length > 10 && (
                            <li className="text-xs italic">
                              ... and {result.errors.length - 10} other errors
                            </li>
                          )}
                        </ul>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>

          <DialogFooter>
            {result?.success ? (
              <Button
                type="button"
                onClick={() => {
                  onOpenChange(false);
                  setManifestUrl("");
                  setSetAsDefault(false);
                  setResult(null);
                }}
              >
                Exit
              </Button>
            ) : (
              <>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => onOpenChange(false)}
                  disabled={importing}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={importing || !manifestUrl}>
                  {importing ? (
                    <>
                      <IconRefresh className="h-4 w-4 mr-2 animate-spin" />
                      Importing...
                    </>
                  ) : (
                    <>Import</>
                  )}
                </Button>
              </>
            )}
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
