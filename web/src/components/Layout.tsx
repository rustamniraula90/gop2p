import Sidebar from "./sidebar/Sidebar.tsx";
import MainArea from "./mainarea/MainArea.tsx";


export default function Layout(){
    return (
        <div className="flex h-screen w-screen bg-bg text-grey-300">
            <div className="flex w-full max-w-[90%] h-[90%] mx-auto my-auto bg-sidebar-dark/70 backdrop-blur-lg border border-border rounded-xl overflow-hidden shadow-2xl">
                <Sidebar/>
                <MainArea/>
            </div>
        </div>
    )
}