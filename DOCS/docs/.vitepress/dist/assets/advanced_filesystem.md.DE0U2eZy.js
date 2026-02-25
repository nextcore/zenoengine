import{_ as s,o as n,c as e,ag as p}from"./chunks/framework.ePeAWSvT.js";const h=JSON.parse('{"title":"Filesystem & Uploads","description":"","frontmatter":{},"headers":[],"relativePath":"advanced/filesystem.md","filePath":"advanced/filesystem.md"}'),l={name:"advanced/filesystem.md"};function t(i,a,o,d,c,r){return n(),e("div",null,[...a[0]||(a[0]=[p(`<h1 id="filesystem-uploads" tabindex="-1">Filesystem &amp; Uploads <a class="header-anchor" href="#filesystem-uploads" aria-label="Permalink to &quot;Filesystem &amp; Uploads&quot;">​</a></h1><p>Handling HTTP request uploads, storing files, and manipulating the filesystem in ZenoEngine revolves around the <code>http.upload</code> and built-in filesystem (<code>fs</code>) slots.</p><h2 id="handling-file-uploads" tabindex="-1">Handling File Uploads <a class="header-anchor" href="#handling-file-uploads" aria-label="Permalink to &quot;Handling File Uploads&quot;">​</a></h2><p>When receiving multipart/form-data POST requests containing files, you can use the <code>http.upload</code> slot. This slot simplifies the safe extraction and storage of uploaded media.</p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>http.post: &#39;/api/profile/upload&#39; {</span></span>
<span class="line"><span>    do: {</span></span>
<span class="line"><span>        // Retrieve the uploaded file using the input name &quot;avatar&quot;</span></span>
<span class="line"><span>        // And save it gracefully into the specified destination folder.</span></span>
<span class="line"><span>        http.upload: &quot;avatar&quot; {</span></span>
<span class="line"><span>            dest: &quot;./storage/avatars&quot;</span></span>
<span class="line"><span>            as: $uploadedFile</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // $uploadedFile will contain metadata like original_name, size, and its saved path</span></span>
<span class="line"><span>        if: $uploadedFile == null {</span></span>
<span class="line"><span>            then: {</span></span>
<span class="line"><span>                http.bad_request: { error: &quot;No file uploaded or invalid format!&quot; }</span></span>
<span class="line"><span>                return</span></span>
<span class="line"><span>            }</span></span>
<span class="line"><span>        }</span></span>
<span class="line"><span>        </span></span>
<span class="line"><span>        // Save the file path to the database</span></span>
<span class="line"><span>        db.table: &quot;users&quot;</span></span>
<span class="line"><span>        db.where: &quot;id&quot; { equals: 1 }</span></span>
<span class="line"><span>        db.update: { avatar_path: $uploadedFile.path }</span></span>
<span class="line"><span></span></span>
<span class="line"><span>        http.ok: { message: &quot;Avatar uploaded successfully!&quot;, file: $uploadedFile }</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div><p>The <code>http.upload</code> slot handles common potential errors securely. It restricts paths gracefully without risking directory traversal attacks while uploading.</p><h2 id="basic-filesystem-operations" tabindex="-1">Basic Filesystem Operations <a class="header-anchor" href="#basic-filesystem-operations" aria-label="Permalink to &quot;Basic Filesystem Operations&quot;">​</a></h2><p>ZenoEngine provides simple OS-level integrations via Native filesystem operations like <code>fs.read</code>, <code>fs.write</code>, and checking if files exist (<code>fs.exists</code>).</p><p><em>Note: Since ZenoEngine targets web apps primarily, file operations are usually abstracted behind DB operations or cloud storage, and large file rendering is handled automatically by Blade/Views, or <code>http.static</code> instead.</em></p><div class="language-zeno vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">zeno</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>// Reading the contents of a text file</span></span>
<span class="line"><span>fs.read: &quot;storage/data/report.csv&quot; {</span></span>
<span class="line"><span>    as: $csvData</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>// Checking if an important configuration exists</span></span>
<span class="line"><span>fs.exists: &quot;.env&quot; {</span></span>
<span class="line"><span>    as: $hasConfig</span></span>
<span class="line"><span>}</span></span>
<span class="line"><span></span></span>
<span class="line"><span>if: $hasConfig == false {</span></span>
<span class="line"><span>    then: {</span></span>
<span class="line"><span>        log: &quot;WARNING: No .env configuration found!&quot;</span></span>
<span class="line"><span>    }</span></span>
<span class="line"><span>}</span></span></code></pre></div>`,10)])])}const f=s(l,[["render",t]]);export{h as __pageData,f as default};
