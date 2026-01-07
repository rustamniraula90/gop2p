import {useApp} from "../../context/AppContext.tsx";
import {useEffect} from "react";
import DownloadCard from "./DownloadCard.tsx";

export default function DownloadsView() {
    const {state, dispatch} = useApp();

    useEffect(() => {
        // Fetch downloads initially and periodically
        const fetchDownloads = async () => {
            try {
                const res = await fetch('/api/downloads');
                const data = await res.json();
                dispatch({type: 'SET_DOWNLOADS', payload: data || []});
            } catch (e) {
                console.error('Failed to fetch downloads', e);
            }
        };

        fetchDownloads();
        const interval = setInterval(fetchDownloads, 2000);
        return () => clearInterval(interval);
    }, [dispatch]);

    const downloads = Object.values(state.downloads);


    const clearCompleted = async () => {
        await fetch('/api/downloads/clear', {method: 'POST'});
        // Refresh
        const res = await fetch('/api/downloads');
        const data = await res.json();
        dispatch({type: 'SET_DOWNLOADS', payload: data || []});
    };

    return (
        <div className="h-full flex flex-col">
            <div className="p-4 border-b border-border-light dark:border-border-dark flex justify-between items-center bg-panel-light dark:bg-panel-dark">
                <h2 className="text-xl font-bold text-gray-900 dark:text-white">Active Downloads</h2>
                <div className="flex space-x-2">
                    {downloads.some(d => d.status === 'complete' || d.status === 'error') && (
                        <button
                            onClick={clearCompleted}
                            className="flex items-center space-x-1 text-sm text-gray-500 dark:text-gray-400 hover:text-red-500 dark:hover:text-red-400 transition-colors border border-border-light dark:border-border-dark px-3 py-1.5 rounded bg-gray-50 dark:bg-gray-800"
                        >
                            <span className="material-symbols-outlined text-base">delete_sweep</span>
                            <span>Clear Finished</span>
                        </button>
                    )}
                    <button
                        onClick={() => fetch('/api/downloads').then(r => r.json()).then(d => dispatch({ type: 'SET_DOWNLOADS', payload: d || [] }))}
                        className="flex items-center space-x-1 text-sm text-gray-500 dark:text-gray-400 hover:text-primary dark:hover:text-primary transition-colors border border-border-light dark:border-border-dark px-3 py-1.5 rounded bg-gray-50 dark:bg-gray-800"
                    >
                        <span className="material-symbols-outlined text-base">refresh</span>
                        <span>Refresh</span>
                    </button>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-background-light dark:bg-background-dark/20">
                {downloads.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-64 text-gray-400 dark:text-gray-600">
                        <span className="material-symbols-outlined text-6xl mb-2">cloud_download</span>
                        <p className="text-lg">No active downloads</p>
                    </div>
                ) : (
                    downloads.map(download => (
                        <DownloadCard key={download.request_id} download={download} />
                    ))
                )}
            </div>
        </div>
    )
}