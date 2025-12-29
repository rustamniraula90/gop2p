import './App.css'
import Layout from "./components/Layout";
import {useWebsocket} from "./hooks/useWebsocket";
import {AppProvider} from "./context/AppContext";
import ThemeToggle from "./components/ThemeToggle.tsx";

function AppContent() {
    useWebsocket();
    return (
        <>
            <Layout/>
            <ThemeToggle/>
        </>
    )
}

function App() {
    return (
        <AppProvider>
            <AppContent/>
        </AppProvider>
    )
}

export default App
