import {useApp} from "../../context/AppContext.tsx";
import ServerConfig from "./ServerConfig.tsx";
import {useState} from "react";
import ConnectionRequest from "./ConnectionRequest.tsx";
import PeerList from "./PeerList.tsx";

export default function Sidebar() {
    const {state} = useApp();
    const [searchTerm, setSearchTerm] = useState('')
    const [copied, setCopied] = useState(false)



    const copyId = () => {
        if (!state.id) return;
        navigator.clipboard.writeText(state.id)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
    }

    return (
        <aside className="w-80 bg-sidebar-light dark:bg-sidebar-dark border-r border-border-light dark:border-border-dark flex flex-col flex-shrink-0 h-screen overflow-y-auto z-10 transition-colors duration-200">
            <div className="p-6 pb-4">
                <div className="flex items-center space-x-3 mb-1">
                    <h1 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">GoP2P</h1>
                </div>
                <p className="text-lg font-semibold text-gray-700 dark:text-gray-300">{state.name}</p>
            </div>

            <div className="px-6 py-4 border-b border-border-light dark:border-border-dark">
                <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Your Identity</h3>
                <div className="flex items-center space-x-2">
                    <div className="flex-1 bg-gray-100 dark:bg-gray-800 rounded px-3 py-2 text-xs font-mono text-gray-600 dark:text-gray-400 truncate border border-gray-200 dark:border-gray-700" title={state.id}>
                        {state.id || 'Loading...'}
                    </div>
                    <button
                        onClick={copyId}
                        className="text-xs border border-border-light dark:border-border-dark px-2 py-2 rounded bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors text-gray-600 dark:text-gray-300 font-medium whitespace-nowrap"
                    >
                        {copied ? 'Copied' : 'Copy ID'}
                    </button>
                </div>
            </div>


            <ServerConfig/>

            {state.connectionRequests.length > 0 && <ConnectionRequest/>}

            <div className="px-6 py-4 flex flex-col min-h-0 border-b border-border-light dark:border-border-dark">
                <div className="flex justify-between items-center mb-3">
                    <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Active Peers</h3>
                    <button
                        onClick={() => fetch('/api/peers')}
                        className="text-gray-400 hover:text-primary transition-colors"
                        title="Refresh"
                    >
                        <span className="material-symbols-outlined text-sm">refresh</span>
                    </button>
                </div>
                <div className="relative mb-4">
                    <input
                        type="text"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        placeholder="Search by name or ID..."
                        className="w-full bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 rounded px-3 py-2 text-sm placeholder-gray-500 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                    />
                </div>
            </div>

            <div className="flex-1 overflow-y-auto px-6 py-4">
                <PeerList filter={searchTerm} />
            </div>

        </aside>
    )
}