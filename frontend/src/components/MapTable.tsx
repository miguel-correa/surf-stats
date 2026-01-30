import type { SurfMap } from "../types/Map";
import type { SortColumn } from "../hooks/useMaps";

interface MapTableProps {
    maps: SurfMap[];
    sortCol: SortColumn;
    sortOrder: 'asc' | 'desc';
    onSort: (column: SortColumn) => void;
}

export function MapTable({ maps, sortCol, sortOrder, onSort }: MapTableProps) {
    const SortableHeader = ({ column, label }: { column: SortColumn, label: string }) => (
        <th
            onClick={() => onSort(column)}
            className="px-6 py-4 text-left text-xs font-bold text-gray-700 uppercase tracking-wider cursor-pointer
  hover:bg-gray-200 transition"
        >
            {label}
            {sortCol === column && (
                <span className="ml-2">{sortOrder === 'asc' ? '↑' : '↓'}</span>
            )}
        </th>
    );


    return (
        <div className="overflow-x-auto bg-white rounded-xl shadow-lg border border-gray-200">
            <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gradient-to-r from-gray-50 to-gray-100">
                    <tr>
                        <th className="px-6 py-4 text-left text-xs font-bold text-gray-700 uppercase tracking-wider">Map Name</th>
                        <th className="px-6 py-4 text-left text-xs font-bold text-gray-700 uppercase tracking-wider">Tier</th>
                        <th className="px-6 py-4 text-left text-xs font-bold text-gray-700 uppercase tracking-wider">Year</th>
                        <SortableHeader column="completions" label="Completions"/>
                        <th className="px-6 py-4 text-left text-xs font-bold text-gray-700 uppercase tracking-wider">Hours Played</th>
                        <SortableHeader column="difficulty" label="Comp/Hour"/>
                    </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                    {maps.map((map, index) => (
                        <tr
                            key={map.id}
                            className={`
                                transition-colors
                                ${index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}
                                hover:bg-blue-50
                            `}
                        >
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-gray-900">
                                {map.name}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                    Tier {map.tier}
                                </span>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                                {map.year}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 font-medium">
                                {map.completions.toLocaleString()}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                                {map.hours_played.toFixed(1)}h
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-gray-900">
                                {map.comp_per_hour.toFixed(2)}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    )
}