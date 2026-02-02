interface PaginationProp {
    currentPage: number;
    totalPages: number;
    onPageChange: (page: number) => void;
}

export function Pagination({ currentPage, totalPages, onPageChange }: PaginationProp) {
    if (totalPages <= 1) return null;

    const isFirstPage = currentPage === 1;
    const isLastPage = currentPage >= totalPages;

    return (
        <div className="flex justify-center items-center gap-4 mt-8 mb-6">

            {/* Previous button */}
            <button
                onClick={() => onPageChange(currentPage - 1)}
                disabled={isFirstPage}
                className="px-6 py-2 rounded-lg font-medium transition-all
                            bg-blue-600 text-white hover:bg-blue-700
                            disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-blue-600
                            shadow-md hover:shadow-lg">
                Previous
            </button>

            {/* Page indicator */}
            <span className="px-4 py-2 text-gray-700 font-semibold text-sm bg-white rounded-lg border border-gray-200 shadow-sm">
                Page {currentPage} of {totalPages}
            </span>

            {/* Next button */}
            <button
                onClick={() => onPageChange(currentPage + 1)}
                disabled={isLastPage}
                className="px-6 py-2 rounded-lg font-medium transition-all
                            bg-blue-600 text-white hover:bg-blue-700
                            disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-blue-600
                            shadow-md hover:shadow-lg">
                Next
            </button>

        </div>
    );
}