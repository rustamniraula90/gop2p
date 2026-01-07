import type { Upload } from '../../types';

interface Props {
    upload: Upload;
}

export default function UploadCard({ upload }: Readonly<Props>) {
    const progressPercent = (upload.progress * 100).toFixed(1);

    const statusColor = {
        complete: 'text-green-500',
        error: 'text-red-500',
        active: 'text-orange-500',
        preparing: 'text-yellow-500',
    }[upload.status] || 'text-gray-500';

    const statusBg = {
        complete: 'bg-green-500',
        error: 'bg-red-500',
        active: 'bg-orange-500',
        preparing: 'bg-yellow-500',
    }[upload.status] || 'bg-gray-500';

    return (
        <div className="border border-border-light dark:border-border-dark rounded-lg bg-white dark:bg-gray-800 p-4 shadow-sm">
            <div className="flex justify-between items-start mb-1">
                <h3 className="font-bold text-gray-900 dark:text-white truncate pr-4" title={upload.file_name}>
                    {upload.file_name || 'Unknown'}
                </h3>
                <span className={`text-xs font-semibold uppercase ${statusColor.replace('text-', 'text-')}`}>
                    {upload.status}
                </span>
            </div>

            <div className="text-sm text-gray-500 dark:text-gray-400 mb-4">
                To: {upload.peer_name || upload.peer_id || 'Unknown'}
            </div>

            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5 mb-2 overflow-hidden">
                <div
                    className={`h-full ${statusBg} rounded-full transition-all duration-500`}
                    style={{ width: `${progressPercent}%` }}
                />
            </div>

            <div className="flex justify-between items-center text-xs text-gray-500 dark:text-gray-400">
                <span>{progressPercent}% complete</span>
                <span>{upload.chunks_sent || 0} / {upload.chunk_count || 0} chunks</span>
            </div>
        </div>
    );
}
