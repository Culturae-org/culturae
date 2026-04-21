"use client";

import { CountryFlag } from "@/components/geography/country/country-flag";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useIsMobile } from "@/hooks/useMobile";
import { geographyService } from "@/lib/services/geography.service";
import type { Country } from "@/lib/types/geography.types";
import { IconDeviceFloppy, IconRefresh } from "@tabler/icons-react";
import * as React from "react";
import { toast } from "sonner";

interface CountryEditDialogProps {
  country: Country | null;
  datasetId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave?: (country: Country) => void;
}

export function CountryEditDialog({
  country,
  datasetId,
  open,
  onOpenChange,
  onSave,
}: CountryEditDialogProps) {
  const isMobile = useIsMobile();
  const [saving, setSaving] = React.useState(false);
  const [formData, setFormData] = React.useState<Partial<Country>>({});

  React.useEffect(() => {
    if (country) {
      setFormData({
        ...country,
      });
    }
  }, [country]);

  if (!country) return null;

  const name = (formData.name || country.name) as Record<string, string>;
  const capital = (formData.capital || country.capital) as Record<
    string,
    string
  >;
  const currency = formData.currency ?? country.currency;

  const updateName = (lang: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      name: { ...name, [lang]: value },
    }));
  };

  const updateCapital = (lang: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      capital: { ...capital, [lang]: value },
    }));
  };

  const updateCurrencyName = (lang: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      currency: {
        ...currency,
        name: {
          ...currency?.name,
          [lang]: value,
        },
      },
    }));
  };

  const updateCurrencyCode = (value: string) => {
    setFormData((prev) => ({
      ...prev,
      currency: {
        ...currency,
        code: value,
      },
    }));
  };

  const updateCurrencySymbol = (value: string) => {
    setFormData((prev) => ({
      ...prev,
      currency: {
        ...currency,
        symbol: value,
      },
    }));
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const updatedCountry = await geographyService.updateCountry(
        datasetId,
        country.slug,
        formData,
      );
      toast.success("Country updated successfully");
      onSave?.(updatedCountry);
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to save country:", error);
      toast.error("Failed to update country");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Drawer
      direction={isMobile ? "bottom" : "right"}
      open={open}
      onOpenChange={onOpenChange}
    >
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle className="flex items-center gap-3">
            <CountryFlag countryCode={country.slug} size="lg" />
            <span>Edit: {name?.en || country.slug}</span>
          </DrawerTitle>
          <DrawerDescription>Modify country information</DrawerDescription>
        </DrawerHeader>

        <ScrollArea className="max-h-[70vh] px-4">
          <Tabs defaultValue="names" className="w-full">
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="names">Names</TabsTrigger>
              <TabsTrigger value="location">Location</TabsTrigger>
              <TabsTrigger value="info">Info</TabsTrigger>
              <TabsTrigger value="currency">Currency</TabsTrigger>
            </TabsList>

            <TabsContent value="names" className="space-y-4 mt-4">
              <div className="space-y-4">
                <h4 className="text-sm font-semibold">Country Names</h4>
                <div className="grid grid-cols-2 gap-4">
                  {Object.entries(name || {}).map(([lang, value]) => (
                    <div key={lang} className="space-y-2">
                      <Label className="uppercase text-xs">{lang}</Label>
                      <Input
                        value={value}
                        onChange={(e) => updateName(lang, e.target.value)}
                      />
                    </div>
                  ))}
                </div>
              </div>

              <div className="space-y-4">
                <h4 className="text-sm font-semibold">Capital Names</h4>
                <div className="grid grid-cols-2 gap-4">
                  {Object.entries(capital || {}).map(([lang, value]) => (
                    <div key={lang} className="space-y-2">
                      <Label className="uppercase text-xs">{lang}</Label>
                      <Input
                        value={value}
                        onChange={(e) => updateCapital(lang, e.target.value)}
                      />
                    </div>
                  ))}
                </div>
              </div>
            </TabsContent>

            <TabsContent value="location" className="space-y-4 mt-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Continent</Label>
                  <Input
                    value={formData.continent ?? country.continent}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        continent: e.target.value,
                      }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Region</Label>
                  <Input
                    value={formData.region ?? country.region ?? ""}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        region: e.target.value,
                      }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Latitude</Label>
                  <Input
                    type="number"
                    step="0.0001"
                    value={
                      (formData.coordinates?.lat ?? country.coordinates?.lat) ||
                      ""
                    }
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        coordinates: {
                          lat: Number.parseFloat(e.target.value) || 0,
                          lng:
                            prev.coordinates?.lng ??
                            country.coordinates?.lng ??
                            0,
                        },
                      }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Longitude</Label>
                  <Input
                    type="number"
                    step="0.0001"
                    value={
                      (formData.coordinates?.lng ?? country.coordinates?.lng) ||
                      ""
                    }
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        coordinates: {
                          lat:
                            prev.coordinates?.lat ??
                            country.coordinates?.lat ??
                            0,
                          lng: Number.parseFloat(e.target.value) || 0,
                        },
                      }))
                    }
                  />
                </div>
              </div>
            </TabsContent>

            <TabsContent value="info" className="space-y-4 mt-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Population</Label>
                  <Input
                    type="number"
                    value={formData.population ?? country.population}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        population: Number.parseInt(e.target.value, 10),
                      }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Area (km²)</Label>
                  <Input
                    type="number"
                    value={formData.area_km2 ?? country.area_km2}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        area_km2: Number.parseInt(e.target.value, 10),
                      }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Phone Code</Label>
                  <Input
                    value={formData.phone_code ?? country.phone_code ?? ""}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        phone_code: e.target.value,
                      }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>TLD</Label>
                  <Input
                    value={formData.tld ?? country.tld ?? ""}
                    onChange={(e) =>
                      setFormData((prev) => ({ ...prev, tld: e.target.value }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Driving Side</Label>
                  <Input
                    value={formData.driving_side ?? country.driving_side ?? ""}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        driving_side: e.target.value,
                      }))
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Flag Emoji</Label>
                  <Input
                    value={formData.flag ?? country.flag ?? ""}
                    onChange={(e) =>
                      setFormData((prev) => ({ ...prev, flag: e.target.value }))
                    }
                  />
                </div>
              </div>
            </TabsContent>

            <TabsContent value="currency" className="space-y-4 mt-4">
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>Currency Code</Label>
                  <Input
                    value={currency?.code || ""}
                    onChange={(e) => updateCurrencyCode(e.target.value)}
                  />
                </div>
                <div className="space-y-4">
                  <h4 className="text-sm font-semibold">
                    Currency Names (EN / FR / ES)
                  </h4>
                  <div className="grid grid-cols-3 gap-4">
                    <div className="space-y-2">
                      <Label className="uppercase text-xs">English</Label>
                      <Input
                        value={currency?.name?.en || ""}
                        onChange={(e) =>
                          updateCurrencyName("en", e.target.value)
                        }
                      />
                    </div>
                    <div className="space-y-2">
                      <Label className="uppercase text-xs">Français</Label>
                      <Input
                        value={currency?.name?.fr || ""}
                        onChange={(e) =>
                          updateCurrencyName("fr", e.target.value)
                        }
                      />
                    </div>
                    <div className="space-y-2">
                      <Label className="uppercase text-xs">Español</Label>
                      <Input
                        value={currency?.name?.es || ""}
                        onChange={(e) =>
                          updateCurrencyName("es", e.target.value)
                        }
                      />
                    </div>
                  </div>
                </div>
                <div className="space-y-4">
                  <h4 className="text-sm font-semibold">Currency Symbol</h4>
                  <Input
                    value={currency?.symbol || ""}
                    onChange={(e) => updateCurrencySymbol(e.target.value)}
                  />
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </ScrollArea>

        <DrawerFooter className="flex flex-row gap-2">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving ? (
              <>
                <IconRefresh className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <IconDeviceFloppy className="mr-2 h-4 w-4" />
                Save Changes
              </>
            )}
          </Button>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
