# Background Jobs & Queues

ZenoEngine contains a built-in, lightweight memory-backed worker pool. This allows you to offload heavy or slow logic (like sending emails, processing images, or calling external APIs) to the background without blocking the main HTTP request.

You don't need Redis, RabbitMQ, or any external service to run background jobs in ZenoEngine natively.

## Configuring the Worker Pool

By default, ZenoEngine boots a worker pool with 5 concurrent workers. You can adjust this configuration when setting up your engine, typically in `main.zl`.

```zeno
// main.zl
worker.config: {
    workers: 10              // Max concurrent Goroutines for jobs
    max_queue_size: 1000     // How many jobs can be queued before blocking
}
```

## Enqueuing a Job

To push logic to the background, use the `job.enqueue` slot. Any code inside the `do` block will be handed off to the worker pool and executed asynchronously. 

The `job.enqueue` block immediately returns, allowing your HTTP response to complete instantly while the work happens behind the scenes.

```zeno
http.post: '/api/register' {
    do: {
        http.json_body: { as: $user }
        
        // 1. Insert user into Database (Fast)
        db.table: "users"
        db.insert: $user

        // 2. Send Welcome Email (Slow)
        job.enqueue: {
            do: {
                // This payload captures variables from the surrounding scope
                $emailData: {
                    to: $user.email
                    subject: "Welcome!"
                    body: "We are glad you are here."
                }
                
                // Simulate an expensive external API call
                http.post: "https://api.mailgun.net/v3/..." {
                    body: $emailData
                }
                
                log: "Background email sent to " + $user.email
            }
        }

        // 3. Respond instantly
        http.ok: { message: "Account created! Check your email." }
    }
}
```

### Scope Capture

When you enqueue a job, ZenoEngine automatically isolates and captures the *current state* of variables inside the `do` block. The background execution runs in an independent, memory-safe context, protecting you from common concurrency bugs.

*Note: Since the background job outlives the original HTTP request, any database updates or API calls inside the job must manage their own connections if they rely on request-specific lifecycles.*
