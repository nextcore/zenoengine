import{_ as a,o as s,c as e,ag as o}from"./chunks/framework.ePeAWSvT.js";const h=JSON.parse('{"title":"Background Jobs & Queues","description":"","frontmatter":{},"headers":[],"relativePath":"advanced/jobs-queues.md","filePath":"advanced/jobs-queues.md"}'),p={name:"advanced/jobs-queues.md"};function t(i,n,l,c,r,u){return s(),e("div",null,[...n[0]||(n[0]=[o(`<h1 id="background-jobs-queues" tabindex="-1">Background Jobs &amp; Queues <a class="header-anchor" href="#background-jobs-queues" aria-label="Permalink to &quot;Background Jobs &amp; Queues&quot;">​</a></h1><p>ZenoEngine contains a built-in, lightweight memory-backed worker pool. This allows you to offload heavy or slow logic (like sending emails, processing images, or calling external APIs) to the background without blocking the main HTTP request.</p><p>You don&#39;t need Redis, RabbitMQ, or any external service to run background jobs in ZenoEngine natively.</p><h2 id="configuring-the-worker-pool" tabindex="-1">Configuring the Worker Pool <a class="header-anchor" href="#configuring-the-worker-pool" aria-label="Permalink to &quot;Configuring the Worker Pool&quot;">​</a></h2><p>By default, ZenoEngine boots a worker pool with 5 concurrent workers. You can adjust this configuration when setting up your engine, typically in <code>main.zl</code>.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// main.zl</span></span>
<span class="line"><span>worker.config: {</span></span>
<span class="line"><span>    workers: 10              // Max concurrent Goroutines for jobs</span></span>
<span class="line"><span>    max_queue_size: 1000     // How many jobs can be queued before blocking</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="enqueuing-a-job" tabindex="-1">Enqueuing a Job <a class="header-anchor" href="#enqueuing-a-job" aria-label="Permalink to &quot;Enqueuing a Job&quot;">​</a></h2><p>To push logic to the background, use the <code>job.enqueue</code> slot. Any code inside the <code>do</code> block will be handed off to the worker pool and executed asynchronously.</p><p>The <code>job.enqueue</code> block immediately returns, allowing your HTTP response to complete instantly while the work happens behind the scenes.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>http.post: &#39;/api/register&#39; {</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        http.json_body: { as: $user }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // 1. Insert user into Database (Fast)</span></span>
<span class="line"><span>        db.table: &quot;users&quot;</span></span>
<span class="line"><span>        db.insert: $user</span></span>
<span class="line"><span></span></span>
<span class="line"><span>        // 2. Send Welcome Email (Slow)</span></span>
<span class="line"><span>        job.enqueue: {</span></span>
<span class="line"><span>            do: {</span></span>
<span class="line"><span>                // This payload captures variables from the surrounding scope</span></span>
<span class="line"><span>                $emailData: {</span></span>
<span class="line"><span>                    to: $user.email</span></span>
<span class="line"><span>                    subject: &quot;Welcome!&quot;</span></span>
<span class="line"><span>                    body: &quot;We are glad you are here.&quot;</span></span>
<span class="line"><span>                }</span></span>
<span class="line"><span>                </span></span>
<span class="line"><span>                // Simulate an expensive external API call</span></span>
<span class="line"><span>                http.post: &quot;https://api.mailgun.net/v3/...&quot; {</span></span>
<span class="line"><span>                    body: $emailData</span></span>
<span class="line"><span>                }</span></span>
<span class="line"><span>                </span></span>
<span class="line"><span>                log: &quot;Background email sent to &quot; + $user.email</span></span>
<span class="line"><span>            }</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span></span></span>
<span class="line"><span>        // 3. Respond instantly</span></span>
<span class="line"><span>        http.ok: { message: &quot;Account created! Check your email.&quot; }</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><h3 id="scope-capture" tabindex="-1">Scope Capture <a class="header-anchor" href="#scope-capture" aria-label="Permalink to &quot;Scope Capture&quot;">​</a></h3><p>When you enqueue a job, ZenoEngine automatically isolates and captures the <em>current state</em> of variables inside the <code>do</code> block. The background execution runs in an independent, memory-safe context, protecting you from common concurrency bugs.</p><p><em>Note: Since the background job outlives the original HTTP request, any database updates or API calls inside the job must manage their own connections if they rely on request-specific lifecycles.</em></p>`,13)])])}const b=a(p,[["render",t]]);export{h as __pageData,b as default};
