import {useEffect, useState} from "react";


export default function ThemeToggle() {
    const [darkMode, setDarkMode] = useState(false)

    useEffect(() => {
        const checkDark = () => setDarkMode(document.documentElement.classList.contains('dark'));
        checkDark();

        const observer = new MutationObserver(checkDark);
        observer.observe(document.documentElement, {attributes: true, attributeFilter: ['class']});
        return () => observer.disconnect();
    }, []);

    const toggleTheme = () => {
        const isDark = document.documentElement.classList.toggle('dark');
        localStorage.theme = isDark ? 'dark' : 'light';
        setDarkMode(isDark)
    };

    return (
        <div className="fixed bottom-4 right-4 z-50">
            <button
                onClick={toggleTheme}
                title={darkMode ? "Switch to Light Mode" : "Switch to Dark Mode"}
                className="bg-gray-800 dark:bg-white text-white dark:text-gray-800 p-3 rounded-full shadow-lg hover:scale-110 active:scale-95 transition-all focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary flex items-center justify-center border border-border-light dark:border-border-dark"
        >
                <span className="material-symbols-outlined text-xl">
                    {darkMode ? 'wb_sunny': 'bedtime'}
                </span>
            </button>
        </div>
    )
}