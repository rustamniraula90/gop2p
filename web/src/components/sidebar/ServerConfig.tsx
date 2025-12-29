import {useEffect, useState} from "react";

export default function ServerConfig() {
    const [servers, setServers] = useState<string[]>([])
    const [selected, setSelected] = useState('')

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

    return (
        <div className="px-6 py-4 border-b border-border-light dark:border-border-dark">
            <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Server
                Config</h3>
            <div className="flex space-x-2">
                <div className="relative flex-1">
                    <select
                        value={selected}
                        onChange={(e) => setServer(e.target.value)}
                        className="w-full bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 rounded px-3 py-2 text-sm appearance-none focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent pr-8 bg-none"
                    >
                        {servers.map(s => <option key={s} value={s}>{s}</option>)}
                    </select>
                    <div
                        className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-500 dark:text-gray-400">
                        <span className="material-symbols-outlined text-sm">expand_more</span>
                    </div>
                </div>
                <button
                    onClick={addServer}
                    className="bg-primary hover:bg-blue-600 text-white rounded px-3 flex items-center justify-center transition-colors"
                >
                    <span className="material-symbols-outlined text-sm">add</span>
                </button>
            </div>
        </div>
    )
}