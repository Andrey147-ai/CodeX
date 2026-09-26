package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ========== PACKAGE REGISTRY (codex list/search/info) ==========
// Registry source order: $CODEX_REGISTRY (file path or http URL),
// then ./packages.json, then the canonical file on GitHub main.

type registryPkg struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type registry struct {
	Packages []registryPkg `json:"packages"`
}

func loadRegistry() (registry, string, error) {
	var reg registry
	if src := os.Getenv("CODEX_REGISTRY"); src != "" {
		var data []byte
		var err error
		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
			client := &http.Client{Timeout: 15 * time.Second}
			var resp *http.Response
			resp, err = client.Get(src)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					err = fmt.Errorf("registry: %s", resp.Status)
				} else {
					data, err = io.ReadAll(io.LimitReader(resp.Body, 1<<20))
				}
			}
		} else {
			data, err = os.ReadFile(src)
		}
		if err != nil {
			return reg, src, err
		}
		if err := json.Unmarshal(data, &reg); err != nil {
			return reg, src, err
		}
		return reg, src, nil
	}
	if data, err := os.ReadFile("packages.json"); err == nil {
		if err := json.Unmarshal(data, &reg); err != nil {
			return reg, "packages.json", err
		}
		return reg, "packages.json", nil
	}
	const remote = "https://raw.githubusercontent.com/Andrey147-ai/CodeX/main/packages.json"
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(remote)
	if err != nil {
		return reg, remote, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return reg, remote, fmt.Errorf("registry: %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return reg, remote, err
	}
	if err := json.Unmarshal(data, &reg); err != nil {
		return reg, remote, err
	}
	return reg, remote, nil
}

func printPkgRow(p registryPkg) {
	tags := ""
	if len(p.Tags) > 0 {
		tags = " [" + strings.Join(p.Tags, ", ") + "]"
	}
	fmt.Printf("%s\n  %s%s\n", p.Name, p.Description, tags)
}

func codexList() int {
	reg, src, err := loadRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list: cannot load registry (%s): %v\n", src, err)
		return 1
	}
	if len(reg.Packages) == 0 {
		fmt.Println("No packages in registry.")
		return 0
	}
	for _, p := range reg.Packages {
		printPkgRow(p)
	}
	fmt.Printf("(%d packages from %s)\n", len(reg.Packages), src)
	return 0
}

func codexSearch(query string) int {
	reg, src, err := loadRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "search: cannot load registry (%s): %v\n", src, err)
		return 1
	}
	q := strings.ToLower(query)
	hits := 0
	for _, p := range reg.Packages {
		hay := strings.ToLower(p.Name + " " + p.Description + " " + strings.Join(p.Tags, " "))
		if strings.Contains(hay, q) {
			printPkgRow(p)
			hits++
		}
	}
	if hits == 0 {
		fmt.Printf("No packages match %q (%s).\n", query, src)
		return 1
	}
	fmt.Printf("(%d/%d from %s)\n", hits, len(reg.Packages), src)
	return 0
}

func codexInfo(spec string) int {
	reg, src, err := loadRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "info: cannot load registry (%s): %v\n", src, err)
		return 1
	}
	for _, p := range reg.Packages {
		if p.Name == spec || strings.HasSuffix(p.Name, "/"+spec) {
			printPkgRow(p)
			fmt.Printf("  import \"%s\"\n", p.Name)
			fmt.Printf("  codex get %s\n", p.Name)
			return 0
		}
	}
	fmt.Fprintf(os.Stderr, "info: no package %q in registry (%s)\n", spec, src)
	return 1
}
