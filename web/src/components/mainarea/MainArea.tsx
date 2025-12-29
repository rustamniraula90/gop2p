import ViewTabs from "./ViewTabs.tsx";
import {useApp} from "../../context/AppContext.tsx";
import ChatView from "./ChatView.tsx";
import UploadsView from "./UploadsView.tsx";
import DownloadsView from "./DownloadsView.tsx";
import FilesView from "./FilesView.tsx";

export default function MainArea() {
    const {state} = useApp()

    return (
        <main
            className="flex-1 flex flex-col min-w-0 bg-background-light dark:bg-background-dark/20 transition-colors duration-200">
            <header
                className="h-16 border-b border-border-light dark:border-border-dark flex items-center px-6 bg-panel-light dark:bg-panel-dark flex-shrink-0 justify-between">
                <ViewTabs/>
            </header>

            <div className="flex-1 min-h-0">
                {state.currentView === 'chat' && <ChatView/>}
                {state.currentView === 'files' && <FilesView/>}
                {state.currentView === 'downloads' && <DownloadsView/>}
                {state.currentView === 'uploads' && <UploadsView/>}
            </div>
        </main>
    )
}