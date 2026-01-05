export interface AppState {
    id: string
    name: string
    peers: Record<string, Peer>
    connectionRequests: Array<{ id: string, name: string }>
    currentPeerId: string | null
    currentView: ViewType
    messages: Record<string, Message[]>
    files: Record<string, FileInfo[]>;
}

export type ViewType = 'files' | 'downloads' | 'uploads' | 'chat';

export interface Peer {
    id: string
    name: string
    status: 0 | 1 | 2 // 0: disconnected, 1: punching, 2: connected
    last_used: string
}

export interface WSMessage {
    type: string
    data: any
}

export interface Message {
    sender_id: string;
    text: string;
    timestamp: string;
    sent?: boolean;
}

export interface FileInfo {
    name: string;
    size: number;
    type: 'file' | 'dir';
}