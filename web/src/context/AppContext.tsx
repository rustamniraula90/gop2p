import {createContext, type Dispatch, type ReactNode, useContext, useReducer} from "react";
import type {AppState, Download, FileInfo, Message, Peer, Upload, ViewType} from "../types";

type Action =
    | { type: 'SET_VIEW'; payload: ViewType }
    | { type: 'SET_IDENTITY'; payload: { id: string; name: string } }
    | { type: 'UPDATE_PEER'; payload: Peer }
    | { type: 'SET_PEERS'; payload: Peer[] }
    | { type: 'ADD_CONNECTION_REQUEST', payload: { id: string, name: string } }
    | { type: 'REMOVE_CONNECTION_REQUEST', payload: string }
    | { type: 'SET_CURRENT_PEER'; payload: string | null }
    | { type: 'SET_MESSAGES'; payload: { peerId: string, messages: Message[] } }
    | { type: 'ADD_MESSAGES'; payload: { peerId: string, message: Message } }
    | { type: 'SET_FILES'; payload: { peerId: string, files: FileInfo[] } }
    | { type: 'SET_DOWNLOADS'; payload: Download[] }
    | { type: 'UPDATE_DOWNLOAD'; payload: any }
    | { type: 'SET_UPLOADS'; payload: Upload[] }
    | { type: 'UPDATE_UPLOAD'; payload: any }

const initialState: AppState = {
    id: '',
    name: 'Loading...',
    peers: {},
    connectionRequests: [],
    currentPeerId: null,
    currentView: 'chat',
    messages: {},
    files: {},
    downloads: {},
    uploads: {},
}

function appReducer(state: AppState, action: Action) {
    switch (action.type) {
        case "SET_VIEW":
            return {...state, currentView: action.payload};
        case "SET_IDENTITY":
            return {...state, id: action.payload.id, name: action.payload.name}
        case "UPDATE_PEER":
            return {...state, peers: {...state.peers, [action.payload.id]: action.payload}}
        case "SET_PEERS": {
            const peers: Record<string, Peer> = {};
            action.payload.forEach(p => peers[p.id] = p);
            return {...state, peers}
        }
        case "SET_CURRENT_PEER":
            return {...state, currentPeerId: action.payload}
        case 'ADD_CONNECTION_REQUEST':
            return {...state, connectionRequests: [...state.connectionRequests, action.payload]}
        case 'REMOVE_CONNECTION_REQUEST':
            return {...state, connectionRequests: state.connectionRequests.filter(r => r.id !== action.payload)}
        case 'SET_MESSAGES':
            return {...state, messages: {...state.messages, [action.payload.peerId]: action.payload.messages}}
        case 'ADD_MESSAGES':
            return {
                ...state,
                messages: {
                    ...state.messages,
                    [action.payload.peerId]: [...(state.messages[action.payload.peerId] || []), action.payload.message]
                }
            }
        case 'SET_FILES':
            return {...state, files: {...state.files, [action.payload.peerId]: action.payload.files}}
        case 'SET_DOWNLOADS': {
            const downloads: Record<string, Download> = {};
            action.payload.forEach(d => downloads[d.request_id] = d);
            return {...state, downloads};
        }
        case 'UPDATE_DOWNLOAD': {
            const payload = action.payload;
            const existing = state.downloads[payload.request_id];
            let chunks = payload.chunks || existing?.chunks;
            if (payload.last_chunk_idx !== undefined && payload.last_chunk_idx !== -1) {
                if (chunks) {
                    chunks = [...chunks];
                    if (chunks[payload.last_chunk_idx]) {
                        chunks[payload.last_chunk_idx] = {
                            ...chunks[payload.last_chunk_idx],
                            status: payload.last_chunk_stat,
                        };
                    }
                }
            }
            return {
                ...state,
                downloads: {
                    ...state.downloads,
                    [payload.request_id]: {
                        ...existing,
                        ...payload,
                        ...(chunks ? {chunks} : {}),
                    },
                },
            };
        }
        case 'SET_UPLOADS': {
            const uploads: Record<string, Upload> = {};
            action.payload.forEach(u => uploads[u.request_id] = u);
            return {...state, uploads};
        }
        case 'UPDATE_UPLOAD': {
            const payload = action.payload;
            const existing = state.uploads[payload.request_id];

            return {
                ...state,
                uploads: {
                    ...state.uploads,
                    [payload.request_id]: {
                        ...existing,
                        ...payload,
                    },
                },
            };
        }
        default:
            return state
    }
}

interface AppContextValue {
    state: AppState
    dispatch: Dispatch<Action>
}

const AppContext = createContext<AppContextValue | undefined>(undefined)

export function AppProvider({children}: Readonly<{ children: ReactNode }>) {
    const [state, dispatch] = useReducer(appReducer, initialState)
    return (
        <AppContext.Provider value={{state, dispatch}}>
            {children}
        </AppContext.Provider>
    )
}

export function useApp() {
    const context = useContext(AppContext);
    if (!context) {
        throw new Error('useApp must be used within AppProvider');
    }
    return context;
}