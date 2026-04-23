import { StarIcon } from "@heroicons/react/20/solid";
import type React from "react";
import type { MapFilters } from "../hooks/useMaps";
import type { Player } from "../types/Player";

interface MapFiltersProps {
    filters: MapFilters;
    onFiltersChange: (filters: MapFilters) => void;
    players: Player[];
    playersLoading: boolean;
    playersError: string | null;
}

export function MapFilters({ filters, onFiltersChange, players, playersLoading, playersError }: MapFiltersProps) {
    const playersDisabled = playersLoading || playersError !== null || players.length === 0;

    const handleTierToggle = (tier: number) => {
        const newTiers = filters.tiers.includes(tier)
            ? filters.tiers.filter(t => t !== tier)
            : [...filters.tiers, tier];

        onFiltersChange({ ...filters, tiers: newTiers });
    };

    const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        onFiltersChange({ ...filters, search: e.target.value });
    }

    const handlePlayerToggle = (steamId: string) => {
        const playerIds = filters.playerIds.includes(steamId)
            ? filters.playerIds.filter((id) => id !== steamId)
            : [...filters.playerIds, steamId];

        const primaryPlayerId = playerIds.includes(filters.primaryPlayerId)
            ? filters.primaryPlayerId
            : playerIds[0] ?? '';

        onFiltersChange({ ...filters, playerIds, primaryPlayerId });
    };

    const handlePrimaryPlayerSelect = (steamId: string) => {
        if (!filters.playerIds.includes(steamId)) {
            return;
        }

        onFiltersChange({ ...filters, primaryPlayerId: steamId });
    };

    return (
        <div className="bg-gray-800/50 rounded-xl shadow-lg border border-gray-700 p-6 mb-8">
            <div className="grid grid-cols-1 lg:grid-cols-6 gap-6 items-start">
                <div className="lg:col-span-3">
                    <label className="block text-sm font-semibold text-gray-300 mb-3">
                        Filter by Tier
                    </label>
                    <div className="flex flex-nowrap gap-2 overflow-x-auto pb-1 pr-1">
                        {[1, 2, 3, 4, 5, 6, 7, 8].map((tier) => (
                            <label
                                key={tier}
                                className={`
                                    shrink-0 flex h-10 items-center gap-2 px-3 rounded-lg border-2 cursor-pointer transition-all
                                    ${filters.tiers.includes(tier)
                                        ? 'bg-indigo-500 border-indigo-400 text-white font-semibold shadow-lg shadow-indigo-500/20 hover:bg-indigo-400 hover:border-indigo-300 active:scale-95'
                                        : 'bg-gray-700/50 border-gray-600 text-gray-300 hover:border-gray-500 hover:bg-gray-600 active:scale-95'}
                                `}
                            >
                                <input
                                    type="checkbox"
                                    checked={filters.tiers.includes(tier)}
                                    onChange={() => handleTierToggle(tier)}
                                    className="hidden"
                                />
                                <span className="text-sm font-medium">Tier {tier}</span>
                            </label>
                        ))}
                    </div>
                </div>

                <div className="lg:col-span-2">
                    <label htmlFor="search" className="block text-sm font-semibold text-gray-300 mb-3">
                        Search Maps
                    </label>
                    <input
                        id="search"
                        type="text"
                        value={filters.search}
                        onChange={handleSearchChange}
                        placeholder="Type map name..."
                        className="h-10 w-full px-4 border-2 border-gray-600 rounded-lg bg-gray-700/50 text-white placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition"
                    />
                </div>

                <div>
                    <label className="block text-sm font-semibold text-gray-300 mb-3">
                        Map Type
                    </label>
                    <div className="inline-flex h-10 w-full rounded-lg border-2 border-gray-600 bg-gray-700/50 p-1">
                        {[
                            { value: 'all', label: 'All' },
                            { value: 'linear', label: 'Linear' },
                            { value: 'staged', label: 'Staged' },
                        ].map((option) => (
                            <button
                                key={option.value}
                                type="button"
                                onClick={() =>
                                    onFiltersChange({
                                        ...filters,
                                        linear: option.value as MapFilters['linear'],
                                    })
                                }
                                className={`h-full flex-1 rounded-md px-3 text-sm font-medium transition ${filters.linear === option.value
                                    ? 'bg-indigo-500 text-white shadow'
                                    : 'text-gray-300 hover:bg-gray-600/60'
                                    }`}
                            >
                                {option.label}
                            </button>
                        ))}
                    </div>
                </div>
            </div>

            <div className="mt-6 grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2">
                    <label className="block text-sm font-semibold text-gray-300 mb-3">
                        Players
                    </label>
                    {playersLoading && (
                        <p className="mb-3 text-sm text-gray-400">Loading players...</p>
                    )}
                    {playersError && (
                        <p className="mb-3 text-sm text-amber-300">Could not load players. Map browsing still works without player filters.</p>
                    )}
                    {!playersLoading && !playersError && players.length === 0 && (
                        <p className="mb-3 text-sm text-gray-400">No players have been ingested yet.</p>
                    )}
                    <div className="flex flex-wrap gap-2">
                        {players.map((player) => {
                            const isSelected = filters.playerIds.includes(player.steam_id);
                            const isPrimary = filters.primaryPlayerId === player.steam_id;
                            const label = player.name || player.steam_id;

                            if (!isSelected) {
                                return (
                                    <button
                                        key={player.steam_id}
                                        type="button"
                                        onClick={() => handlePlayerToggle(player.steam_id)}
                                        disabled={playersDisabled}
                                        className="rounded-lg border border-gray-600 bg-gray-700/50 px-3 py-2 text-sm font-medium text-gray-300 transition hover:border-gray-500 disabled:cursor-not-allowed disabled:border-gray-700 disabled:bg-gray-800 disabled:text-gray-500"
                                    >
                                        {label}
                                    </button>
                                );
                            }

                            return (
                                <span
                                    key={player.steam_id}
                                    className={`inline-flex overflow-hidden rounded-lg border text-sm font-medium transition ${isPrimary
                                        ? 'border-amber-300 bg-amber-400/15 text-amber-100'
                                        : 'border-cyan-400 bg-cyan-500/20 text-cyan-100'
                                        }`}
                                >
                                    <button
                                        type="button"
                                        onClick={() => handlePlayerToggle(player.steam_id)}
                                        disabled={playersDisabled}
                                        className="px-3 py-2 transition hover:bg-white/10 disabled:cursor-not-allowed disabled:text-gray-500"
                                    >
                                        {label}
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => handlePrimaryPlayerSelect(player.steam_id)}
                                        disabled={playersDisabled}
                                        aria-label={`Use ${label} as primary player`}
                                        title="Use as primary player"
                                        className={`border-l px-2 transition disabled:cursor-not-allowed ${isPrimary
                                            ? 'border-amber-200/30 bg-amber-300/20 text-amber-100'
                                            : 'border-cyan-200/20 text-cyan-100 hover:bg-white/10'
                                            }`}
                                    >
                                        <StarIcon className="size-4" />
                                    </button>
                                </span>
                            );
                        })}
                    </div>
                    <p className="mt-3 text-xs text-gray-400">
                        Click a name to select or remove it. Use the star on selected players to choose whose PB and group appear in the table.
                    </p>
                </div>

                <div>
                    <label className="block text-sm font-semibold text-gray-300 mb-3">
                        Completion
                    </label>
                    <div className="inline-flex h-10 w-full rounded-lg border-2 border-gray-600 bg-gray-700/50 p-1">
                        {[
                            { value: 'all', label: 'All Maps' },
                            { value: 'completed', label: 'All Comped' },
                            { value: 'incomplete', label: 'All Missed' },
                        ].map((option) => (
                            <button
                                key={option.value}
                                type="button"
                                onClick={() =>
                                    onFiltersChange({
                                        ...filters,
                                        completion: option.value as MapFilters['completion'],
                                    })
                                }
                                className={`h-full flex-1 rounded-md px-3 text-sm font-medium transition ${filters.completion === option.value
                                    ? 'bg-cyan-500 text-white shadow'
                                    : 'text-gray-300 hover:bg-gray-600/60'
                                    }`}
                            >
                                {option.label}
                            </button>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    )
}
