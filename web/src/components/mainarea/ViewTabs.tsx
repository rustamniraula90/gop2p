import {useApp} from "../../context/AppContext.tsx";
import type {ViewType} from "../../types";


export default function ViewTabs() {
    const {state, dispatch} = useApp()

    const tabs : Array<{ id: ViewType; label: string; icon: string }> = [
        {id: 'chat', label: 'Chat', icon: 'chat'},
        {id: 'files', label: 'Files', icon: 'folder'},
        {id: 'downloads', label: 'Downloads', icon: 'download'},
        {id: 'uploads', label: 'Uploads', icon: 'upload'},
    ]

    return (
        <div className="flex space-x-4">
            {tabs.map(tab => (
                <button
                    key={tab.id}
                    onClick={() => dispatch({type: 'SET_VIEW', payload: tab.id})}
                    className={`flex items-center space-x-2 px-4 py-2 rounded text-sm font-medium transition-all ${state.currentView === tab.id
                        ? 'bg-primary text-white shadow-sm active:scale-95'
                        : 'bg-transparent text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 border border-transparent hover:border-border-light dark:hover:border-border-dark'
                    }`}
                >
                    <span className="material-symbols-outlined text-lg">{tab.icon}</span>
                    <span>{tab.label}</span>
                </button>
            ))}
        </div>
    )
}