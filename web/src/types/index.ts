export interface AppState {
    id: string
    name: string
    peers: Record<string, Peer>
}

export interface Peer {
    id: string
    name: string
    status: 0 | 1 | 2 // 0: disconnected, 1: punching, 2: connected
    last_seen: string
}

export interface WSMessage {
    type: string
    data: any
}