package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"zeno/internal/app"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"

	"github.com/go-chi/chi/v5"
)

func main() {
	eng := engine.NewEngine()
	r := chi.NewRouter()
	dbMgr := dbmanager.NewDBManager()

	// Register all slots to populate eng.Docs
	app.RegisterAllSlots(eng, r, dbMgr, nil, nil)

	var sb strings.Builder

	sb.WriteString("# Standard Library API Reference\n\n")
	sb.WriteString("This document is auto-generated from the ZenoEngine source code. It contains the complete reference for all built-in ZenoLang slots.\n\n")

	// Extract and sort keys
	var keys []string
	for k := range eng.Docs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Group keys by prefix (e.g. "db.", "http.")
	groups := make(map[string][]string)
	for _, k := range keys {
		parts := strings.Split(k, ".")
		group := "Core / General"
		if len(parts) > 1 {
			group = strings.ToUpper(parts[0][:1]) + parts[0][1:]
		}
		groups[group] = append(groups[group], k)
	}

	var groupNames []string
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)

	for _, g := range groupNames {
		sb.WriteString(fmt.Sprintf("## %s\n\n", g))
		
		for _, name := range groups[g] {
			meta := eng.Docs[name]
			
			sb.WriteString(fmt.Sprintf("### `%s`\n\n", name))
			
			if meta.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", meta.Description))
			} else {
				sb.WriteString("No description available.\n\n")
			}
			
			if len(meta.Inputs) > 0 {
				sb.WriteString("**Inputs:**\n\n")
				sb.WriteString("| Name | Type | Required | Description |\n")
				sb.WriteString("|------|------|----------|-------------|\n")
				
				var inputKeys []string
				for k := range meta.Inputs {
					inputKeys = append(inputKeys, k)
				}
				sort.Strings(inputKeys)
				
				for _, ik := range inputKeys {
					opt := meta.Inputs[ik]
					req := "No"
					if opt.Required {
						req = "**Yes**"
					}
					typ := opt.Type
					if typ == "" {
						typ = "any"
					}
					
					sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s |\n", ik, typ, req, opt.Description))
				}
				sb.WriteString("\n")
			}
			
			if len(meta.RequiredBlocks) > 0 {
				sb.WriteString(fmt.Sprintf("**Required Blocks:** `%s`\n\n", strings.Join(meta.RequiredBlocks, "`, `")))
			}
			
			if meta.ValueType != "" {
				sb.WriteString(fmt.Sprintf("**Main Value Type:** `%s`\n\n", meta.ValueType))
			}
			
			if meta.Example != "" {
				sb.WriteString("**Example:**\n```zeno\n")
				sb.WriteString(meta.Example)
				sb.WriteString("\n```\n\n")
			}
			
			sb.WriteString("---\n\n")
		}
	}

	err := os.MkdirAll("DOCS/docs/ecosystem", 0755)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		return
	}

	err = os.WriteFile("DOCS/docs/ecosystem/stdlib-reference.md", []byte(sb.String()), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Printf("Successfully generated %d slots documentation at DOCS/docs/ecosystem/stdlib-reference.md\n", len(keys))
}
