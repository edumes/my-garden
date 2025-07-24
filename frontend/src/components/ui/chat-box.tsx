import React, { useEffect, useRef, useState } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { useWebSocket } from '../../contexts/WebSocketContext';
import { cn } from '../../lib/utils';
import { WebSocketChatEventData } from '../../types/api';
import { Button } from './button';
import { Input } from './input';
import { ScrollArea } from './scroll-area';

interface ChatBoxProps {
    gardenId: string;
    className?: string;
}

export const ChatBox: React.FC<ChatBoxProps> = ({ gardenId, className }) => {
    const [messages, setMessages] = useState<WebSocketChatEventData[]>([]);
    const [inputMessage, setInputMessage] = useState('');
    const { user } = useAuth();
    const { connected, lastMessage } = useWebSocket();
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const socket = useRef<WebSocket | null>(null);

    useEffect(() => {
        // Initialize WebSocket connection
        const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/api/ws`;
        socket.current = new WebSocket(wsUrl);

        socket.current.onopen = () => {
            console.log('Chat WebSocket connected');
        };

        return () => {
            if (socket.current) {
                socket.current.close();
            }
        };
    }, []);

    useEffect(() => {
        if (lastMessage?.type === 'chat' && lastMessage.garden_id === gardenId) {
            const chatMessage = lastMessage.data as WebSocketChatEventData;
            setMessages(prev => [...prev, chatMessage]);
            scrollToBottom();
        }
    }, [lastMessage, gardenId]);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    const sendMessage = () => {
        if (!inputMessage.trim() || !socket.current || !connected) return;

        const message = {
            type: 'chat',
            garden_id: gardenId,
            data: {
                message: inputMessage.trim()
            }
        };

        socket.current.send(JSON.stringify(message));
        setInputMessage('');
    };

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    };

    return (
        <div className={cn('flex flex-col h-[400px] border rounded-lg bg-background', className)}>
            <div className="p-3 border-b">
                <h3 className="font-semibold">Garden Chat</h3>
            </div>

            <ScrollArea className="flex-1 p-4">
                <div className="space-y-4">
                    {messages.map((msg, i) => (
                        <div
                            key={`${msg.timestamp}-${i}`}
                            className={cn(
                                'flex flex-col max-w-[80%] space-y-1',
                                msg.user_id === user?.id ? 'ml-auto items-end' : 'items-start'
                            )}
                        >
                            <span className="text-sm text-muted-foreground">
                                {msg.username}
                            </span>
                            <div
                                className={cn(
                                    'rounded-lg px-3 py-2',
                                    msg.user_id === user?.id
                                        ? 'bg-primary text-primary-foreground'
                                        : 'bg-muted'
                                )}
                            >
                                {msg.message}
                            </div>
                        </div>
                    ))}
                    <div ref={messagesEndRef} />
                </div>
            </ScrollArea>

            <div className="p-4 border-t">
                <div className="flex gap-2">
                    <Input
                        value={inputMessage}
                        onChange={(e) => setInputMessage(e.target.value)}
                        onKeyPress={handleKeyPress}
                        placeholder="Type a message..."
                        disabled={!connected}
                    />
                    <Button
                        onClick={sendMessage}
                        disabled={!connected || !inputMessage.trim()}
                    >
                        Send
                    </Button>
                </div>
            </div>
        </div>
    );
}; 