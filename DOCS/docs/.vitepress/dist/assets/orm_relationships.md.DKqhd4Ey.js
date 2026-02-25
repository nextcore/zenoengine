import{_ as a,o as n,c as e,ag as p}from"./chunks/framework.ePeAWSvT.js";const u=JSON.parse('{"title":"ORM Relationships","description":"","frontmatter":{},"headers":[],"relativePath":"orm/relationships.md","filePath":"orm/relationships.md"}'),o={name:"orm/relationships.md"};function l(i,s,t,r,c,d){return n(),e("div",null,[...s[0]||(s[0]=[p(`<h1 id="orm-relationships" tabindex="-1">ORM Relationships <a class="header-anchor" href="#orm-relationships" aria-label="Permalink to &quot;ORM Relationships&quot;">​</a></h1><h2 id="introduction" tabindex="-1">Introduction <a class="header-anchor" href="#introduction" aria-label="Permalink to &quot;Introduction&quot;">​</a></h2><p>Database tables are often related to one another. ZenoEngine makes managing and working with these relationships easy, supporting the four relationship types you know from Laravel Eloquent.</p><h2 id="one-to-one-hasone" tabindex="-1">One To One (hasOne) <a class="header-anchor" href="#one-to-one-hasone" aria-label="Permalink to &quot;One To One (hasOne)&quot;">​</a></h2><p>A one-to-one relationship is a very basic type of database relationship. Define it using <code>orm.hasOne</code> inside your model block:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>orm.model: &#39;users&#39; {</span></span>
<span class="line"><span>    orm.hasOne: &#39;profiles&#39; {</span></span>
<span class="line"><span>        as: &#39;profile&#39;</span></span>
<span class="line"><span>        foreign_key: &#39;user_id&#39;</span></span>
<span class="line"><span>        local_key: &#39;id&#39;</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><p>Then eager load it:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>orm.model: &#39;users&#39;</span></span>
<span class="line"><span>db.get { as: $users }</span></span>
<span class="line"><span></span></span>
<span class="line"><span>orm.model: &#39;users&#39;</span></span>
<span class="line"><span>orm.with: &#39;profile&#39; {</span></span>
<span class="line"><span>    set: $users { val: $users }</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// Each user now has a $user.profile object attached</span></span></code></pre></div><h2 id="one-to-many-hasmany" tabindex="-1">One To Many (hasMany) <a class="header-anchor" href="#one-to-many-hasmany" aria-label="Permalink to &quot;One To Many (hasMany)&quot;">​</a></h2><p>One-to-many relationships are used when a single parent model owns many child models. For example, a user may have many posts:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>orm.model: &#39;users&#39; {</span></span>
<span class="line"><span>    orm.hasMany: &#39;posts&#39; {</span></span>
<span class="line"><span>        as: &#39;posts&#39;</span></span>
<span class="line"><span>        foreign_key: &#39;user_id&#39;</span></span>
<span class="line"><span>        local_key: &#39;id&#39;</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="many-to-one-belongsto" tabindex="-1">Many To One (belongsTo) <a class="header-anchor" href="#many-to-one-belongsto" aria-label="Permalink to &quot;Many To One (belongsTo)&quot;">​</a></h2><p>The inverse of <code>hasMany</code> is <code>belongsTo</code>. A post belongs to a user:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>orm.model: &#39;posts&#39; {</span></span>
<span class="line"><span>    orm.belongsTo: &#39;users&#39; {</span></span>
<span class="line"><span>        as: &#39;author&#39;</span></span>
<span class="line"><span>        foreign_key: &#39;user_id&#39;</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="many-to-many-belongstomany" tabindex="-1">Many To Many (belongsToMany) <a class="header-anchor" href="#many-to-many-belongstomany" aria-label="Permalink to &quot;Many To Many (belongsToMany)&quot;">​</a></h2><p>Many-to-many relations involve an intermediary &quot;pivot&quot; table. For example, a User may have many Roles, and Roles may be shared by many Users:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>orm.model: &#39;users&#39; {</span></span>
<span class="line"><span>    orm.belongsToMany: &#39;roles&#39; {</span></span>
<span class="line"><span>        as: &#39;roles&#39;</span></span>
<span class="line"><span>        table: &#39;role_user&#39;          // Pivot table name</span></span>
<span class="line"><span>        foreign_pivot_key: &#39;user_id&#39;</span></span>
<span class="line"><span>        related_pivot_key: &#39;role_id&#39;</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><p>Then eager load:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>orm.model: &#39;users&#39;</span></span>
<span class="line"><span>db.get { as: $users }</span></span>
<span class="line"><span></span></span>
<span class="line"><span>orm.model: &#39;users&#39;</span></span>
<span class="line"><span>orm.with: &#39;roles&#39; {</span></span>
<span class="line"><span>    set: $users { val: $users }</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// Each user now has $user.roles = [...roles array...]</span></span></code></pre></div><div class="tip custom-block"><p class="custom-block-title">N+1 Prevention</p><p>ZenoEngine&#39;s eager loading always resolves any relationship in exactly <strong>2 SQL queries</strong> total, regardless of how many parent records you have loaded — identical to Laravel&#39;s <code>with()</code> behavior.</p></div>`,20)])])}const m=a(o,[["render",l]]);export{u as __pageData,m as default};
