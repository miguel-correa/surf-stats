export interface SurfMap {
    id: number;
    name: string;
    tier: number;
    year: number;
    completions: number;
    hours_played: number;
    comp_per_hour: number;
    notes: string;
};

export interface PaginatedMaps {
    maps: SurfMap[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}