import type React from "react";
import type { MapFilters } from "../hooks/useMaps";

interface MapFiltersProps {
    filters: MapFilters;
    onFiltersChange: (filters: MapFilters) => void;
}

export function MapFilters({ filters, onFiltersChange }: MapFiltersProps) {
    const handleTierToggle = (tier: number) => {
        const newTiers = filters.tiers.includes(tier)
            ? filters.tiers.filter(t => t !== tier)
            : [...filters.tiers, tier];

        onFiltersChange({ ...filters, tiers: newTiers });
    };

    const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        onFiltersChange({ ...filters, search: e.target.value });
    }

    return (
        <div className="bg-white rounded-xl shadow-lg border border-gray-200 p-6 mb-8">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
                <div className="lg:col-span-2">
                    <label className="block text-sm font-semibold text-gray-800 mb-3">
                        Filter by Tier
                    </label>
                    <div className="flex flex-wrap gap-2">
                        {[1, 2, 3, 4, 5, 6, 7, 8].map((tier) => (
                            <label
                                key={tier}
                                className={`
                                    flex items-center gap-2 px-4 py-2 rounded-lg border-2 cursor-pointer transition-all
                                    ${filters.tiers.includes(tier)
                                        ? 'bg-blue-500 border-blue-500 text-white font-medium'
                                        : 'bg-white border-gray-300 text-gray-700 hover:border-blue-300 hover:bg-blue-50'
                                    }
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

                <div>
                    <label htmlFor="search" className="block text-sm font-semibold text-gray-800 mb-3">
                        Search Maps
                    </label>
                    <input
                        id="search"
                        type="text"
                        value={filters.search}
                        onChange={handleSearchChange}
                        placeholder="Type map name..."
                        className="w-full px-4 py-2.5 border-2 border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition"
                    />
                </div>

            </div>
        </div>
    )
}