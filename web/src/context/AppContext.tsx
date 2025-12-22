import {createContext, type Dispatch, type ReactNode, useContext, useReducer} from "react";
import type {AppState, Peer} from "../types";

type Action =
    | { type: 'SET_IDENTITY'; payload: { id: string; name: string } }
    | { type: 'UPDATE_PEER'; payload: Peer }

const initialState: AppState = {
    id: '',
    name: 'Loading...',
    peers: {}
}

function appReducer(state: AppState, action: Action) {
    switch (action.type) {
        case "SET_IDENTITY":
            return {...state, id: action.payload.id, name: action.payload.name}
        case "UPDATE_PEER":
            return {...state, peers: {...state.peers, [action.payload.id]: action.payload}}
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