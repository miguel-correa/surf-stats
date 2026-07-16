import { StarIcon, XMarkIcon } from "@heroicons/react/20/solid";
import type React from "react";
import type { MapFilters, PrimaryGroupFilter } from "../hooks/useMaps";
import type { Player } from "../types/Player";

interface MapFiltersProps {
    filters: MapFilters;
    onFiltersChange: (filters: MapFilters) => void;
    searchText: string;
    onSearchChange: (value: string) => void;
    players: Player[];
    playersLoading: boolean;
    playersError: string | null;
}

const tierOptions = [1, 2, 3, 4, 5, 6, 7, 8];

const primaryGroupOptions: { value: PrimaryGroupFilter; label: string }[] = [
    { value: 'ungrouped', label: 'Ungrouped' },
    { value: '6', label: 'G6' },
    { value: '5', label: 'G5' },
    { value: '4', label: 'G4' },
    { value: '3', label: 'G3' },
    { value: '2', label: 'G2' },
    { value: '1', label: 'G1' },
    { value: '0', label: 'G0' },
];

export function MapFilters({ filters, onFiltersChange, searchText, onSearchChange, players, playersLoading, playersError }: MapFiltersProps) {
    const playersDisabled = playersLoading || playersError !== null || players.length === 0;
    const primaryGroupDisabled = filters.primaryPlayerId === '';
    const primaryPlayer = players.find((player) => player.steam_id === filters.primaryPlayerId);
    const primaryPlayerLabel = primaryPlayer ? (primaryPlayer.name || primaryPlayer.steam_id) : 'No primary player';
    const activeFilterCount =
        filters.tiers.length +
        filters.primaryGroups.length +
        filters.playerIds.length +
        (filters.linear === 'all' ? 0 : 1) +
        (filters.completion === 'all' ? 0 : 1) +
        (searchText.trim() === '' ? 0 : 1);

    const sectionLabelClass = "text-xs font-semibold uppercase tracking-wider text-gray-400";
    const inactivePillClass = "border-gray-600 bg-gray-800/70 text-gray-300 hover:border-gray-500 hover:bg-gray-700";
    const disabledPillClass = "disabled:cursor-not-allowed disabled:border-gray-700 disabled:bg-gray-800/60 disabled:text-gray-500 disabled:shadow-none";

    const handleTierToggle = (tier: number) => {
        const tiers = filters.tiers.includes(tier)
            ? filters.tiers.filter(t => t !== tier)
            : [...filters.tiers, tier];

        onFiltersChange({ ...filters, tiers });
    };

    const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        onSearchChange(e.target.value);
    };

    const handlePlayerToggle = (steamId: string) => {
        const playerIds = filters.playerIds.includes(steamId)
            ? filters.playerIds.filter((id) => id !== steamId)
            : [...filters.playerIds, steamId];

        const primaryPlayerId = playerIds.includes(filters.primaryPlayerId)
            ? filters.primaryPlayerId
            : playerIds[0] ?? '';

        const primaryGroups = primaryPlayerId === '' ? [] : filters.primaryGroups;
        onFiltersChange({ ...filters, playerIds, primaryPlayerId, primaryGroups });
    };

    const handlePrimaryPlayerSelect = (steamId: string) => {
        if (!filters.playerIds.includes(steamId)) {
            return;
        }

        onFiltersChange({ ...filters, primaryPlayerId: steamId });
    };

    const handlePrimaryGroupToggle = (group: PrimaryGroupFilter) => {
        if (primaryGroupDisabled) {
            return;
        }

        const primaryGroups = filters.primaryGroups.includes(group)
            ? filters.primaryGroups.filter((value) => value !== group)
            : [...filters.primaryGroups, group];

        onFiltersChange({ ...filters, primaryGroups });
    };

    const handleClearFilters = () => {
        onSearchChange('');
        onFiltersChange({
            ...filters,
            tiers: [],
            search: '',
            linear: 'all',
            playerIds: [],
            primaryPlayerId: '',
            primaryGroups: [],
            completion: 'all',
        });
    };

    return (
        <div className="mb-8 rounded-lg border border-gray-700 bg-gray-900/70 shadow-lg">
            <div className="flex flex-col gap-4 border-b border-gray-700 px-5 py-4 lg:flex-row lg:items-end">
                <div className="min-w-0 flex-1">
                    <label htmlFor="search" className={sectionLabelClass}>
                        Search
                    </label>
                    <input
                        id="search"
                        type="text"
                        value={searchText}
                        onChange={handleSearchChange}
                        placeholder="Map name"
                        className="mt-2 h-10 w-full rounded-md border border-gray-600 bg-gray-800/80 px-3 text-sm text-white placeholder:text-gray-500 outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/40"
                    />
                </div>

                <div className="w-full lg:w-56">
                    <span className={sectionLabelClass}>Map Type</span>
                    <div className="mt-2 grid h-10 grid-cols-3 rounded-md border border-gray-600 bg-gray-800/80 p-1">
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
                                className={`rounded px-2 text-sm font-medium transition ${filters.linear === option.value
                                    ? 'bg-indigo-500 text-white'
                                    : 'text-gray-300 hover:bg-gray-700'
                                    }`}
                            >
                                {option.label}
                            </button>
                        ))}
                    </div>
                </div>

                <div className="w-full lg:w-72">
                    <span className={sectionLabelClass}>Completion</span>
                    <div className="mt-2 grid h-10 grid-cols-3 rounded-md border border-gray-600 bg-gray-800/80 p-1">
                        {[
                            { value: 'all', label: 'All' },
                            { value: 'completed', label: 'Comped' },
                            { value: 'incomplete', label: 'Missed' },
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
                                className={`rounded px-2 text-sm font-medium transition ${filters.completion === option.value
                                    ? 'bg-cyan-500 text-white'
                                    : 'text-gray-300 hover:bg-gray-700'
                                    }`}
                            >
                                {option.label}
                            </button>
                        ))}
                    </div>
                </div>

                <div className="flex items-center justify-between gap-3 lg:w-36 lg:justify-end">
                    <span className="text-sm text-gray-400">{activeFilterCount} active</span>
                    <button
                        type="button"
                        onClick={handleClearFilters}
                        disabled={activeFilterCount === 0}
                        className="inline-flex h-10 items-center gap-1.5 rounded-md border border-gray-600 px-3 text-sm font-medium text-gray-300 transition hover:border-gray-500 hover:bg-gray-800 disabled:cursor-not-allowed disabled:border-gray-700 disabled:text-gray-600 disabled:hover:bg-transparent"
                    >
                        <XMarkIcon className="size-4" />
                        Clear
                    </button>
                </div>
            </div>

            <div className="grid gap-5 px-5 py-4 lg:grid-cols-[minmax(0,1fr)_minmax(340px,0.85fr)]">
                <section>
                    <div className="mb-3 flex items-center justify-between gap-3">
                        <span className={sectionLabelClass}>Map Tier</span>
                    </div>
                    <div className="flex flex-wrap gap-2">
                        {tierOptions.map((tier) => {
                            const isSelected = filters.tiers.includes(tier);

                            return (
                                <button
                                    key={tier}
                                    type="button"
                                    onClick={() => handleTierToggle(tier)}
                                    className={`h-9 rounded-md border px-3 text-sm font-medium transition ${isSelected
                                        ? 'border-indigo-300 bg-indigo-500/80 text-white'
                                        : inactivePillClass
                                        }`}
                                >
                                    T{tier}
                                </button>
                            );
                        })}
                    </div>
                </section>

                <section className="border-t border-gray-800 pt-4 lg:border-l lg:border-t-0 lg:pl-5 lg:pt-0">
                    <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
                        <span className={sectionLabelClass}>Primary Player</span>
                        <span className="truncate text-sm font-medium text-amber-100">{primaryPlayerLabel}</span>
                    </div>
                    <div className="flex flex-wrap gap-2">
                        {primaryGroupOptions.map((option) => {
                            const isSelected = filters.primaryGroups.includes(option.value);

                            return (
                                <button
                                    key={option.value}
                                    type="button"
                                    onClick={() => handlePrimaryGroupToggle(option.value)}
                                    disabled={primaryGroupDisabled}
                                    className={`h-9 rounded-md border px-3 text-sm font-medium transition ${isSelected
                                        ? 'border-amber-300 bg-amber-400/20 text-amber-100'
                                        : inactivePillClass
                                        } ${disabledPillClass}`}
                                >
                                    {option.label}
                                </button>
                            );
                        })}
                    </div>
                </section>
            </div>

            <div className="border-t border-gray-700 px-5 py-4">
                <div className="mb-3 flex items-center justify-between gap-3">
                    <span className={sectionLabelClass}>Players</span>
                    {playersLoading && <span className="text-sm text-gray-400">Loading</span>}
                    {playersError && <span className="text-sm text-amber-300">Unavailable</span>}
                    {!playersLoading && !playersError && players.length === 0 && <span className="text-sm text-gray-400">None ingested</span>}
                </div>

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
                                    className={`h-9 rounded-md border px-3 text-sm font-medium transition ${inactivePillClass} ${disabledPillClass}`}
                                >
                                    {label}
                                </button>
                            );
                        }

                        return (
                            <span
                                key={player.steam_id}
                                className={`inline-flex h-9 overflow-hidden rounded-md border text-sm font-medium transition ${isPrimary
                                    ? 'border-amber-300 bg-amber-400/15 text-amber-100'
                                    : 'border-cyan-400 bg-cyan-500/15 text-cyan-100'
                                    }`}
                            >
                                <button
                                    type="button"
                                    onClick={() => handlePlayerToggle(player.steam_id)}
                                    disabled={playersDisabled}
                                    className="px-3 transition hover:bg-white/10 disabled:cursor-not-allowed disabled:text-gray-500"
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
                                        ? 'border-amber-200/30 bg-amber-300/15 text-amber-100'
                                        : 'border-cyan-200/20 text-cyan-100 hover:bg-white/10'
                                        }`}
                                >
                                    <StarIcon className="size-4" />
                                </button>
                            </span>
                        );
                    })}
                </div>
            </div>
        </div>
    );
}
