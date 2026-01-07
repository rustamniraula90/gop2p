import type {Download} from "../../types";
import ChunkGrid from "./ChunkGrid.tsx";

interface Props {
    download: Download;
}

export default function DownloadCard({download}: Readonly<Props>) {
    const progressPercent = (download.progress * 100).toFixed(1);

    const statusColor = {
        complete: 'text-green-500',
        error: 'text-red-500',
        active: 'text-blue-500',
        pending: 'text-yellow-500',
        requesting: 'text-blue-500',
    }[download.status] || 'text-gray-500';

    const statusBg = {
        complete: 'bg-green-500',
        error: 'bg-red-500',
        active: 'bg-blue-500',
        pending: 'bg-yellow-500',
        requesting: 'bg-blue-500',
    }[download.status] || 'bg-gray-500';

    return (
        <div
            className="border border-border-light dark:border-border-dark rounded-lg bg-white dark:bg-gray-800 p-4 shadow-sm">
            <div className="flex justify-between items-start mb-1">
                <h3 className="font-bold text-gray-900 dark:text-white truncate pr-4" title={download.file_name}>
                    {download.file_name || 'Unknown'}
                </h3>
                <span className={`text-xs font-semibold uppercase ${statusColor.replace('text-', 'text-')}`}>
                    {download.status}
                </span>
            </div>

            <div className="text-sm text-gray-500 dark:text-gray-400 mb-4">
                From: {download.peer_name || download.peer_id || 'Unknown'}
            </div>

            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5 mb-2 overflow-hidden">
                <div
                    className={`h-full ${statusBg} rounded-full transition-all duration-500`}
                    style={{width: `${progressPercent}%`}}
                />
            </div>

            <div className="flex justify-between items-center mb-4 text-xs text-gray-500 dark:text-gray-400">
                <span>{progressPercent}% complete</span>
            </div>

            {download.chunks && download.chunks.length > 0 && (
                <details className="group" open={false}>
                    <summary
                        className="list-none cursor-pointer mb-2 flex items-center text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200">
                        <span
                            className="material-symbols-outlined text-sm mr-1 transition-transform group-open:rotate-90">chevron_right</span>
                        Chunk Map
                    </summary>
                    <div className="bg-gray-900 rounded p-1 border border-gray-700 overflow-hidden">
                        <ChunkGrid chunks={download.chunks}/>
                    </div>
                </details>
            )}
        </div>
    );
}