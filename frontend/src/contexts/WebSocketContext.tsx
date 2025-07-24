import React, { createContext, useContext, useEffect, useState } from 'react';
import { useAuth } from './AuthContext';
import { WebSocketEvent } from '../types/api';

interface WebSocketContextType {
  connected: boolean;
  lastMessage: WebSocketEvent | null;
}

const WebSocketContext = createContext<WebSocketContextType>({
  connected: false,
  lastMessage: null,
});

export const useWebSocket = () => useContext(WebSocketContext);

interface WebSocketProviderProps {
  children: React.ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WebSocketEvent | null>(null);
  const { user } = useAuth();

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (!token || !user) {
      if (socket) {
        socket.close();
        setSocket(null);
        setConnected(false);
      }
      return;
    }

    const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/api/ws`;
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('WebSocket connected');
      setConnected(true);
      
      // Send authentication token
      ws.send(JSON.stringify({ type: 'auth', token }));
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setConnected(false);
      
      // Attempt to reconnect after 5 seconds
      setTimeout(() => {
        if (token && user) {
          setSocket(null);
        }
      }, 5000);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as WebSocketEvent;
        console.log('WebSocket message:', message);
        setLastMessage(message);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    setSocket(ws);

    return () => {
      if (ws) {
        ws.close();
      }
    };
  }, [user]);

  const value = {
    connected,
    lastMessage,
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}; 