import type {Chunk} from '../../types';

interface Props {
    chunks: Chunk[];
}

export default function ChunkGrid({chunks}: Props) {
    const getChunkColor = (status: string) => {
        switch (status) {
            case 'received':
                return 'bg-green-500';
            case 'requesting':
                return 'bg-blue-500 animate-pulse';
            case 'error':
                return 'bg-red-500';
            default:
                return 'bg-gray-700';
        }
    };

    return (
        <div className="chunk-grid">
            {chunks.map((chunk, i) => (
                <div
                    key={i}
                    className={`chunk ${getChunkColor(chunk.status)}`}
                    title={`Chunk ${i}: ${chunk.status}`}
                />
            ))}
        </div>
    );
}