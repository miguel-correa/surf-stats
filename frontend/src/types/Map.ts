export interface SurfMap {
    id: number;
    name: string;
    tier: number;
    added: number;
    completions: number;
    playtime_seconds: number;
    comp_per_hour: number;
    linear: number;
    notes: string;
};

export interface PaginatedMaps {
    maps: SurfMap[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}
