import {useApp} from "../../context/AppContext.tsx";
import {useEffect, useRef, useState} from "react";
import * as React from "react";

export default function ChatView() {
    const {state, dispatch} = useApp();
    const [message, setMessage] = useState('');

    const currentPeer = state.currentPeerId ? state.peers[state.currentPeerId] : null;
    const messages = state.currentPeerId ? state.messages[state.currentPeerId] || [] : [];
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const canSend = currentPeer?.status === 2;


    useEffect(() => {
        if (state.currentPeerId) {
            fetch(`api/messages?peer_id=${state.currentPeerId}`)
                .then(res => res.json())
                .then(msgs => {
                    dispatch({type: 'SET_MESSAGES', payload: {peerId: state.currentPeerId!, messages: msgs || []}});
                });
        }
    }, [state.currentPeerId]);

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({behavior: 'smooth'})
    }, [messages]);

    const sendMessage = async () => {
        if (!message.trim() || !state.currentPeerId || !canSend) return;

        await fetch('/api/message', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({target_id: state.currentPeerId, text: message})
        });

        setMessage('');
    }

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key == 'Enter' && !e.shiftKey) {
            e.preventDefault()
            sendMessage()
        }
    }

    return (
        <div className="flex flex-col h-full bg-white dark:bg-gray-800/50 transition-colors duration-200">
            <div
                className="px-6 py-4 border-b border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark">
                <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                    Chat with {currentPeer?.name || 'Select a peer'}
                </h2>
            </div>
            <div className="flex-1 p-4 overflow-y-auto scrollbar-thin flex flex-col space-y-2 min-h-0">
                {messages.length === 0 ? (
                    <div
                        className="flex-1 flex items-center justify-center text-gray-400 dark:text-gray-600 italic text-sm">
                        No messages yet
                    </div>
                ) : (
                    messages.map((msg, i) => {
                        const isMe = msg.sent || msg.sender_id === 'me' || msg.sender_id === state.id;
                        return (
                            <div
                                key={i}
                                className={`px-3 py-2 rounded-lg text-sm max-w-[85%] break-words shadow-sm ${isMe
                                    ? 'bg-primary text-white ml-auto rounded-br-none'
                                    : 'bg-white dark:bg-gray-800 border border-border-light dark:border-border-dark text-gray-700 dark:text-gray-300 rounded-bl-none'
                                }`}
                            >
                                {msg.text}
                            </div>
                        );
                    })
                )}
                <div ref={messagesEndRef}/>
            </div>
            <div className="p-4 pt-2">
                <div className="flex space-x-2">
                    <input
                        type="text"
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        onKeyPress={handleKeyPress}
                        disabled={!canSend}
                        placeholder={canSend ? "Message..." : "Not connected"}
                        className="flex-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all disabled:opacity-50"
                    />
                    <button
                        onClick={sendMessage}
                        disabled={!canSend}
                        className="bg-primary hover:bg-blue-600 disabled:bg-gray-400 text-white rounded px-3 flex items-center justify-center transition-colors shadow-sm"
                    >
                        <span
                            className="material-symbols-outlined text-sm transform rotate-[-45deg] relative left-[1px]">send</span>
                    </button>
                </div>
            </div>
        </div>
    )
}