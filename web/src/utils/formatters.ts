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