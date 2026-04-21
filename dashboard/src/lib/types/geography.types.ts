export interface GeographyDataset {
  id: string;
  slug: string;
  name: string;
  description: string;
  version: string;
  manifest_url?: string;
  manifest_data?: GeographyManifest;
  source: "cultpedia" | "custom" | "imported";
  import_job_id?: string;
  import_job?: ImportJob;
  imported_at: string;
  country_count: number;
  continent_count: number;
  region_count: number;
  flag_count: number;
  flag_png512_count: number;
  flag_png1024_count: number;
  is_active: boolean;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface ImportJob {
  id: string;
  manifest_url: string;
  dataset: string;
  version: string;
  started_at: string;
  finished_at?: string;
  success: boolean;
  added: number;
  updated: number;
  skipped: number;
  errors: number;
  message?: string;
}

export interface GeographyManifest {
  schema_version: string;
  dataset: string;
  version: string;
  type: string;
  license?: string;
  created_at: string;
  updated_at: string;
  sources?: Array<{
    name: string;
    url: string;
    license: string;
  }>;
  includes: string[];
  assets?: Record<string, string>;
  counts: {
    countries: number;
    continents: number;
    regions: number;
    flags: number;
  };
  checksums?: Record<string, string>;
}

export interface Country {
  id: string;
  dataset_id: string;
  slug: string;
  iso_alpha2: string;
  iso_alpha3: string;
  iso_numeric: string;
  name: Record<string, string>;
  official_name: Record<string, string>;
  capital: Record<string, string>;
  continent: string;
  region: string;
  coordinates: { lat: number; lng: number };
  flag: string;
  population: number;
  area_km2: number;
  currency: {
    code: string;
    name: Record<string, string>;
    symbol: string;
  };
  languages: string[];
  neighbors: string[];
  tld: string;
  phone_code: string;
  driving_side: string;
  independent: boolean;
  created_at: string;
  updated_at: string;
}

export interface Continent {
  id: string;
  dataset_id: string;
  slug: string;
  name: Record<string, string>;
  countries: string[];
  area_km2: number;
  population: number;
  created_at: string;
  updated_at: string;
}

export interface Region {
  id: string;
  dataset_id: string;
  slug: string;
  name: Record<string, string>;
  continent: string;
  countries: string[];
  created_at: string;
  updated_at: string;
}

export interface GeographyImportResult {
  success: boolean;
  message: string;
  countries_added: number;
  countries_updated: number;
  countries_skipped: number;
  continents_added: number;
  continents_updated: number;
  regions_added: number;
  regions_updated: number;
  flags_added: number;
  flags_png512_added: number;
  flags_png1024_added: number;
  errors: string[];
}

export interface GeographyDatasetStats {
  dataset_id: string;
  name: string;
  version: string;
  country_count: number;
  continent_count: number;
  region_count: number;
  flag_count: number;
  flag_png512_count: number;
  flag_png1024_count: number;
  is_active: boolean;
  is_default: boolean;
  imported_at: string;
}

export interface DatasetUpdateInfo {
  has_update: boolean;
  current_version: string;
  latest_version: string;
  updated_at?: string;
  manifest?: GeographyManifest;
  name?: string;
  dataset_id?: string;
}
