# gop2p Web UI

A modern, responsive, and dark-themed web interface for the gop2p system, built with React and Vite.

## Tech Stack

- **React**: UI library.
- **TypeScript**: Type-safe development.
- **Tailwind CSS**: Utility-first styling with a custom dark theme.
- **Vite**: Ultra-fast build tool and dev server.
- **Lucide React**: For consistent and beautiful iconography.
- **Framer Motion**: Smooth animations and transitions.

## Features

- **Real-time Updates**: Uses WebSockets to synchronize with the Go client backend.
- **Peer Management**: Interactive list of discovered and connected peers.
- **E2E Chat**: Immersive chat interface with glassmorphism design.
- **File Sharing**: Integrated file viewer and sharing manager with real-time transfer progress.
- **Responsive Design**: Optimized for various screen sizes.

## Communication with Backend

The frontend communicates with the Go client's `APIServer` via:
- **WebSocket (`/ws`)**: For real-time event streaming (messages, peer updates, progress).
- **REST API (`/api/*`)**: For state transitions and data fetching.

## Development

1.  Install dependencies:
    ```bash
    npm install
    ```
2.  Start the development server:
    ```bash
    npm run dev
    ```
    *Note: The frontend expects the Go client to be running on `:8081` to handle API requests.*

## Build

The frontend is built and bundled into `web/dist`. This process is automated by the root `Makefile`.

```bash
npm run build
```

The resulting `dist` folder is then embedded into the Go client binary during the Go build process.
