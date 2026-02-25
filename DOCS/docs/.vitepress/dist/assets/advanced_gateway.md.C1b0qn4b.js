import{_ as s,o as n,c as e,ag as t}from"./chunks/framework.ePeAWSvT.js";const u=JSON.parse('{"title":"Web Server Gateway & Virtual Hosts","description":"","frontmatter":{},"headers":[],"relativePath":"advanced/gateway.md","filePath":"advanced/gateway.md"}'),p={name:"advanced/gateway.md"};function o(i,a,l,r,c,d){return n(),e("div",null,[...a[0]||(a[0]=[t(`<h1 id="web-server-gateway-virtual-hosts" tabindex="-1">Web Server Gateway &amp; Virtual Hosts <a class="header-anchor" href="#web-server-gateway-virtual-hosts" aria-label="Permalink to &quot;Web Server Gateway &amp; Virtual Hosts&quot;">​</a></h1><p>ZenoEngine is designed not just as an application framework but as a <strong>full Web Server Gateway</strong>. You don&#39;t necessarily need to put it behind Nginx or Caddy. ZenoEngine provides built-in slots to act as a reverse proxy, host static single-page applications (SPAs), and manage multi-tenant domains.</p><h2 id="virtual-hosting" tabindex="-1">Virtual Hosting <a class="header-anchor" href="#virtual-hosting" aria-label="Permalink to &quot;Virtual Hosting&quot;">​</a></h2><p>You can route traffic based on the incoming domain name using the <code>http.host</code> slot. This is the cornerstone of ZenoEngine&#39;s &quot;Multi-App Architecture&quot;, allowing one single Go binary to host dozens of distinct applications.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// main.zl</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// App 1: API Server</span></span>
<span class="line"><span>http.host: &quot;api.mycompany.com&quot; {</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        http.get: &#39;/v1/users&#39; { ... }</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// App 2: Landing Page</span></span>
<span class="line"><span>http.host: &quot;www.mycompany.com&quot; {</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        http.get: &#39;/&#39; { ... }</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="static-asset-spa-hosting" tabindex="-1">Static Asset &amp; SPA Hosting <a class="header-anchor" href="#static-asset-spa-hosting" aria-label="Permalink to &quot;Static Asset &amp; SPA Hosting&quot;">​</a></h2><p>Instead of using Nginx to serve your Vue, React, or Svelte applications, you can use ZenoEngine&#39;s <code>http.static</code> slot. This slot securely serves files from a given directory and protects against path traversal attacks.</p><p>If you are hosting a Single Page Application (SPA) where the frontend router needs to take over, add the <code>spa: true</code> flag. This ensures that any 404s will automatically return <code>index.html</code> instead.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// Host a React App from the /dist folder</span></span>
<span class="line"><span>http.host: &quot;app.mydomain.com&quot; {</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        // Serve everything under / prefix from the ./frontend/dist directory</span></span>
<span class="line"><span>        http.static: &quot;./frontend/dist&quot; {</span></span>
<span class="line"><span>            path: &quot;/&quot;</span></span>
<span class="line"><span>            spa: true</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// Host a regular folder of images</span></span>
<span class="line"><span>http.static: &quot;./storage/images&quot; {</span></span>
<span class="line"><span>    path: &quot;/images/&quot;</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="reverse-proxying-caddy-style" tabindex="-1">Reverse Proxying (Caddy-style) <a class="header-anchor" href="#reverse-proxying-caddy-style" aria-label="Permalink to &quot;Reverse Proxying (Caddy-style)&quot;">​</a></h2><p>Sometimes you have a legacy API in Node.js, Python, or even PHP that you want to host on the same domain as your new ZenoEngine built features. You can seamlessly proxy traffic to those services using <code>http.proxy</code>.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>http.host: &quot;api.mydomain.com&quot; {</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        // ZenoEngine handles these new endpoints directly</span></span>
<span class="line"><span>        http.get: &#39;/v2/fast-search&#39; { ... }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // EVERYTHING else gets securely proxied to a legacy Node.js app</span></span>
<span class="line"><span>        http.proxy: &quot;http://localhost:4000&quot; {</span></span>
<span class="line"><span>            path: &quot;/&quot;</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Proxy a specific path to another internal service</span></span>
<span class="line"><span>        http.proxy: &quot;http://10.0.0.5:8080&quot; {</span></span>
<span class="line"><span>            path: &quot;/ai-models/&quot;</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><p>The <code>http.proxy</code> automatically preserves HTTP headers, handles streaming, and propagates the correct <code>X-Forwarded-*</code> headers, just like Nginx or Caddy.</p>`,13)])])}const g=s(p,[["render",o]]);export{u as __pageData,g as default};
