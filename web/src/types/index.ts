export interface AppState {
    id: string
    name: string
    peers: Record<string, Peer>
    connectionRequests: Array<{ id: string, name: string }>
    currentPeerId: string | null
    currentView: ViewType
    messages: Record<string, Message[]>
    files: Record<string, FileInfo[]>;
    downloads: Record<string, Download>;
    uploads: Record<string, Upload>;
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

export type DownloadStatus = 'pending' | 'requesting' | 'active' | 'complete' | 'error';
export type ChunkStatus = 'pending' | 'requesting' | 'received' | 'error';

export interface Chunk {
    index: number;
    status: ChunkStatus;
    sha: string;
    retry_count: number;
}

export interface Download {
    request_id: string;
    peer_id: string;
    peer_name: string;
    file_name: string;
    original_size: number;
    compressed_size: number;
    chunk_size: number;
    chunk_count: number;
    chunks: Chunk[];
    status: DownloadStatus;
    progress: number;
    start_time?: string;
}

export type UploadStatus = 'preparing' | 'active' | 'complete' | 'error';

export interface Upload {
    request_id: string;
    peer_id: string;
    peer_name: string;
    file_name: string;
    original_size: number;
    compressed_size: number;
    chunk_size: number;
    chunk_count: number;
    chunks_sent: number;
    status: UploadStatus;
    progress: number;
    start_time?: string;
}