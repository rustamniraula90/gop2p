import {useApp} from "../../context/AppContext.tsx";
import PeerItem from "./PeerItem.tsx";

interface Props {
    filter: string
}

export default function PeerList({filter}: Readonly<Props>) {
    const {state} = useApp()

    const filteredPeers = Object.values(state.peers).filter(p =>
        p.name.toLowerCase().includes(filter.toLowerCase()) || p.id.includes(filter)
    ).sort((a, b) => a.status - b.status)

    return (
        <ul className="space-y-3">
            {filteredPeers.map(peer => (
                <PeerItem key={peer.id} peer={peer} />
            ))}
            {filteredPeers.length === 0 && (
                <li className="text-gray-500 text-sm text-center py-4">No peers found</li>
            )}
        </ul>
    )
}