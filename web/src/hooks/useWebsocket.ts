import {useEffect, useRef} from "react";
import type {WSMessage} from "../types";
import {useApp} from "../context/AppContext.tsx";

export function useWebsocket() {
    const {dispatch} = useApp()
    const wsRef = useRef<WebSocket | null>(null)

    useEffect(() => {
        const connect = () => {
            let protocol = "ws"
            if (window.location.protocol === "https") {
                protocol = "wss"
            }
            const ws = new WebSocket(`${protocol}://${window.location.host}/ws`)
            // const ws = new WebSocket(`ws://localhost:8081/ws`);


            ws.onopen = () => {
                console.log("websocket connected")
            }

            ws.onmessage = (event) => {
                const msg: WSMessage = JSON.parse(event.data)
                handleMessage(msg)
            }

            ws.onerror = (error) => {
                console.error("websocket error", error)
            }

            ws.onclose = () => {
                console.log("websocket disconnected, reconnecting...")
                setTimeout(connect, 3000)
            }
        }

        connect();

        return () => {
            if (wsRef.current) {
                wsRef.current.close()
            }
        }
    }, []);

    const handleMessage = (msg: WSMessage) => {
        console.log("Received message:", msg)
        switch (msg.type) {
            case 'identity':
                dispatch({type: 'SET_IDENTITY', payload: msg.data})
                break
            case 'peers':
                dispatch({type: 'SET_PEERS', payload: msg.data});
                break;
            case 'peer_update':
                dispatch({type: 'UPDATE_PEER', payload: msg.data});
                break;
            case 'connection_request':
                dispatch({type: 'ADD_CONNECTION_REQUEST', payload: {id: msg.data.id, name: msg.data.name}})
                break;
        }
    }
    return null;
}