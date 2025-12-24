import {createContext, type Dispatch, type ReactNode, useContext, useReducer} from "react";
import type {AppState, Peer} from "../types";

type Action =
    | { type: 'SET_IDENTITY'; payload: { id: string; name: string } }
    | { type: 'UPDATE_PEER'; payload: Peer }
    | { type: 'SET_PEERS'; payload: Peer[] }
    | { type: 'ADD_CONNECTION_REQUEST', payload: { id: string, name: string } }
    | { type: 'REMOVE_CONNECTION_REQUEST', payload: string }
    | { type: 'SET_CURRENT_PEER'; payload: string | null }

const initialState: AppState = {
    id: '',
    name: 'Loading...',
    peers: {},
    connectionRequests: [],
    currentPeerId: null
}

function appReducer(state: AppState, action: Action) {
    switch (action.type) {
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
        default:
            return state
    }
}

interface AppContextValue {
    state: AppState
    dispatch: Dispatch<Action>
}

const AppContext = createContext<AppContextValue | undefined>(undefined)

export function AppProvider({children}: { children: ReactNode }) {
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