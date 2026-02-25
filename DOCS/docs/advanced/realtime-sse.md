# Realtime Server-Sent Events (SSE)

ZenoEngine provides built-in, first-class support for **Server-Sent Events (SSE)**. This allows you to stream data from the server to the client in realtime without the overhead or complexity of WebSockets. SSE is perfect for dashboards, live feeds, progress bars, and notifications.

## Basic Server-Sent Events

To start an SSE stream, use the `sse.stream` slot on any HTTP GET route. This will automatically set the correct headers (`Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`) and keep the connection open.

```zeno
http.get: '/api/stream' {
    do: {
        sse.stream: {
            do: {
                // Your SSE logic goes here
                // Note: sse.keepalive is called automatically every 15s
            }
        }
    }
}
```

## Sending Events

Inside an `sse.stream` block, you use `sse.send` to push data to the connected client. You can send plain text, HTML (useful for HTMX/Datastar), or JSON.

```zeno
http.get: '/api/notifications' {
    do: {
        sse.stream: {
            do: {
                // Send a simple string message
                sse.send: "Connection established!"
                
                // Send JSON data with a specific event name
                $payload: { status: "processing", progress: 50 }
                sse.send: $payload { event: "update" }
                
                // Send HTML (example: for Datastar/HTMX morphing)
                sse.send: "<div>New Notification</div>" { event: "datastar-fragment" }
            }
        }
    }
}
```

## Using SSE Loops

Usually, SSE endpoints run continuously or tick at intervals to send updates. ZenoLang provides `sse.loop` as a lightweight construct for periodic streaming.

```zeno
http.get: '/api/ticker' {
    do: {
        sse.stream: {
            do: {
                // Loop 10 times, pausing for 1 second between each tick
                sse.loop: 10 {
                    delay: "1s"
                    do: {
                        date.now: { as: $time }
                        sse.send: "Current time is: " + $time
                    }
                }
            }
        }
    }
}
```

If you don't specify the main iteration count (e.g., `sse.loop: { delay: "2s" }`), it will loop indefinitely until the client disconnects.

## Consuming SSE on the Frontend

Because ZenoEngine uses standard SSE, you can consume it using vanilla JavaScript's `EventSource` API in any frontend framework, or with HTML-over-the-wire libraries like HTMX and Datastar.

### Vanilla JavaScript
```javascript
const eventSource = new EventSource('/api/notifications');

// Listen to default messages
eventSource.onmessage = function(event) {
    console.log("New message:", event.data);
};

// Listen to specific named events
eventSource.addEventListener('update', function(event) {
    const data = JSON.parse(event.data);
    console.log("Update received:", data.progress);
});
```
