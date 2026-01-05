export function getStatusText(status: number): string {
    switch (status) {
        case 0:
            return 'Disconnected'
        case 1:
            return 'Connecting...'
        case 2:
            return 'Connected'
        default:
            return 'Unknown'
    }
}

export function formatFileSize(bytes: number): string {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}