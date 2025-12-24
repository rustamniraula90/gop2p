import type {Peer} from "../../types";
import {useApp} from "../../context/AppContext.tsx";
import {useState} from "react";
import {getStatusText} from "../../utils/formatters.ts";

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
            className={`p-3 bg-bg border border-border rounded-lg cursor-pointer transition hiver:border-primary flex justify-between items-center ${isActive ? 'border-primary bg-sidebar' : ''}`}
        >
            <div className='flex-1 min-w-0'>
                <div className="flex items-center gap-2">
                    <span
                        className="font-medium truncate text-gray-200 group-hover:text-white transition">{peer.name || peer.id.slice(0, 8)}</span>
                    <span
                        className='absolute left-full ml-2 px-1.5 py-0.5 bg-sidebar-dark border border-border text-[10px] rounded opacity-0 group-hover/status:opacity-100 transition whitespace-nowrap z-10 pointer-events-none lowercase'>
                        {getStatusText(peer.status)}
                    </span>
                </div>
                <small className="text-[10px] text-gray-500 font-mono">{lastUsed}</small>
            </div>
            {peer.status == 0 ? (
                <button
                    onClick={requestConnect}
                    disabled={requested}
                    className={`ml-2 px-2 text-xs text-white rounded transition ${requested ? 'bg-gray-600 cursor-default' : 'bg-primary hover:bg-primary-hover'}`}
                >
                    {requested ? 'Requested' : 'Connect'}
                </button>
            ) : (
                <button onClick={removePeer}
                        className='ml-2 px-2 py-1 text-xs border border-red-500 text-red-500 hover:bg-red-500 hover:text-white rounded transition'>
                    Remove
                </button>
            )}
        </li>
    )
}