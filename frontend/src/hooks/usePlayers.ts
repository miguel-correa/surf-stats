import { useEffect, useState } from 'react';
import type { Player } from '../types/Player';

export function usePlayers(): {
    players: Player[];
    isLoading: boolean;
    error: string | null;
} {
    const [players, setPlayers] = useState<Player[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        let cancelled = false;

        async function fetchPlayers() {
            setIsLoading(true);
            setError(null);

            try {
                const response = await fetch('/api/players');
                if (!response.ok) {
                    throw new Error('Failed to fetch players');
                }

                const data: Player[] = await response.json();
                if (!cancelled) {
                    setPlayers(data);
                }
            } catch (err) {
                if (!cancelled) {
                    setError(err instanceof Error ? err.message : 'An error occurred');
                }
            } finally {
                if (!cancelled) {
                    setIsLoading(false);
                }
            }
        }

        fetchPlayers();

        return () => {
            cancelled = true;
        };
    }, []);

    return { players, isLoading, error };
}
