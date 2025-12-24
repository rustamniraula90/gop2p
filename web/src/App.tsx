import './App.css'
import Layout from "./components/Layout";
import {useWebsocket} from "./hooks/useWebsocket";
import {AppProvider} from "./context/AppContext";

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
