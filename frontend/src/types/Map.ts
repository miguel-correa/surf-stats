export interface PlayerMapSummary {
    steam_id: string;
    best_time_ms?: number | null;
    rank?: number | null;
    group_tier?: number | null;
    completed: boolean;
}

export interface SurfMap {
    id: number;
    ksf_map_id: number;
    name: string;
    tier: number;
    added: number;
    completions: number;
    playtime_seconds: number;
    comp_per_hour: number;
    linear: number;
    notes: string;
    player_best_time_ms?: number | null;
    player_rank?: number | null;
    completed?: boolean | null;
    player_records?: PlayerMapSummary[];
}

export interface PaginatedMaps {
    maps: SurfMap[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}
