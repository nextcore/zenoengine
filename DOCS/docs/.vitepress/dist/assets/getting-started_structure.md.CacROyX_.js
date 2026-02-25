import{_ as a,o as n,c as e,ag as p}from"./chunks/framework.ePeAWSvT.js";const u=JSON.parse('{"title":"Directory Structure","description":"","frontmatter":{},"headers":[],"relativePath":"getting-started/structure.md","filePath":"getting-started/structure.md"}'),l={name:"getting-started/structure.md"};function t(i,s,o,r,c,d){return n(),e("div",null,[...s[0]||(s[0]=[p(`<h1 id="directory-structure" tabindex="-1">Directory Structure <a class="header-anchor" href="#directory-structure" aria-label="Permalink to &quot;Directory Structure&quot;">​</a></h1><h2 id="introduction" tabindex="-1">Introduction <a class="header-anchor" href="#introduction" aria-label="Permalink to &quot;Introduction&quot;">​</a></h2><p>ZenoEngine is <strong>&quot;structure-agnostic&quot;</strong> — it does not enforce any particular directory layout. Instead, it provides a flexible set of primitives (<code>include:</code>, <code>view.root:</code>, <code>http.static:</code>) that allow you to build your project in any style you prefer.</p><h2 id="the-one-rule" tabindex="-1">The One Rule <a class="header-anchor" href="#the-one-rule" aria-label="Permalink to &quot;The One Rule&quot;">​</a></h2><p>The only thing ZenoEngine requires is a single entry point:</p><div class="language-text vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">text</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>src/main.zl        ← Always the starting point</span></span></code></pre></div><p>Everything else is completely up to you.</p><h2 id="structure-styles" tabindex="-1">Structure Styles <a class="header-anchor" href="#structure-styles" aria-label="Permalink to &quot;Structure Styles&quot;">​</a></h2><h3 id="_1-laravel-style-recommended-for-single-apps" tabindex="-1">1. Laravel Style (Recommended for single apps) <a class="header-anchor" href="#_1-laravel-style-recommended-for-single-apps" aria-label="Permalink to &quot;1. Laravel Style (Recommended for single apps)&quot;">​</a></h3><p>Familiar to any Laravel developer. Great for teams migrating from PHP.</p><div class="language-text vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">text</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>src/main.zl</span></span>
<span class="line"><span>routes/</span></span>
<span class="line"><span>├── web.zl          # Web routes (HTML responses)</span></span>
<span class="line"><span>└── api.zl          # API routes (JSON responses)</span></span>
<span class="line"><span>resources/</span></span>
<span class="line"><span>└── views/          # Blade template files</span></span>
<span class="line"><span>database/</span></span>
<span class="line"><span>├── migrations/</span></span>
<span class="line"><span>└── seeders/</span></span>
<span class="line"><span>public/             # CSS, JS, images</span></span></code></pre></div><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// src/main.zl</span></span>
<span class="line"><span>include: routes/web.zl</span></span>
<span class="line"><span>include: routes/api.zl</span></span></code></pre></div><hr><h3 id="_2-multi-app-modular-recommended-for-hosting-multiple-apps" tabindex="-1">2. Multi-App / Modular (Recommended for hosting multiple apps) <a class="header-anchor" href="#_2-multi-app-modular-recommended-for-hosting-multiple-apps" aria-label="Permalink to &quot;2. Multi-App / Modular (Recommended for hosting multiple apps)&quot;">​</a></h3><p>ZenoEngine can host <strong>multiple independent applications</strong> in a single instance. Each app lives in its own directory and can use any internal style.</p><div class="language-text vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">text</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>src/</span></span>
<span class="line"><span>└── main.zl              ← Loads all apps</span></span>
<span class="line"><span></span></span>
<span class="line"><span>apps/</span></span>
<span class="line"><span>├── blog/                ← App 1</span></span>
<span class="line"><span>│   ├── routes/</span></span>
<span class="line"><span>│   │   ├── web.zl</span></span>
<span class="line"><span>│   │   └── api.zl</span></span>
<span class="line"><span>│   ├── resources/</span></span>
<span class="line"><span>│   │   └── views/       ← Blog&#39;s own Blade templates</span></span>
<span class="line"><span>│   └── public/          ← Blog&#39;s own CSS/JS</span></span>
<span class="line"><span>│</span></span>
<span class="line"><span>├── shop/                ← App 2</span></span>
<span class="line"><span>│   ├── routes/</span></span>
<span class="line"><span>│   │   ├── web.zl</span></span>
<span class="line"><span>│   │   └── api.zl</span></span>
<span class="line"><span>│   ├── resources/</span></span>
<span class="line"><span>│   │   └── views/       ← Shop&#39;s own Blade templates</span></span>
<span class="line"><span>│   └── public/          ← Shop&#39;s own CSS/JS</span></span>
<span class="line"><span>│</span></span>
<span class="line"><span>└── shared/</span></span>
<span class="line"><span>    └── public/          ← Shared assets (fonts, icons, etc.)</span></span></code></pre></div><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// src/main.zl</span></span>
<span class="line"><span>http.static: &#39;shared/public&#39; {</span></span>
<span class="line"><span>    path: &#39;/assets&#39;</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>include: apps/blog/routes/web.zl</span></span>
<span class="line"><span>include: apps/shop/routes/web.zl</span></span></code></pre></div><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// apps/blog/routes/web.zl</span></span>
<span class="line"><span>view.root: &#39;apps/blog/resources/views&#39;   // ← isolate views per app</span></span>
<span class="line"><span></span></span>
<span class="line"><span>http.static: &#39;apps/blog/public&#39; {</span></span>
<span class="line"><span>    path: &#39;/blog/assets&#39;</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>http.get: &#39;/blog&#39; {</span></span>
<span class="line"><span>    view: &#39;index&#39;   // → apps/blog/resources/views/index.blade.zl</span></span>
<span class="line"><span>}</span></span></code></pre></div><hr><h3 id="_3-domain-driven-design-ddd" tabindex="-1">3. Domain-Driven Design (DDD) <a class="header-anchor" href="#_3-domain-driven-design-ddd" aria-label="Permalink to &quot;3. Domain-Driven Design (DDD)&quot;">​</a></h3><div class="language-text vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">text</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>src/main.zl</span></span>
<span class="line"><span>domain/</span></span>
<span class="line"><span>├── users/</span></span>
<span class="line"><span>│   ├── routes.zl</span></span>
<span class="line"><span>│   └── views/</span></span>
<span class="line"><span>├── products/</span></span>
<span class="line"><span>│   ├── routes.zl</span></span>
<span class="line"><span>│   └── views/</span></span>
<span class="line"><span>└── orders/</span></span>
<span class="line"><span>    ├── routes.zl</span></span>
<span class="line"><span>    └── views/</span></span></code></pre></div><hr><h3 id="_4-feature-based-angular-next-js-style" tabindex="-1">4. Feature-Based (Angular / Next.js-style) <a class="header-anchor" href="#_4-feature-based-angular-next-js-style" aria-label="Permalink to &quot;4. Feature-Based (Angular / Next.js-style)&quot;">​</a></h3><div class="language-text vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">text</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>src/main.zl</span></span>
<span class="line"><span>features/</span></span>
<span class="line"><span>├── auth/</span></span>
<span class="line"><span>│   ├── login.zl</span></span>
<span class="line"><span>│   ├── register.zl</span></span>
<span class="line"><span>│   └── views/</span></span>
<span class="line"><span>├── dashboard/</span></span>
<span class="line"><span>│   ├── index.zl</span></span>
<span class="line"><span>│   └── views/</span></span>
<span class="line"><span>└── reports/</span></span>
<span class="line"><span>    ├── index.zl</span></span>
<span class="line"><span>    └── views/</span></span></code></pre></div><hr><h3 id="_5-micro-app-one-file-per-concern" tabindex="-1">5. Micro-App (One file per concern) <a class="header-anchor" href="#_5-micro-app-one-file-per-concern" aria-label="Permalink to &quot;5. Micro-App (One file per concern)&quot;">​</a></h3><div class="language-text vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">text</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>src/main.zl</span></span>
<span class="line"><span>apps/</span></span>
<span class="line"><span>├── landing.zl          ← entire landing page in one file</span></span>
<span class="line"><span>├── api-v1.zl           ← all v1 API endpoints</span></span>
<span class="line"><span>└── webhooks.zl         ← webhook handlers</span></span></code></pre></div><hr><h3 id="_6-mixed-styles-the-real-world" tabindex="-1">6. Mixed Styles (The real world) <a class="header-anchor" href="#_6-mixed-styles-the-real-world" aria-label="Permalink to &quot;6. Mixed Styles (The real world)&quot;">​</a></h3><p>Different apps inside the same engine can each use a different style. There is no conflict.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// src/main.zl</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// App A: Laravel-style (team of PHP devs)</span></span>
<span class="line"><span>include: apps/blog/routes/web.zl</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// App B: DDD-style (Go team)</span></span>
<span class="line"><span>include: domain/shop/routes.zl</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// App C: Single file</span></span>
<span class="line"><span>include: apps/landing.zl</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// Shared virtual host</span></span>
<span class="line"><span>http.host: &#39;api.mysite.com&#39; { do: {</span></span>
<span class="line"><span>    include: apps/api/routes.zl</span></span>
<span class="line"><span>}}</span></span></code></pre></div><h2 id="the-three-isolation-primitives" tabindex="-1">The Three Isolation Primitives <a class="header-anchor" href="#the-three-isolation-primitives" aria-label="Permalink to &quot;The Three Isolation Primitives&quot;">​</a></h2><table tabindex="0"><thead><tr><th>Primitive</th><th>Purpose</th><th>How</th></tr></thead><tbody><tr><td><code>include:</code></td><td>Load any <code>.zl</code> file from any path</td><td>Free-form, no restrictions</td></tr><tr><td><code>view.root:</code></td><td>Set Blade view directory per app</td><td>Declared at top of each app&#39;s route file</td></tr><tr><td><code>http.static:{path:}</code></td><td>Serve static files at a unique URL path</td><td>Each app specifies its own <code>root</code> and <code>path</code></td></tr></tbody></table><h2 id="multi-app-isolation-summary" tabindex="-1">Multi-App Isolation Summary <a class="header-anchor" href="#multi-app-isolation-summary" aria-label="Permalink to &quot;Multi-App Isolation Summary&quot;">​</a></h2><div class="language- vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang"></span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>src/main.zl</span></span>
<span class="line"><span>└── include: apps/blog/routes/web.zl   ← blog views → apps/blog/resources/views/</span></span>
<span class="line"><span>    include: apps/shop/routes/web.zl   ← shop views → apps/shop/resources/views/</span></span>
<span class="line"><span>    include: apps/admin/main.zl        ← admin views → apps/admin/views/</span></span></code></pre></div><p>Each app is <strong>completely independent</strong> — its own routes, views, and static assets. Yet they all run inside <strong>one process, one binary</strong>.</p>`,36)])])}const g=a(l,[["render",t]]);export{u as __pageData,g as default};
