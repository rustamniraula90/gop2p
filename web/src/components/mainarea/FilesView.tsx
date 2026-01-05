import {useApp} from "../../context/AppContext.tsx";
import FileItem from "./FileItem.tsx";

export default function FilesView() {
    const {state} = useApp();
    console.log({state})

    if (!state.currentPeerId) {
        return (
            <div className="flex items-center justify-center h-full">
                <h2 className="text-2xl text-gray-500">Select a peer to share files</h2>
            </div>
        )
    }

    const peer = state.peers[state.currentPeerId];
    const files = state.files[state.currentPeerId];
    console.log("FILES", files)

    const refreshFiles = () => {
        if (!state.currentPeerId) return;
        fetch('/api/files/request', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({peer_id: state.currentPeerId}),
        })
    }
    return (
        <div className="h-full flex flex-col">
            <div
                className="p-4 border-b border-border-light dark:border-border-dark flex justify-between items-center bg-panel-light dark:bg-panel-dark">
                <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                    Files on {peer?.name || state.currentPeerId.slice(0, 8)}
                </h2>
                <button
                    onClick={refreshFiles}
                    className="flex items-center space-x-1 text-sm text-gray-500 dark:text-gray-400 hover:text-primary dark:hover:text-primary transition-colors border border-border-light dark:border-border-dark px-3 py-1.5 rounded bg-gray-50 dark:bg-gray-800"
                >
                    <span className="material-symbols-outlined text-base">refresh</span>
                    <span>Refresh</span>
                </button>
            </div>

            <div className="overflow-x-auto flex-1 bg-background-light dark:bg-background-dark/20">
                <table className="w-full text-left border-collapse">
                    <thead>
                    <tr className="border-b border-border-light dark:border-border-dark text-xs uppercase text-gray-500 dark:text-gray-400 tracking-wider">
                        <th className="px-6 py-4 font-semibold">Name</th>
                        <th className="px-6 py-4 font-semibold">Size</th>
                        <th className="px-6 py-4 font-semibold">Type</th>
                        <th className="px-6 py-4 font-semibold text-right">Action</th>
                    </tr>
                    </thead>
                    <tbody className="text-sm">
                    {!files || files.length === 0 ? (
                        <tr>
                            <td colSpan={4} className="px-6 py-12 text-center text-gray-500 italic">
                                {peer?.status === 2 ? 'No files shared (or directory empty)' : 'Not connected to peer'}
                            </td>
                        </tr>
                    ) : (
                        files.map((file, i) => (
                            <FileItem key={i} file={file} peerId={state.currentPeerId!}/>
                        ))
                    )}
                    </tbody>
                </table>
            </div>
        </div>
    )
}