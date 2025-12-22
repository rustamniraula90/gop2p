import {useApp} from "../../context/AppContext.tsx";

export default function Sidebar() {
    const { state } = useApp();
    return (
        <aside className="w-80 bg-sidebar border-r border-border flex flex-col p-5 space-y-6">
            <header className="space-y-2">
                <h2 className="text-xl font-bold text-white tracking-wider">GoP2P</h2>
                <div className="text-grey-400">
                    <span className="font-semibold text-lg block">{state.name}</span>
                </div>
            </header>
        </aside>
    )
}