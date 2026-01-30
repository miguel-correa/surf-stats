import { useState, useEffect } from 'react';
import type { SurfMap } from '../types/Map';

export type SortColumn = 'difficulty' | 'completions';

export interface MapFilters {
    tiers: number[];
    search: string;
    sort: SortColumn;
    order: 'asc' | 'desc';
}

export function useMaps(filters: MapFilters): {
    maps: SurfMap[],
    isLoading: boolean,
    error: string | null
} {
    const [maps, setMaps] = useState<SurfMap[]>([])
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

                const response = await fetch(`http://localhost:8080/api/maps?${params}`)
            
                if(!response.ok) {
                    throw new Error('Failed to fetch maps');
                }

                const data: SurfMap[] = await response.json();
                setMaps(data);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'An error occurred');
            } finally {
                setIsLoading(false)
            }
        }

        fetchMaps();
    }, [filters.tiers, filters.search, filters.sort, filters.order])

    return { maps, isLoading, error }
}