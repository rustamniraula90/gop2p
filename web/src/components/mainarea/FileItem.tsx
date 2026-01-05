import type {FileInfo} from "../../types";
import {useState} from "react";
import {formatFileSize} from "../../utils/formatters.ts";
import {useApp} from "../../context/AppContext.tsx";

interface Props {
    file: FileInfo
    peerId: string
}

export default function FileItem({file, peerId}: Readonly<Props>) {
    const {dispatch} = useApp()
    const [loading, setLoading] = useState(false)

    const downloadFile = async () => {
        if (!peerId || !file.name) {
            alert(`Cannot start download: peer ID (${peerId}) or file name (${file.name}) is missing.`);
            return
        }
        setLoading(true)
        try {
            const res = await fetch('/api/download/start', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({peer_id: peerId, file_name: file.name})
            });

            if (res.ok) {
                dispatch({type: 'SET_VIEW', payload: 'downloads'})
            } else {
                const text = await res.text();
                alert(`Failed to start download: ${text}`);
            }
        } catch (e) {
            console.error('Failed to start download', e);
            alert('Failed to start download. Check console for details.');
        } finally {
            setLoading(false)
        }
    }

    const icon = file.type === 'dir' ? 'folder' : 'description';
    const iconColor = file.type === 'dir' ? 'text-yellow-500' : 'text-blue-400';
    return (
        <tr className="group hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors border-b border-border-light dark:border-border-dark last:border-0">
            <td className="px-6 py-4">
                <div className="flex items-center space-x-3">
                    <span className={`material-symbols-outlined text-sm ${iconColor}`}>{icon}</span>
                    <span
                        className={`font-medium transition-colors ${file.type === 'dir' ? 'text-gray-700 dark:text-gray-200 group-hover:text-primary' : 'text-gray-700 dark:text-gray-200'}`}>
                        {file.name}
                    </span>
                </div>
            </td>
            <td className="px-6 py-4 text-gray-500 dark:text-gray-400 font-mono">
                {file.type === 'file' ? formatFileSize(file.size) : '-'}
            </td>
            <td className="px-6 py-4 text-gray-500 dark:text-gray-400">{file.type}</td>
            <td className="px-6 py-4 text-right">
                {file.type === 'file' && (
                    <button
                        onClick={downloadFile}
                        disabled={loading}
                        className={`inline-flex items-center space-x-1 text-white text-xs px-3 py-1.5 rounded transition-colors shadow-sm ${loading ? 'bg-gray-400 cursor-default' : 'bg-primary hover:bg-blue-600'}`}
                    >
                        <span className="material-symbols-outlined text-xs">
                            {loading ? 'hourglass_empty' : 'arrow_downward'}
                        </span>
                        <span>{loading ? 'Starting...' : 'Download'}</span>
                    </button>
                )}
            </td>
        </tr>
    )
}