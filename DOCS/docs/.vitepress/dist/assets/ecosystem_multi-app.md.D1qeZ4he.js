import{_ as a,o as n,c as e,ag as p}from"./chunks/framework.ePeAWSvT.js";const u=JSON.parse('{"title":"Multi-App Architecture","description":"","frontmatter":{},"headers":[],"relativePath":"ecosystem/multi-app.md","filePath":"ecosystem/multi-app.md"}'),t={name:"ecosystem/multi-app.md"};function i(l,s,o,c,d,r){return n(),e("div",null,[...s[0]||(s[0]=[p(`<h1 id="multi-app-architecture" tabindex="-1">Multi-App Architecture <a class="header-anchor" href="#multi-app-architecture" aria-label="Permalink to &quot;Multi-App Architecture&quot;">​</a></h1><h2 id="overview" tabindex="-1">Overview <a class="header-anchor" href="#overview" aria-label="Permalink to &quot;Overview&quot;">​</a></h2><p>One of ZenoEngine&#39;s most powerful features is its ability to host <strong>multiple independent applications</strong> within a single engine instance and binary. Rather than running separate processes or servers, all apps share one runtime.</p><div class="language- vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang"></span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>One binary → Multiple apps → Virtual hosting / path-based routing</span></span></code></pre></div><h2 id="how-it-works" tabindex="-1">How It Works <a class="header-anchor" href="#how-it-works" aria-label="Permalink to &quot;How It Works&quot;">​</a></h2><p>ZenoEngine uses three independent primitives for multi-app isolation:</p><table tabindex="0"><thead><tr><th>Primitive</th><th>What It Does</th></tr></thead><tbody><tr><td><a href="#modules-with-include"><code>include:</code></a></td><td>Loads any <code>.zl</code> file, enabling free-form modular organization</td></tr><tr><td><a href="#per-app-blade-views"><code>view.root:</code></a></td><td>Sets the Blade view root directory for the current app</td></tr><tr><td><a href="#per-app-static-assets"><code>http.static:</code></a></td><td>Serves static files from any directory at any URL path</td></tr></tbody></table><h2 id="entry-point" tabindex="-1">Entry Point <a class="header-anchor" href="#entry-point" aria-label="Permalink to &quot;Entry Point&quot;">​</a></h2><p>Every ZenoEngine project has exactly one global entry point:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// src/main.zl — bootstraps all apps</span></span>
<span class="line"><span>http.static: &#39;shared/public&#39; {</span></span>
<span class="line"><span>    path: &#39;/assets&#39;</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>include: apps/blog/routes/web.zl</span></span>
<span class="line"><span>include: apps/shop/routes/web.zl</span></span>
<span class="line"><span>include: apps/admin/main.zl</span></span></code></pre></div><h2 id="modules-with-include" tabindex="-1">Modules with <code>include:</code> <a class="header-anchor" href="#modules-with-include" aria-label="Permalink to &quot;Modules with \`include:\`&quot;">​</a></h2><p>The <code>include:</code> slot loads any ZenoLang file. There are no restrictions on the path, allowing complete freedom in organizing your code.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// You can include from anywhere</span></span>
<span class="line"><span>include: apps/blog/routes/web.zl</span></span>
<span class="line"><span>include: domain/products/routes.zl</span></span>
<span class="line"><span>include: features/auth/login.zl</span></span></code></pre></div><h2 id="per-app-blade-views" tabindex="-1">Per-App Blade Views <a class="header-anchor" href="#per-app-blade-views" aria-label="Permalink to &quot;Per-App Blade Views&quot;">​</a></h2><p>By default, the <code>view:</code> slot looks for templates in the <code>views/</code> directory. Use <code>view.root:</code> to override this for each app:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// apps/blog/routes/web.zl</span></span>
<span class="line"><span>view.root: &#39;apps/blog/resources/views&#39;</span></span>
<span class="line"><span></span></span>
<span class="line"><span>http.get: &#39;/blog&#39; {</span></span>
<span class="line"><span>    view: &#39;index&#39;</span></span>
<span class="line"><span>    // → apps/blog/resources/views/index.blade.zl ✅</span></span>
<span class="line"><span>}</span></span></code></pre></div><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// apps/shop/routes/web.zl</span></span>
<span class="line"><span>view.root: &#39;apps/shop/resources/views&#39;</span></span>
<span class="line"><span></span></span>
<span class="line"><span>http.get: &#39;/shop&#39; {</span></span>
<span class="line"><span>    view: &#39;catalog&#39;</span></span>
<span class="line"><span>    // → apps/shop/resources/views/catalog.blade.zl ✅</span></span>
<span class="line"><span>}</span></span></code></pre></div><div class="info custom-block"><p class="custom-block-title">Scope Isolation</p><p><code>view.root:</code> is scoped to the current execution context. The blog&#39;s <code>view.root</code> does not affect the shop&#39;s views, and vice versa.</p></div><p><code>view.root:</code> applies to all Blade operations in its scope: <code>view:</code>, <code>@extends</code>, <code>@include</code>, <code>@component</code>, and <code>meta.template</code>.</p><h2 id="per-app-static-assets" tabindex="-1">Per-App Static Assets <a class="header-anchor" href="#per-app-static-assets" aria-label="Permalink to &quot;Per-App Static Assets&quot;">​</a></h2><p>Each app can serve its own CSS, JavaScript, and images from its own public directory:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// apps/blog/routes/web.zl</span></span>
<span class="line"><span>http.static: &#39;apps/blog/public&#39; {</span></span>
<span class="line"><span>    path: &#39;/blog/assets&#39;</span></span>
<span class="line"><span>}</span></span></code></pre></div><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// apps/shop/routes/web.zl</span></span>
<span class="line"><span>http.static: &#39;apps/shop/public&#39; {</span></span>
<span class="line"><span>    path: &#39;/shop/assets&#39;</span></span>
<span class="line"><span>}</span></span></code></pre></div><p>Reference them in Blade templates:</p><div class="language-blade vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">blade</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">&lt;</span><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">link</span><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;"> rel</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;stylesheet&quot;</span><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;"> href</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;/blog/assets/css/app.css&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">&gt;</span></span></code></pre></div><p>For assets shared across all apps, serve a shared directory from <code>src/main.zl</code>:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// src/main.zl</span></span>
<span class="line"><span>http.static: &#39;shared/public&#39; {</span></span>
<span class="line"><span>    path: &#39;/assets&#39;</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="virtual-hosting-multi-domain" tabindex="-1">Virtual Hosting (Multi-Domain) <a class="header-anchor" href="#virtual-hosting-multi-domain" aria-label="Permalink to &quot;Virtual Hosting (Multi-Domain)&quot;">​</a></h2><p>Use <code>http.host:</code> to route different domains to different apps within the same engine:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>http.host: &#39;blog.mysite.com&#39; { do: {</span></span>
<span class="line"><span>    include: apps/blog/routes/web.zl</span></span>
<span class="line"><span>}}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>http.host: &#39;shop.mysite.com&#39; { do: {</span></span>
<span class="line"><span>    include: apps/shop/routes/web.zl</span></span>
<span class="line"><span>}}</span></span></code></pre></div><h2 id="complete-example" tabindex="-1">Complete Example <a class="header-anchor" href="#complete-example" aria-label="Permalink to &quot;Complete Example&quot;">​</a></h2><div class="language-text vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">text</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>src/</span></span>
<span class="line"><span>└── main.zl</span></span>
<span class="line"><span></span></span>
<span class="line"><span>apps/</span></span>
<span class="line"><span>├── blog/</span></span>
<span class="line"><span>│   ├── routes/</span></span>
<span class="line"><span>│   │   ├── web.zl        ← view.root + http.static + routes</span></span>
<span class="line"><span>│   │   └── api.zl</span></span>
<span class="line"><span>│   ├── resources/views/  ← Blade templates</span></span>
<span class="line"><span>│   └── public/           ← CSS, JS, images</span></span>
<span class="line"><span>│</span></span>
<span class="line"><span>├── shop/</span></span>
<span class="line"><span>│   ├── routes/</span></span>
<span class="line"><span>│   │   ├── web.zl</span></span>
<span class="line"><span>│   │   └── api.zl</span></span>
<span class="line"><span>│   ├── resources/views/</span></span>
<span class="line"><span>│   └── public/</span></span>
<span class="line"><span>│</span></span>
<span class="line"><span>└── shared/</span></span>
<span class="line"><span>    └── public/           ← Shared fonts, icons, global CSS</span></span></code></pre></div><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// src/main.zl</span></span>
<span class="line"><span>http.static: &#39;shared/public&#39; { path: &#39;/assets&#39; }</span></span>
<span class="line"><span></span></span>
<span class="line"><span>include: apps/blog/routes/web.zl</span></span>
<span class="line"><span>include: apps/shop/routes/web.zl</span></span></code></pre></div><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// apps/blog/routes/web.zl</span></span>
<span class="line"><span>view.root: &#39;apps/blog/resources/views&#39;</span></span>
<span class="line"><span></span></span>
<span class="line"><span>http.static: &#39;apps/blog/public&#39; { path: &#39;/blog/assets&#39; }</span></span>
<span class="line"><span></span></span>
<span class="line"><span>http.get: &#39;/blog&#39; {</span></span>
<span class="line"><span>    view: &#39;index&#39;</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>http.post: &#39;/blog/posts&#39; {</span></span>
<span class="line"><span>    validate { title: &#39;required&#39; }</span></span>
<span class="line"><span>    orm.model: &#39;posts&#39;</span></span>
<span class="line"><span>    orm.save: $form</span></span>
<span class="line"><span>    redirect: &#39;/blog&#39;</span></span>
<span class="line"><span>}</span></span></code></pre></div>`,34)])])}const b=a(t,[["render",i]]);export{u as __pageData,b as default};
