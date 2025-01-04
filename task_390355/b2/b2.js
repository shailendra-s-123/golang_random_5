const ws = new WebSocket('wss://your-server.com/ws');

ws.onopen = () => {
    console.log('Connected');
};

ws.onclose = () => {
    console.log('Disconnected');
    setTimeout(() => {
        ws.connect();
    }, Math.pow(2, Math.floor(Math.random() * 8)) * 1000); // Exponential backoff
};

ws.onmessage = (event) => {
    console.log('Received message:', event.data);
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};