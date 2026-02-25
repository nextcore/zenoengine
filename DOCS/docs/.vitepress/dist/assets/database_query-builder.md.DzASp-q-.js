import{_ as s,o as n,c as e,ag as p}from"./chunks/framework.ePeAWSvT.js";const u=JSON.parse('{"title":"Query Builder","description":"","frontmatter":{},"headers":[],"relativePath":"database/query-builder.md","filePath":"database/query-builder.md"}'),l={name:"database/query-builder.md"};function i(t,a,o,r,c,d){return n(),e("div",null,[...a[0]||(a[0]=[p(`<h1 id="query-builder" tabindex="-1">Query Builder <a class="header-anchor" href="#query-builder" aria-label="Permalink to &quot;Query Builder&quot;">​</a></h1><h2 id="introduction" tabindex="-1">Introduction <a class="header-anchor" href="#introduction" aria-label="Permalink to &quot;Introduction&quot;">​</a></h2><p>ZenoEngine&#39;s database query builder provides a convenient, fluent interface to creating and running database queries. It can be used to perform most database operations in your application and works perfectly with all database systems supported by ZenoEngine.</p><p>The query builder uses PDO parameter binding to protect your application against SQL injection attacks. There is no need to clean or sanitize strings passed to the query builder as query bindings.</p><h2 id="retrieving-results" tabindex="-1">Retrieving Results <a class="header-anchor" href="#retrieving-results" aria-label="Permalink to &quot;Retrieving Results&quot;">​</a></h2><h3 id="retrieving-all-rows-from-a-table" tabindex="-1">Retrieving All Rows From a Table <a class="header-anchor" href="#retrieving-all-rows-from-a-table" aria-label="Permalink to &quot;Retrieving All Rows From a Table&quot;">​</a></h3><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.get { as: $users }</span></span></code></pre></div><h3 id="retrieving-a-single-row" tabindex="-1">Retrieving A Single Row <a class="header-anchor" href="#retrieving-a-single-row" aria-label="Permalink to &quot;Retrieving A Single Row&quot;">​</a></h3><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.where {</span></span>
<span class="line"><span>    col: &#39;name&#39;</span></span>
<span class="line"><span>    val: &#39;John&#39;</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.first { as: $user }</span></span></code></pre></div><h2 id="where-clauses" tabindex="-1">Where Clauses <a class="header-anchor" href="#where-clauses" aria-label="Permalink to &quot;Where Clauses&quot;">​</a></h2><h3 id="basic-where-clauses" tabindex="-1">Basic Where Clauses <a class="header-anchor" href="#basic-where-clauses" aria-label="Permalink to &quot;Basic Where Clauses&quot;">​</a></h3><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.where {</span></span>
<span class="line"><span>    col: &#39;votes&#39;</span></span>
<span class="line"><span>    op: &#39;&gt;&#39;</span></span>
<span class="line"><span>    val: 100</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.get { as: $users }</span></span></code></pre></div><h3 id="or-where-clauses" tabindex="-1">Or Where Clauses <a class="header-anchor" href="#or-where-clauses" aria-label="Permalink to &quot;Or Where Clauses&quot;">​</a></h3><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.where {</span></span>
<span class="line"><span>    col: &#39;votes&#39;</span></span>
<span class="line"><span>    op: &#39;&gt;&#39;</span></span>
<span class="line"><span>    val: 100</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.or_where {</span></span>
<span class="line"><span>    col: &#39;name&#39;</span></span>
<span class="line"><span>    val: &#39;Dayle&#39;</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.get { as: $result }</span></span></code></pre></div><h3 id="where-between" tabindex="-1">Where Between <a class="header-anchor" href="#where-between" aria-label="Permalink to &quot;Where Between&quot;">​</a></h3><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;orders&#39;</span></span>
<span class="line"><span>db.where_between {</span></span>
<span class="line"><span>    col: &#39;price&#39;</span></span>
<span class="line"><span>    val: [100, 500]</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.get { as: $orders }</span></span></code></pre></div><h3 id="where-in" tabindex="-1">Where In <a class="header-anchor" href="#where-in" aria-label="Permalink to &quot;Where In&quot;">​</a></h3><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.where {</span></span>
<span class="line"><span>    col: &#39;id&#39;</span></span>
<span class="line"><span>    op: &#39;IN&#39;</span></span>
<span class="line"><span>    val: [1, 2, 3]</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.get { as: $users }</span></span></code></pre></div><h2 id="ordering-grouping-limit-offset" tabindex="-1">Ordering, Grouping, Limit &amp; Offset <a class="header-anchor" href="#ordering-grouping-limit-offset" aria-label="Permalink to &quot;Ordering, Grouping, Limit &amp; Offset&quot;">​</a></h2><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.order_by {</span></span>
<span class="line"><span>    col: &#39;name&#39;</span></span>
<span class="line"><span>    dir: &#39;desc&#39;</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.limit: 10</span></span>
<span class="line"><span>db.offset: 20</span></span>
<span class="line"><span>db.get { as: $users }</span></span></code></pre></div><h2 id="aggregates" tabindex="-1">Aggregates <a class="header-anchor" href="#aggregates" aria-label="Permalink to &quot;Aggregates&quot;">​</a></h2><p>The query builder also provides a variety of methods for retrieving aggregate values like <code>count</code>, <code>max</code>, <code>min</code>, <code>avg</code>, and <code>sum</code>.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;orders&#39;</span></span>
<span class="line"><span>db.count { as: $total }</span></span>
<span class="line"><span></span></span>
<span class="line"><span>db.table: &#39;orders&#39;</span></span>
<span class="line"><span>db.max {</span></span>
<span class="line"><span>    col: &#39;price&#39;</span></span>
<span class="line"><span>    as: $maxPrice</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>db.table: &#39;orders&#39;</span></span>
<span class="line"><span>db.sum {</span></span>
<span class="line"><span>    col: &#39;price&#39;</span></span>
<span class="line"><span>    as: $totalRevenue</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="joins" tabindex="-1">Joins <a class="header-anchor" href="#joins" aria-label="Permalink to &quot;Joins&quot;">​</a></h2><h3 id="inner-join-clause" tabindex="-1">Inner Join Clause <a class="header-anchor" href="#inner-join-clause" aria-label="Permalink to &quot;Inner Join Clause&quot;">​</a></h3><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.join {</span></span>
<span class="line"><span>    table: &#39;contacts&#39;</span></span>
<span class="line"><span>    on: [&#39;users.id&#39;, &#39;=&#39;, &#39;contacts.user_id&#39;]</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.get { as: $result }</span></span></code></pre></div><h3 id="left-join-clause" tabindex="-1">Left Join Clause <a class="header-anchor" href="#left-join-clause" aria-label="Permalink to &quot;Left Join Clause&quot;">​</a></h3><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.left_join {</span></span>
<span class="line"><span>    table: &#39;posts&#39;</span></span>
<span class="line"><span>    on: [&#39;users.id&#39;, &#39;=&#39;, &#39;posts.user_id&#39;]</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.get { as: $result }</span></span></code></pre></div><h2 id="insert-statements" tabindex="-1">Insert Statements <a class="header-anchor" href="#insert-statements" aria-label="Permalink to &quot;Insert Statements&quot;">​</a></h2><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.insert {</span></span>
<span class="line"><span>    name: &#39;Alice&#39;</span></span>
<span class="line"><span>    email: &#39;alice@example.com&#39;</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="update-statements" tabindex="-1">Update Statements <a class="header-anchor" href="#update-statements" aria-label="Permalink to &quot;Update Statements&quot;">​</a></h2><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.where {</span></span>
<span class="line"><span>    col: &#39;id&#39;</span></span>
<span class="line"><span>    val: 1</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.update {</span></span>
<span class="line"><span>    name: &#39;Alice Updated&#39;</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="delete-statements" tabindex="-1">Delete Statements <a class="header-anchor" href="#delete-statements" aria-label="Permalink to &quot;Delete Statements&quot;">​</a></h2><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.where {</span></span>
<span class="line"><span>    col: &#39;id&#39;</span></span>
<span class="line"><span>    val: 1</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.delete</span></span></code></pre></div><h2 id="pagination" tabindex="-1">Pagination <a class="header-anchor" href="#pagination" aria-label="Permalink to &quot;Pagination&quot;">​</a></h2><p>ZenoEngine makes it easy to paginate results, returning both the results and metadata automatically.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.paginate {</span></span>
<span class="line"><span>    per_page: 15</span></span>
<span class="line"><span>    page: 1</span></span>
<span class="line"><span>    as: $paginator</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>// $paginator.data = list of items</span></span>
<span class="line"><span>// $paginator.total = total records</span></span>
<span class="line"><span>// $paginator.last_page = number of pages</span></span>
<span class="line"><span>// $paginator.current_page = current page</span></span></code></pre></div><h2 id="checking-existence" tabindex="-1">Checking Existence <a class="header-anchor" href="#checking-existence" aria-label="Permalink to &quot;Checking Existence&quot;">​</a></h2><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>db.table: &#39;users&#39;</span></span>
<span class="line"><span>db.where {</span></span>
<span class="line"><span>    col: &#39;email&#39;</span></span>
<span class="line"><span>    val: &#39;alice@example.com&#39;</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span>db.exists { as: $hasUser }</span></span></code></pre></div>`,39)])])}const b=s(l,[["render",i]]);export{u as __pageData,b as default};
