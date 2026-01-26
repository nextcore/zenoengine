package console

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"zeno/pkg/engine"

	"github.com/go-chi/chi/v5"
)

// Struktur Data Explorer
type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Type     string      `json:"type"`
	Ext      string      `json:"ext"`
	Children []*FileNode `json:"children"`
}

// Struktur Data API Doc
type ApiDocItem struct {
	FilePath string `json:"file_path"`
	Route    string `json:"route"`
	Method   string `json:"method"`
	Summary  string `json:"summary"`
	Params   string `json:"params"`
}

const htmlUI = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>zeno Console</title>
    <style>
        :root { --bg: #1e1e1e; --sidebar: #252526; --accent: #007acc; --text: #cccccc; --hover: #2a2d2e; --border: #333; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 0; display: flex; height: 100vh; background: var(--bg); color: var(--text); overflow: hidden; }

        #sidebar { width: 340px; background: var(--sidebar); border-right: 1px solid var(--border); display: flex; flex-direction: column; }

        /* TABS (3 Items) */
        .sidebar-tabs { display: flex; border-bottom: 1px solid var(--border); background: #1a1a1a; }
        .tab { flex: 1; text-align: center; padding: 10px 0; cursor: pointer; font-size: 0.75rem; font-weight: bold; color: #666; border-bottom: 3px solid transparent; transition: all 0.2s; }
        .tab.active { background: var(--sidebar); color: #fff; border-bottom-color: var(--accent); }
        .tab:hover { color: #fff; background: #333; }

        .tab-content { flex: 1; overflow-y: auto; display: none; }
        .tab-content.active { display: block; }

        /* File Tree */
        #file-tree { padding-bottom: 20px; }
        .tree-node { cursor: pointer; display: flex; align-items: center; padding: 4px 10px; font-size: 0.9rem; white-space: nowrap; }
        .tree-node:hover { background: var(--hover); color: #fff; }
        .children-container { display: none; margin-left: 15px; border-left: 1px solid #444; }
        .children-container.open { display: block; }
        .node-icon { margin-right: 6px; }

        /* API Docs Styles */
        .api-item { background: #222; margin: 8px; border-radius: 4px; border: 1px solid #444; overflow: hidden; }
        .api-header { padding: 8px; display: flex; align-items: center; gap: 8px; cursor: pointer; background: #2a2a2a; }
        .api-header:hover { background: #333; }
        .method { padding: 2px 6px; border-radius: 3px; font-weight: bold; font-size: 0.7rem; color: #fff; min-width: 40px; text-align: center; }
        .method.GET { background: #61affe; }
        .method.POST { background: #49cc90; }
        .method.PUT { background: #fca130; }
        .method.DELETE { background: #f93e3e; }
        .route { font-family: monospace; font-weight: bold; color: #eee; font-size: 0.85rem; }
        .summary { color: #888; font-size: 0.75rem; margin-left: auto; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 80px; }
        .api-details { padding: 10px; background: #1e1e1e; border-top: 1px solid #444; display: none; font-size: 0.85rem; }
        .api-details.open { display: block; }
        .detail-row { margin-bottom: 6px; }
        .detail-label { color: #666; font-weight: bold; font-size: 0.7rem; text-transform: uppercase; width: 50px; display: inline-block; }
        .file-link { color: #007acc; cursor: pointer; text-decoration: underline; }

        /* Slots Docs Styles */
        #slots-search { width: 90%; margin: 10px auto; padding: 6px; background: #111; border: 1px solid #444; color: #fff; border-radius: 4px; display: block; font-size: 0.85rem;}
        .slot-item { padding: 10px 15px; border-bottom: 1px solid #333; }
        .slot-name { color: #dcdcaa; font-weight: bold; font-family: monospace; font-size: 0.9rem; margin-bottom: 4px; }
        .slot-desc { font-size: 0.8rem; color: #aaa; margin-bottom: 8px; line-height: 1.4; }
        .slot-code { background: #111; padding: 8px; border-radius: 4px; font-family: monospace; font-size: 0.8rem; color: #ce9178; white-space: pre-wrap; cursor: pointer; position: relative; }
        .slot-code:hover { background: #000; }
        .slot-code::after { content: "Click to Insert"; position: absolute; top: 2px; right: 5px; font-size: 9px; opacity: 0; color: #666; }
        .slot-code:hover::after { opacity: 1; }

        /* Main Area */
        #main { flex: 1; display: flex; flex-direction: column; min-width: 0; }
        #toolbar { padding: 8px 15px; background: #2d2d2d; display: flex; align-items: center; gap: 10px; border-bottom: 1px solid var(--border); }
        #current-file { font-size: 0.9rem; color: #fff; margin-right: auto; }
        .btn { background: #3c3c3c; color: #eee; border: 1px solid #444; padding: 4px 12px; cursor: pointer; font-size: 0.8rem; border-radius: 3px; }
        .btn:hover { background: #505050; }
        .btn-primary { background: var(--accent); border-color: var(--accent); color: white; }
        .btn-green { background: #2e7d32; border-color: #2e7d32; color: white; }

        #editor { flex: 1; }
        #output-panel { height: 25%; background: #111; border-top: 1px solid var(--border); display: flex; flex-direction: column; }
        .panel-header { padding: 5px 15px; background: #1e1e1e; font-size: 0.75rem; font-weight: bold; color: #888; text-transform: uppercase; display: flex; justify-content: space-between; border-bottom: 1px solid #333; }
        #output { flex: 1; padding: 10px; font-family: 'Consolas', 'Courier New', monospace; font-size: 0.85rem; white-space: pre-wrap; overflow-y: auto; color: #e0e0e0; }
        .log-success { color: #4caf50; }
        .log-error { color: #f44336; }
    </style>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/ace/1.4.12/ace.js"></script>
</head>
<body>
    <div id="sidebar">
        <div class="sidebar-tabs">
            <div class="tab active" onclick="switchTab('files')">FILES</div>
            <div class="tab" onclick="switchTab('api')">API</div>
            <div class="tab" onclick="switchTab('slots')">SLOTS</div>
        </div>

        <div id="tab-files" class="tab-content active">
            <div style="padding:10px; display:flex; justify-content:flex-end;">
                 <span style="cursor:pointer; font-size:11px; color:#666" onclick="loadFiles()" title="Refresh">↻ Refresh</span>
            </div>
            <div id="file-tree">Loading...</div>
        </div>

        <div id="tab-api" class="tab-content">
             <div style="padding:10px; display:flex; justify-content:space-between; align-items:center;">
                 <span style="font-size:11px; font-weight:bold; color:#888">ENDPOINTS</span>
                 <span style="cursor:pointer; font-size:11px; color:#666" onclick="loadApiDocs()" title="Refresh">↻</span>
            </div>
            <div id="api-list">Scanning...</div>
        </div>

        <div id="tab-slots" class="tab-content">
            <input type="text" id="slots-search" placeholder="Search function..." onkeyup="filterSlots()">
            <div id="slots-list">Loading...</div>
        </div>
    </div>

    <div id="main">
        <div id="toolbar">
            <span id="current-file">Scratchpad</span>
            <button class="btn" onclick="newDraft()">+ New</button>
            <button class="btn btn-primary" onclick="saveFile()">💾 Save</button>
            <button class="btn btn-green" onclick="runScript()">▶ Run</button>
        </div>
        <div id="editor"></div>
        <div id="output-panel">
            <div class="panel-header">
                <span>Console Output</span>
                <span style="cursor:pointer" onclick="document.getElementById('output').innerText=''">Clear</span>
            </div>
            <div id="output">// Ready.</div>
        </div>
    </div>

    <script>
        var editor = ace.edit("editor");
        editor.setTheme("ace/theme/monokai");
        editor.setFontSize(14);
        editor.setShowPrintMargin(false);

        let currentPath = "";
        let allSlots = {};

        // --- TABS LOGIC ---
        function switchTab(name) {
            // Reset active class
            document.querySelectorAll('.tab').forEach(el => el.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(el => el.classList.remove('active'));

            // Activate selected
            if(name === 'files') {
                document.querySelectorAll('.tab')[0].classList.add('active');
                document.getElementById('tab-files').classList.add('active');
            } else if (name === 'api') {
                document.querySelectorAll('.tab')[1].classList.add('active');
                document.getElementById('tab-api').classList.add('active');
                loadApiDocs();
            } else if (name === 'slots') {
                document.querySelectorAll('.tab')[2].classList.add('active');
                document.getElementById('tab-slots').classList.add('active');
                if(!Object.keys(allSlots).length) loadSlotDocs();
            }
        }

        // --- 1. FILES LOGIC ---
        async function loadFiles() {
            const res = await fetch('/console/api/files');
            const data = await res.json();
            const container = document.getElementById('file-tree');
            container.innerHTML = "";
            data.forEach(n => container.appendChild(createNode(n)));
        }

        function createNode(node) {
            const wrap = document.createElement('div');
            const row = document.createElement('div');
            row.className = 'tree-node';
            row.innerHTML = '<span class="node-icon">'+(node.type==='folder'?'📁':(node.ext==='.html'?'🌐':'📄'))+'</span><span>'+node.name+'</span>';
            wrap.appendChild(row);

            if(node.type==='folder') {
                const childDiv = document.createElement('div');
                childDiv.className = 'children-container';
                if(node.name==='src' || node.name==='views') childDiv.classList.add('open');
                if(node.children) node.children.forEach(c => childDiv.appendChild(createNode(c)));
                wrap.appendChild(childDiv);
                row.onclick = () => childDiv.classList.toggle('open');
            } else {
                row.onclick = () => openFile(node.path);
            }
            return wrap;
        }

        // --- 2. API DOCS LOGIC ---
        async function loadApiDocs() {
            const container = document.getElementById('api-list');
            try {
                const res = await fetch('/console/api/endpoints');
                const endpoints = await res.json();

                if(!endpoints || endpoints.length === 0) {
                    container.innerHTML = '<div style="padding:20px; text-align:center; color:#666">No endpoints found.</div>';
                    return;
                }

                let html = "";
                endpoints.forEach(api => {
                    html += '<div class="api-item">';
                    html += '<div class="api-header" onclick="this.nextElementSibling.classList.toggle(\'open\')">';
                    html += '<span class="method '+api.method+'">'+api.method+'</span>';
                    html += '<span class="route">'+api.route+'</span>';
                    html += '<span class="summary">'+api.summary+'</span>';
                    html += '</div>';
                    html += '<div class="api-details">';
                    html += '<div class="detail-row"><span class="detail-label">FILE</span> <span class="file-link" onclick="openFile(\''+api.file_path+'\')">'+api.file_path+'</span></div>';
                    html += '<div class="detail-row"><span class="detail-label">DESC</span> '+api.summary+'</div>';
                    if(api.params) html += '<div class="detail-row"><span class="detail-label">PARAMS</span> '+api.params+'</div>';
                    html += '</div></div>';
                });
                container.innerHTML = html;
            } catch(e) { container.innerHTML = "Error loading API docs"; }
        }

        // --- 3. SLOTS DOCS LOGIC ---
        async function loadSlotDocs() {
            try {
                const res = await fetch('/console/api/docs');
                allSlots = await res.json();
                renderSlots(allSlots);
            } catch (e) { console.error(e); }
        }

        function renderSlots(docs) {
            const container = document.getElementById('slots-list');
            let html = "";
            // Sort keys
            const keys = Object.keys(docs).sort();
            keys.forEach(name => {
                const meta = docs[name];
                html += '<div class="slot-item" data-name="'+name+'">';
                html += '<div class="slot-name">' + name + '</div>';
                html += '<div class="slot-desc">' + meta.description + '</div>';
                if(meta.example) {
                    const safeCode = meta.example.replace(/</g, "&lt;").replace(/>/g, "&gt;");
                    html += '<div class="slot-code" onclick="insertCode(this)" title="Insert Code">' + safeCode + '</div>';
                }
                html += '</div>';
            });
            container.innerHTML = html;
        }

        function filterSlots() {
            const term = document.getElementById('slots-search').value.toLowerCase();
            document.querySelectorAll('.slot-item').forEach(item => {
                const name = item.getAttribute('data-name');
                item.style.display = name.toLowerCase().includes(term) ? 'block' : 'none';
            });
        }

        function insertCode(el) {
            editor.insert(el.innerText + "\n");
            editor.focus();
        }

        // --- EDITOR & RUN ---
        async function openFile(path) {
            currentPath = path;
            document.getElementById('current-file').innerText = path;
            editor.session.setMode(path.endsWith('.html') ? "ace/mode/html" : "ace/mode/yaml");
            const res = await fetch('/console/api/read?path='+encodeURIComponent(path));
            editor.setValue(await res.text(), -1);
        }

        function newDraft() {
            currentPath = "";
            document.getElementById('current-file').innerText = "Scratchpad";
            editor.setValue("", -1);
        }

        async function saveFile() {
            if(!currentPath) {
                const name = prompt("Path file baru:");
                if(!name) return;
                currentPath = name;
            }
            await fetch('/console/api/save', {
                method: 'POST',
                body: JSON.stringify({path: currentPath, content: editor.getValue()})
            });
            alert("Saved!");
            loadFiles();
        }

        async function runScript() {
            if(currentPath.endsWith('.html')) return alert("Cannot run HTML");
            const res = await fetch('/console/api/run', { method:'POST', body: editor.getValue() });
            const text = await res.text();
            const outDiv = document.getElementById('output');
            outDiv.innerHTML = text.includes("Error") || text.includes("❌") ?
                '<span class="log-error">'+text+'</span>' : '<span class="log-success">'+text+'</span>';
        }

        loadFiles();
        newDraft();
    </script>
</body>
</html>
`

// --- HELPERS (Scan Directory & Script) ---

func scanDirectory(rootPath string) (*FileNode, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}
	node := &FileNode{Name: info.Name(), Path: filepath.ToSlash(rootPath), Type: "file", Ext: filepath.Ext(info.Name())}
	if info.IsDir() {
		node.Type = "folder"
		entries, err := os.ReadDir(rootPath)
		if err == nil {
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				childPath := filepath.Join(rootPath, entry.Name())
				if childNode, err := scanDirectory(childPath); err == nil {
					if childNode.Type == "folder" || childNode.Ext == ".zl" || childNode.Ext == ".html" {
						node.Children = append(node.Children, childNode)
					}
				}
			}
			sort.Slice(node.Children, func(i, j int) bool {
				if node.Children[i].Type != node.Children[j].Type {
					return node.Children[i].Type == "folder"
				}
				return node.Children[i].Name < node.Children[j].Name
			})
		}
	}
	return node, nil
}

func scanScriptForDocs(root *engine.Node) *ApiDocItem {
	if root.Name == "doc.api" {
		doc := &ApiDocItem{Method: "GET"}
		for _, c := range root.Children {
			val := fmt.Sprintf("%v", c.Value)
			switch c.Name {
			case "route":
				doc.Route = val
			case "method":
				doc.Method = strings.ToUpper(val)
			case "summary", "desc":
				doc.Summary = val
			case "params":
				doc.Params = val
			}
		}
		if doc.Route != "" {
			return doc
		}
	}
	for _, child := range root.Children {
		if found := scanScriptForDocs(child); found != nil {
			return found
		}
	}
	return nil
}

// --- REGISTER ROUTES ---

func RegisterRoutes(r chi.Router, eng *engine.Engine) {
	r.Route("/console", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(htmlUI))
		})

		// 1. FILES API
		r.Get("/api/files", func(w http.ResponseWriter, r *http.Request) {
			var roots []*FileNode
			if n, err := scanDirectory("src"); err == nil {
				roots = append(roots, n)
			}
			if n, err := scanDirectory("views"); err == nil {
				roots = append(roots, n)
			}
			json.NewEncoder(w).Encode(roots)
		})

		// 2. ENDPOINTS API (SCANNER)
		r.Get("/api/endpoints", func(w http.ResponseWriter, r *http.Request) {
			var docs []ApiDocItem
			filepath.Walk("src", func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && strings.HasSuffix(path, ".zl") {
					root, err := engine.LoadScript(path)
					if err == nil {
						if doc := scanScriptForDocs(root); doc != nil {
							doc.FilePath = filepath.ToSlash(path)
							docs = append(docs, *doc)
						}
					}
				}
				return nil
			})
			json.NewEncoder(w).Encode(docs)
		})

		// 3. SLOTS DOCS API (METADATA)
		r.Get("/api/docs", func(w http.ResponseWriter, r *http.Request) {
			docs := eng.GetDocumentation()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(docs)
		})

		// ... Read, Save, Run ...
		r.Get("/api/read", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Query().Get("path")
			if strings.Contains(path, "..") {
				http.Error(w, "Forbidden", 403)
				return
			}
			content, _ := os.ReadFile(path)
			w.Write(content)
		})

		r.Post("/api/save", func(w http.ResponseWriter, r *http.Request) {
			var payload struct{ Path, Content string }
			json.NewDecoder(r.Body).Decode(&payload)
			if strings.Contains(payload.Path, "..") {
				http.Error(w, "Forbidden", 403)
				return
			}
			os.MkdirAll(filepath.Dir(payload.Path), 0755)
			os.WriteFile(payload.Path, []byte(payload.Content), 0644)
			w.Write([]byte("OK"))
		})

		r.Post("/api/run", func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			tmpFile := "src/temp_scratchpad.zl"
			os.WriteFile(tmpFile, body, 0644)
			defer os.Remove(tmpFile)
			root, err := engine.LoadScript(tmpFile)
			if err != nil {
				w.Write([]byte("Error: " + err.Error()))
				return
			}
			eng.Execute(context.WithValue(r.Context(), "httpWriter", w), root, engine.NewScope(nil))
		})
	})
}
