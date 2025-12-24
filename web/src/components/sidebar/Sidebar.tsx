import {useApp} from "../../context/AppContext.tsx";
import ServerConfig from "./ServerConfig.tsx";
import {useState} from "react";
import ConnectionRequest from "./ConnectionRequest.tsx";
import PeerList from "./PeerList.tsx";

export default function Sidebar() {
    const {state} = useApp();
    const [searchTerm, setSearchTerm] = useState('')
    return (
        <aside className="w-80 bg-sidebar border-r border-border flex flex-col p-5 space-y-6">
            <header className="space-y-2">
                <h2 className="text-xl font-bold text-white tracking-wider">GoP2P</h2>
                <div className="text-grey-400">
                    <span className="font-semibold text-lg block">{state.name}</span>
                </div>
            </header>
            <ServerConfig onSearch={setSearchTerm}/>

            {state.connectionRequests.length > 0 && <ConnectionRequest/>}

            <PeerList filter={searchTerm}/>

        </aside>
    )
}