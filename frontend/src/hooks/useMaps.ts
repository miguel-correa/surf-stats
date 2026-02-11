import { useState, useEffect } from 'react';
import type { PaginatedMaps } from '../types/Map';

export type SortColumn = 'difficulty' | 'completions';

export interface MapFilters {
    tiers: number[];
    search: string;
    sort: SortColumn;
    order: 'asc' | 'desc';
}

export function useMaps(filters: MapFilters,
    options: { page?: number, perPage?: number } = {}): {
        paginatedData: PaginatedMaps | null,
        isLoading: boolean,
        error: string | null
    } {
    // const [maps, setMaps] = useState<SurfMap[]>([])
    const [paginatedData, setPaginatedData] = useState<PaginatedMaps | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        async function fetchMaps() {
            try {
                const params = new URLSearchParams();
                filters.tiers.forEach(t => params.append('tier', t.toString()));
                if (filters.search) params.append('search', filters.search);
                params.append('sort', filters.sort)
                params.append('order', filters.order)
                const {page = 1, perPage = 10} = options;
                params.append('page', page.toString())
                params.append('per_page', perPage.toString())

                const response = await fetch(`/api/v2/maps?${params}`)

                if (!response.ok) {
                    throw new Error('Failed to fetch maps');
                }

                const data: PaginatedMaps = await response.json();
                setPaginatedData(data);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'An error occurred');
            } finally {
                setIsLoading(false)
            }
        }

        fetchMaps();
    }, [filters.tiers, filters.search, filters.sort, filters.order, options.page, options.perPage])

    return { paginatedData, isLoading, error }
}