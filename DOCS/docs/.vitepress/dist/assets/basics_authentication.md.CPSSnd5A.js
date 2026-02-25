import{_ as e,o as l,c as p,ag as i,j as s,a}from"./chunks/framework.ePeAWSvT.js";const g=JSON.parse('{"title":"Authentication","description":"","frontmatter":{},"headers":[],"relativePath":"basics/authentication.md","filePath":"basics/authentication.md"}'),t={name:"basics/authentication.md"};function h(r,n,d,o,k,c){return l(),p("div",null,[...n[0]||(n[0]=[i(`<h1 id="authentication" tabindex="-1">Authentication <a class="header-anchor" href="#authentication" aria-label="Permalink to &quot;Authentication&quot;">​</a></h1><p>Building a secure login and registration system is one of the most common tasks when developing web applications. ZenoEngine provides built-in mechanisms for password hashing, session management, and middleware to make building authentication straightforward.</p><p>In this guide, we&#39;ll walk through building a complete, basic Authentication flow from scratch.</p><h2 id="_1-database-setup" tabindex="-1">1. Database Setup <a class="header-anchor" href="#_1-database-setup" aria-label="Permalink to &quot;1. Database Setup&quot;">​</a></h2><p>First, you need a place to store your users. Create a table using the Query Builder <code>db.query</code> or the Schema Builder (if available). The minimum fields required are an identifier (like <code>email</code>) and a robust <code>password</code> field.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// In your database setup script or migration</span></span>
<span class="line"><span>db.query: &quot;</span></span>
<span class="line"><span>    CREATE TABLE IF NOT EXISTS users (</span></span>
<span class="line"><span>        id INTEGER PRIMARY KEY AUTOINCREMENT,</span></span>
<span class="line"><span>        name TEXT NOT NULL,</span></span>
<span class="line"><span>        email TEXT UNIQUE NOT NULL,</span></span>
<span class="line"><span>        password TEXT NOT NULL,</span></span>
<span class="line"><span>        created_at DATETIME DEFAULT CURRENT_TIMESTAMP</span></span>
<span class="line"><span>    );</span></span>
<span class="line"><span>&quot;</span></span></code></pre></div><h2 id="_2-the-login-view-zenoblade" tabindex="-1">2. The Login View (ZenoBlade) <a class="header-anchor" href="#_2-the-login-view-zenoblade" aria-label="Permalink to &quot;2. The Login View (ZenoBlade)&quot;">​</a></h2><p>Let&#39;s create the HTML form for users to enter their credentials. Create a file at <code>resources/views/auth/login.blade.zl</code>.</p><p>Notice how we include a <code>{{ csrf_field() }}</code>. ZenoEngine requires all <code>POST</code> requests to have a CSRF token for security. We also use the <code>$errors</code> variable to display any validation or login failures.</p>`,9),s("div",null,[s("div",{class:"language-html vp-adaptive-theme"},[s("button",{title:"Copy Code",class:"copy"}),s("span",{class:"lang"},"html"),s("pre",{class:"shiki shiki-themes github-light github-dark vp-code",tabindex:"0","v-pre":""},[s("code",null,[s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"}},"<!-- resources/views/auth/login.blade.zl -->")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"<!"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"DOCTYPE"),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," html"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"<"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"html"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"<"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"head"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"    <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"title"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">Login</"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"title"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"</"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"head"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"<"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"body"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"    <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"main"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"        <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"h2"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">Sign In</"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"h2"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"}),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"}},"        <!-- Display authentication errors -->")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"        @if(isset($errors['auth']))")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"            <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"div"),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," style"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"="),s("span",{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"}},'"color: red; padding: 10px;"'),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"                {{ $errors['auth'] }}")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"            </"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"div"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"        @endif")]),a(`
`),s("span",{class:"line"}),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"        <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"form"),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," method"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"="),s("span",{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"}},'"POST"'),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," action"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"="),s("span",{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"}},'"/login"'),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"            {{ csrf_field() }}")]),a(`
`),s("span",{class:"line"}),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"            <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"div"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"                <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"label"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">Email</"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"label"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"                <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"input"),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," type"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"="),s("span",{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"}},'"email"'),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," name"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"="),s("span",{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"}},'"email"'),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," required"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"            </"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"div"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"}),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"            <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"div"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"                <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"label"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">Password</"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"label"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"                <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"input"),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," type"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"="),s("span",{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"}},'"password"'),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," name"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"="),s("span",{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"}},'"password"'),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," required"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"            </"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"div"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"}),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"            <"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"button"),s("span",{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"}}," type"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"="),s("span",{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"}},'"submit"'),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">Log In</"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"button"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"        </"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"form"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"    </"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"main"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"</"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"body"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")]),a(`
`),s("span",{class:"line"},[s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},"</"),s("span",{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"}},"html"),s("span",{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"}},">")])])])])],-1),i(`<h2 id="_3-handling-the-routes-zenolang" tabindex="-1">3. Handling the Routes (ZenoLang) <a class="header-anchor" href="#_3-handling-the-routes-zenolang" aria-label="Permalink to &quot;3. Handling the Routes (ZenoLang)&quot;">​</a></h2><p>Now, we need two routes: one <code>GET</code> route to display the form, and a <code>POST</code> route to process the submission.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// src/main.zl</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// 1. Show the Login Form</span></span>
<span class="line"><span>http.get: &#39;/login&#39; {</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        http.view: &#39;auth/login&#39;</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// 2. Process the Login</span></span>
<span class="line"><span>http.post: &#39;/login&#39; {</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        // Read the submitted form data</span></span>
<span class="line"><span>        http.form: { as: $credentials }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Find the user in the database</span></span>
<span class="line"><span>        db.table: &quot;users&quot;</span></span>
<span class="line"><span>        db.where: &quot;email&quot; { equals: $credentials.email }</span></span>
<span class="line"><span>        db.first: { as: $user }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // If user doesn&#39;t exist, redirect back with an error</span></span>
<span class="line"><span>        if: $user == null {</span></span>
<span class="line"><span>            then: {</span></span>
<span class="line"><span>                http.redirect: &#39;/login&#39; {</span></span>
<span class="line"><span>                    flash: { error: &quot;Invalid credentials.&quot; }</span></span>
<span class="line"><span>                }</span></span>
<span class="line"><span>                return</span></span>
<span class="line"><span>            }</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Verify the password using ZenoEngine&#39;s built-in Bcrypt hasher</span></span>
<span class="line"><span>        hash.verify: {</span></span>
<span class="line"><span>            text: $credentials.password</span></span>
<span class="line"><span>            hash: $user.password</span></span>
<span class="line"><span>            as: $isValid</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        if: $isValid == false {</span></span>
<span class="line"><span>            then: {</span></span>
<span class="line"><span>                http.redirect: &#39;/login&#39; {</span></span>
<span class="line"><span>                    flash: { error: &quot;Invalid credentials.&quot; }</span></span>
<span class="line"><span>                }</span></span>
<span class="line"><span>                return</span></span>
<span class="line"><span>            }</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Password is correct! Log the user in by saving their ID to the session.</span></span>
<span class="line"><span>        session.set: &quot;user_id&quot; { val: $user.id }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Regenerate the session ID to prevent Session Fixation attacks</span></span>
<span class="line"><span>        session.regenerate: true</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Redirect to the dashboard</span></span>
<span class="line"><span>        http.redirect: &#39;/dashboard&#39;</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="_4-protecting-routes-with-middleware" tabindex="-1">4. Protecting Routes with Middleware <a class="header-anchor" href="#_4-protecting-routes-with-middleware" aria-label="Permalink to &quot;4. Protecting Routes with Middleware&quot;">​</a></h2><p>Now that users can log in, we need to protect certain pages (like the <code>/dashboard</code>) so that only authenticated users can access them. We achieve this using <strong>Middleware</strong>.</p><p>First, define the <code>auth</code> middleware logic. If the session does not contain a <code>user_id</code>, we redirect them back to the login page.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// src/middleware/auth.zl</span></span>
<span class="line"><span>http.middleware: &#39;auth&#39; {</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        session.get: &quot;user_id&quot; { as: $userId }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // If no user_id is found in the session, deny access!</span></span>
<span class="line"><span>        if: $userId == null {</span></span>
<span class="line"><span>            then: {</span></span>
<span class="line"><span>                http.redirect: &#39;/login&#39;</span></span>
<span class="line"><span>                return</span></span>
<span class="line"><span>            }</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Optional: Load the full user object and attach it to the request</span></span>
<span class="line"><span>        db.table: &quot;users&quot;</span></span>
<span class="line"><span>        db.where: &quot;id&quot; { equals: $userId }</span></span>
<span class="line"><span>        db.first: { as: $authenticatedUser }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Store it in the request scope so downstream controllers can use it</span></span>
<span class="line"><span>        var: $currentUser { val: $authenticatedUser }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Continue to the intended route</span></span>
<span class="line"><span>        http.next: true</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><p>Then, apply this middleware to your protected routes:</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// src/main.zl</span></span>
<span class="line"><span>include: &#39;src/middleware/auth.zl&#39;</span></span>
<span class="line"><span></span></span>
<span class="line"><span>http.get: &#39;/dashboard&#39; {</span></span>
<span class="line"><span>    middleware: [&#39;auth&#39;]</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        // Because the &#39;auth&#39; middleware ran first, we know $currentUser exists here!</span></span>
<span class="line"><span>        http.view: &#39;dashboard&#39; {</span></span>
<span class="line"><span>            user: $currentUser</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="_5-logging-out" tabindex="-1">5. Logging Out <a class="header-anchor" href="#_5-logging-out" aria-label="Permalink to &quot;5. Logging Out&quot;">​</a></h2><p>To log a user out, you clear their session data using <code>session.destroy</code> or <code>session.delete</code>.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>http.post: &#39;/logout&#39; {</span></span>
<span class="line"><span>    middleware: [&#39;auth&#39;]</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        // Clear all session data</span></span>
<span class="line"><span>        session.destroy: true</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Redirect back to the homepage</span></span>
<span class="line"><span>        http.redirect: &#39;/&#39;</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><h2 id="summary" tabindex="-1">Summary <a class="header-anchor" href="#summary" aria-label="Permalink to &quot;Summary&quot;">​</a></h2><p>You now have a fully functioning, secure authentication system!</p><ul><li>You safely store passwords using <code>hash.make</code> (when registering) and <code>hash.verify</code> (when logging in).</li><li>You prevent Cross-Site Request Forgery (CSRF) on your forms using <code>{{ csrf_field() }}</code>.</li><li>You protect against Session Fixation using <code>session.regenerate</code>.</li><li>You secure private routes using <code>http.middleware</code>.</li></ul>`,15)])])}const u=e(t,[["render",h]]);export{g as __pageData,u as default};
