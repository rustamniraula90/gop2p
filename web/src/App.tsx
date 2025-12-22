import './App.css'
import Layout from "./components/Layout.tsx";
import {useWebsocket} from "./hooks/useWebsocket.ts";
import {AppProvider} from "./context/AppContext.tsx";

function AppContent() {
    useWebsocket();
    return <Layout/>
}

function App() {
    return (
        <AppProvider>
            <AppContent/>
        </AppProvider>
    )
}

export default App
