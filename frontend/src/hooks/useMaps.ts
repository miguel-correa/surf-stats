import { useState, useEffect } from 'react';
import type { PaginatedMaps } from '../types/Map';

export type SortColumn = 'difficulty' | 'completions';
export type CompletionFilter = 'all' | 'completed' | 'incomplete';

export interface MapFilters {
    tiers: number[];
    search: string;
    linear: 'all' | 'linear' | 'staged';
    playerIds: string[];
    primaryPlayerId: string;
    completion: CompletionFilter;
    sort: SortColumn;
    order: 'asc' | 'desc';
}

export function useMaps(filters: MapFilters,
    options: { page?: number, perPage?: number } = {}): {
        paginatedData: PaginatedMaps | null,
        isLoading: boolean,
        error: string | null
    } {
    const { page = 1, perPage = 10 } = options;
    const [paginatedData, setPaginatedData] = useState<PaginatedMaps | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const controller = new AbortController();

        async function fetchMaps() {
            setIsLoading(true);
            setError(null);

            try {
                const params = new URLSearchParams();
                filters.tiers.forEach(t => params.append('tier', t.toString()));
                if (filters.search) params.append('search', filters.search);
                if (filters.linear === 'linear') params.append('linear', '1');
                if (filters.linear === 'staged') params.append('linear', '0');
                filters.playerIds.forEach((steamId) => params.append('steam_id', steamId));
                if (filters.primaryPlayerId) params.append('primary_steam_id', filters.primaryPlayerId);
                if (filters.completion !== 'all') params.append('completion_status', filters.completion);
                params.append('sort', filters.sort)
                params.append('order', filters.order)
                params.append('page', page.toString())
                params.append('per_page', perPage.toString())

                const response = await fetch(`/api/maps?${params}`, {
                    signal: controller.signal,
                })

                if (!response.ok) {
                    throw new Error('Failed to fetch maps');
                }

                const data: PaginatedMaps = await response.json();
                setPaginatedData(data);
            } catch (err) {
                if (controller.signal.aborted) return;
                setError(err instanceof Error ? err.message : 'An error occurred');
            } finally {
                if (!controller.signal.aborted) {
                    setIsLoading(false);
                }
            }
        }

        fetchMaps();

        return () => controller.abort();
    }, [filters.tiers, filters.search, filters.linear, filters.playerIds, filters.primaryPlayerId, filters.completion, filters.sort, filters.order, page, perPage])

    return { paginatedData, isLoading, error }
}
