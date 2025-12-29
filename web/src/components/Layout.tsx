import Sidebar from "./sidebar/Sidebar.tsx";
import MainArea from "./mainarea/MainArea.tsx";


export default function Layout() {
    return (
        <div
            className="flex h-screen w-screen bg-background-light dark:bg-background-dark text-gray-800 dark:text-gray-200 transition-colors duration-200 font-sans overflow-hidden">
            <Sidebar/>
            <MainArea/>
        </div>
    )
}