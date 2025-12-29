import type {Peer} from "../../types";
import {useApp} from "../../context/AppContext.tsx";
import {useState} from "react";

interface Props {
    peer: Peer;
}

export default function PeerItem({peer}: Props) {
    const {state, dispatch} = useApp()
    const [requested, setRequested] = useState(false)
    const isActive = state.currentPeerId == peer.id

    const selectPeer = () => {
        dispatch({type: 'SET_CURRENT_PEER', payload: peer.id})
    }

    const requestConnect = async (e: React.MouseEvent) => {
        e.stopPropagation()
        setRequested(true)
        await fetch('api/connect/request', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({target_id: peer.id})
        })
    }

    const removePeer = async (e: React.MouseEvent) => {
        e.stopPropagation()
        if (!confirm("Remove this peer?")) return;
        await fetch('/api/peers/remove', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({id: peer.id})
        })
    }


    const lastUsed = peer.status == 2 ? 'Active Now' :
        peer.last_used ? `Last: ${new Date(peer.last_used).toLocaleTimeString()}` : `Never`

    return (
        <li
            onClick={selectPeer}
            className={`p-3 rounded border transition-colors group relative cursor-pointer ${isActive
                ? 'border-primary/50 bg-blue-50/50 dark:bg-blue-900/20'
                : 'border-border-light dark:border-border-dark bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-800'
            }`}
        >
            <div className="flex justify-between items-start mb-1">
                <h4 className={`font-semibold transition-colors ${isActive ? 'text-primary dark:text-blue-400' : 'text-gray-900 dark:text-white'}`}>
                    {peer.name || peer.id.slice(0, 8)}
                </h4>
                {peer.status === 0 ? (
                    <button
                        onClick={requestConnect}
                        disabled={requested}
                        className={`text-xs px-3 py-1 rounded transition-colors ${requested
                            ? 'bg-gray-300 dark:bg-gray-700 text-gray-500 cursor-default'
                            : 'bg-primary hover:bg-blue-600 text-white shadow-sm'
                        }`}
                    >
                        {requested ? 'Requested' : 'Connect'}
                    </button>
                ) : (
                    <button
                        onClick={removePeer}
                        className="text-xs text-red-500 border border-red-500/30 px-2 py-0.5 rounded hover:bg-red-500 hover:text-white transition-colors"
                    >
                        Remove
                    </button>
                )}
            </div>
            <div className={`flex items-center text-xs ${peer.status === 2 ? 'text-green-600 dark:text-green-400' : 'text-gray-500 dark:text-gray-400'}`}>
                <span className={`w-2 h-2 rounded-full mr-2 ${peer.status === 2 ? 'bg-green-500' : 'bg-gray-400'}`}></span>
                {lastUsed}
            </div>
        </li>
    )
}