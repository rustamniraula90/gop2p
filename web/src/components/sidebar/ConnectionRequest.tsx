import {useApp} from "../../context/AppContext.tsx";


export default function ConnectionRequest() {
    const {state, dispatch} = useApp()

    const acceptRequest = async (id: string) => {
        await fetch('/api/connect/accept', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({requester_id: id}),
        });
        dispatch({type: `REMOVE_CONNECTION_REQUEST`, payload: id})
    }

    return (
        <div className="space-y-2">
            <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wide">Incoming Requests</h3>
            {state.connectionRequests.map(req => (
                <div key={req.id}
                     className="p-3 bg-primary/10 border border-primary rounded-lg flex justify-between items-center">
                    <span className="text-sm"><strong>{req.name}</strong> wants to connect</span>
                    <button
                        onClick={() => acceptRequest(req.id)}
                        className="px-3 py-1 bg-green-500 hover:bg-green-600 text-white text-sm rounded transition">
                        Accept
                    </button>
                </div>
            ))}
        </div>
    )
}