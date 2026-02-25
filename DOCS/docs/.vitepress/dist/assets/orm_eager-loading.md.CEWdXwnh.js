import{_ as s,o as n,c as e,ag as t}from"./chunks/framework.ePeAWSvT.js";const u=JSON.parse('{"title":"Eager Loading","description":"","frontmatter":{},"headers":[],"relativePath":"orm/eager-loading.md","filePath":"orm/eager-loading.md"}'),o={name:"orm/eager-loading.md"};function p(l,a,i,r,c,d){return n(),e("div",null,[...a[0]||(a[0]=[t(`<h1 id="eager-loading" tabindex="-1">Eager Loading <a class="header-anchor" href="#eager-loading" aria-label="Permalink to &quot;Eager Loading&quot;">​</a></h1><h2 id="introduction" tabindex="-1">Introduction <a class="header-anchor" href="#introduction" aria-label="Permalink to &quot;Introduction&quot;">​</a></h2><p>When accessing Eloquent relationships, the relationship data is <strong>lazy loaded</strong> by default. This means the relationship data is not actually loaded until you access the property. However, this creates the classic <strong>N+1 query problem</strong>.</p><h2 id="the-n-1-problem" tabindex="-1">The N+1 Problem <a class="header-anchor" href="#the-n-1-problem" aria-label="Permalink to &quot;The N+1 Problem&quot;">​</a></h2><p>Imagine you have a post with authors:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// ❌ BAD: This causes N+1 queries!</span></span>
<span class="line"><span>// 1 query for posts, then 1 query PER POST for the user</span></span>
<span class="line"><span>orm.model: &#39;posts&#39;</span></span>
<span class="line"><span>db.get { as: $posts }</span></span>
<span class="line"><span></span></span>
<span class="line"><span>@foreach($posts as $post)</span></span>
<span class="line"><span>    // This triggers a new DB query for EACH post!</span></span>
<span class="line"><span>    print: $post.user.name</span></span>
<span class="line"><span>@endforeach</span></span></code></pre></div><h2 id="solving-n-1-with-orm-with" tabindex="-1">Solving N+1 with <code>orm.with</code> <a class="header-anchor" href="#solving-n-1-with-orm-with" aria-label="Permalink to &quot;Solving N+1 with \`orm.with\`&quot;">​</a></h2><p>ZenoEngine solves this natively using <code>orm.with</code>, which executes exactly <strong>2 queries total</strong> regardless of the number of parent records:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// ✅ GOOD: Only 2 SQL queries total, no matter how many posts</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// Step 1: Define model with relationship</span></span>
<span class="line"><span>orm.model: &#39;posts&#39; {</span></span>
<span class="line"><span>    orm.belongsTo: &#39;users&#39; {</span></span>
<span class="line"><span>        as: &#39;author&#39;</span></span>
<span class="line"><span>        foreign_key: &#39;user_id&#39;</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// Step 2: Fetch all posts</span></span>
<span class="line"><span>orm.model: &#39;posts&#39;</span></span>
<span class="line"><span>db.get { as: $posts }</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// Step 3: Eager load authors in a SINGLE query</span></span>
<span class="line"><span>orm.model: &#39;posts&#39;</span></span>
<span class="line"><span>orm.with: &#39;author&#39; {</span></span>
<span class="line"><span>    set: $posts { val: $posts }</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// Now each post has $post.author with no extra queries!</span></span>
<span class="line"><span>@foreach($posts as $post)</span></span>
<span class="line"><span>    print: $post.author.name</span></span>
<span class="line"><span>@endforeach</span></span></code></pre></div><h2 id="how-it-works-internally" tabindex="-1">How it Works Internally <a class="header-anchor" href="#how-it-works-internally" aria-label="Permalink to &quot;How it Works Internally&quot;">​</a></h2><p><code>orm.with</code> follows a 3-step process identical to Laravel&#39;s eager loading:</p><ol><li><strong>Collect Keys</strong>: Scan all parent records and collect unique foreign key values (e.g., all <code>user_id</code> values from <code>$posts</code>).</li><li><strong>Single Batch Query</strong>: Execute one optimized <code>WHERE id IN (...)</code> query to fetch all related records at once.</li><li><strong>Map In Memory</strong>: Match and attach related records back to the correct parent objects in memory using a dictionary (O(1) lookup).</li></ol><p>This is the same algorithm Laravel uses internally in <code>Eloquent::with()</code>.</p><div class="info custom-block"><p class="custom-block-title">Query Count Comparison</p><table tabindex="0"><thead><tr><th>Approach</th><th>Queries for 1000 posts</th></tr></thead><tbody><tr><td>Lazy loading</td><td>1001 queries</td></tr><tr><td><code>orm.with</code> Eager Loading</td><td><strong>2 queries</strong></td></tr></tbody></table></div>`,14)])])}const g=s(o,[["render",p]]);export{u as __pageData,g as default};
