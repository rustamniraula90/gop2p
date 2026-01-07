import {useApp} from "../../context/AppContext.tsx";
import {useEffect} from "react";
import UploadCard from "./UploadCard.tsx";

export default function UploadsView() {
    const {state, dispatch} = useApp();

    useEffect(() => {
        // Fetch uploads initially and periodically
        const fetchUploads = async () => {
            try {
                const res = await fetch('/api/uploads');
                const data = await res.json();
                dispatch({type: 'SET_UPLOADS', payload: data || []});
            } catch (e) {
                console.error('Failed to fetch uploads', e);
            }
        };

        fetchUploads();
        const interval = setInterval(fetchUploads, 2000);
        return () => clearInterval(interval);
    }, [dispatch]);

    const uploads = Object.values(state.uploads);

    const clearFinished = async () => {
        await fetch('/api/uploads/clear', {method: 'POST'});
        // Refresh
        const res = await fetch('/api/uploads');
        const data = await res.json();
        dispatch({type: 'SET_UPLOADS', payload: data || []});
    };

    return (
        <div className="h-full flex flex-col">
            <div
                className="p-4 border-b border-border-light dark:border-border-dark flex justify-between items-center bg-panel-light dark:bg-panel-dark">
                <h2 className="text-xl font-bold text-gray-900 dark:text-white">Active Uploads</h2>
                <div className="flex space-x-2">
                    {uploads.some(u => u.status === 'complete' || u.status === 'error') && (
                        <button
                            onClick={clearFinished}
                            className="flex items-center space-x-1 text-sm text-gray-500 dark:text-gray-400 hover:text-red-500 dark:hover:text-red-400 transition-colors border border-border-light dark:border-border-dark px-3 py-1.5 rounded bg-gray-50 dark:bg-gray-800"
                        >
                            <span className="material-symbols-outlined text-base">delete_sweep</span>
                            <span>Clear Finished</span>
                        </button>
                    )}
                    <button
                        onClick={() => fetch('/api/uploads').then(r => r.json()).then(d => dispatch({
                            type: 'SET_UPLOADS',
                            payload: d || []
                        }))}
                        className="flex items-center space-x-1 text-sm text-gray-500 dark:text-gray-400 hover:text-primary dark:hover:text-primary transition-colors border border-border-light dark:border-border-dark px-3 py-1.5 rounded bg-gray-50 dark:bg-gray-800"
                    >
                        <span className="material-symbols-outlined text-base">refresh</span>
                        <span>Refresh</span>
                    </button>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-background-light dark:bg-background-dark/20">
                {uploads.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-64 text-gray-400 dark:text-gray-600">
                        <span className="material-symbols-outlined text-6xl mb-2">cloud_upload</span>
                        <p className="text-lg">No active uploads</p>
                    </div>
                ) : (
                    uploads.map(upload => (
                        <UploadCard key={upload.request_id} upload={upload}/>
                    ))
                )}
            </div>
        </div>
    );
}