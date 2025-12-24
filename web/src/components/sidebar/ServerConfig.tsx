import {useApp} from "../../context/AppContext.tsx";
import {useEffect, useState} from "react";

interface Props {
    onSearch: (term: string) => void;
}

export default function ServerConfig({onSearch}: Props) {
    const {state} = useApp()
    const [servers, setServers] = useState<string[]>([])
    const [selected, setSelected] = useState('')
    const [copied, setCopied] = useState(false)

    useEffect(() => {
        loadServers();
    }, []);

    const loadServers = async () => {
        try {
            const res = await fetch('/api/server/config/list')
            const data = await res.json()
            if (data && data.length > 0) {
                setServers(data.map((c: any) => c.address));
                const currentServerRes = await fetch('/api/server/config/current')
                const currentServ = await currentServerRes.json()
                if (currentServ) {
                    setSelected(data.address)
                } else {
                    setSelected('127.0.0.1:8080')
                }
            } else {
                setServers(['127.0.0.1:8080'])
                setSelected('127.0.0.1:8080')
            }
        } catch (e) {
            console.error('Failed to load servers', e)
        }
    }

    const setServer = async (addr: string) => {
        if (!addr) return;
        try {
            await fetch('/api/server/config', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({server_addr: addr})
            });
            setSelected(addr)
        } catch (e) {
            console.error('Failed to set server', e)
        }
    }

    const addServer = () => {
        const addr = prompt('Enter server address (e.g. 1.2.3.4:8080')
        if (!addr) return;
        setServer(addr).then(loadServers)
    }

    const copyId = () => {
        if (!state.id) return;
        navigator.clipboard.writeText(state.id)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
    }

    const refreshPeers = async () => {
        await fetch('/api/peers')
    }

    return (
        <div className="space-y-4">
            <div className="space-y-2">
                <h3 className="text-xs font-semibold text-graay-500 uppercase tracking-widest">Your Identity</h3>
                <div className="flex items-center gap-2 p-2 bg-bg border border-border rounded-lg group">
                    <div className="flex-1 truncate font-mono text-xs text-primary">
                        {state.id || 'Loading'}
                    </div>
                    <button
                        onClick={copyId}
                        className={`text-[10px] px-2 py-0.5 rounded border transition ${copied ? 'bg-green-600 border-green-600 text-white' : 'border-border hover:border-primary text-gray-400 group-hover:text-primary'}`}
                    >
                        {copied ? 'Copied' : 'Copy Id'}
                    </button>
                </div>

            </div>
            <div className="space-y-2">
                <h3 className="text-xs font-semibold text-graay-500 uppercase tracking-widest">Server Config</h3>
                <div className="flex gap-2">
                    <select
                        value={selected}
                        onChange={(e) => setServer(e.target.value)}
                        className="flex-1 px-3 py-2 bg-bg border border-border rounded text-gray-300 text-sm focus:outline-none focus:border-primary transition"
                    >
                        {servers.map(s => <option key={s} value={s}>{s}</option>)}
                    </select>
                    <button
                        onClick={addServer}
                        className="w-10 h-10 flex items-center justify-center bg-primary hover:bg-primary-hover text-white rounded transition text-xl"
                        title="Add New">
                        +
                    </button>
                </div>
            </div>
            <div className="pt-2">
                <div className="flex items-center justify-between mb-2">
                    <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-widest">Active Peers</h3>
                    <button
                        onClick={refreshPeers}
                        className="text-lg hover:text-primary transition p-1"
                        title="Refresh Peers">
                        ↻
                    </button>
                </div>
                <input
                    id="peer-search"
                    type="text"
                    placeholder="search by name or ID..."
                    onChange={(e) => onSearch(e.target.value)}
                    className="w-full px-3 py-2 bg-bg border border-border rounded text-gray-300 text-sm focus:outline-none focus:border-primary transition"
                />
            </div>
        </div>
    )
}