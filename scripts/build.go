package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Config
var platforms = []struct {
	OS   string
	Arch string
}{
	{"windows", "amd64"},
	{"linux", "amd64"},
	{"darwin", "arm64"}, // macOS M1/M2
	{"darwin", "amd64"}, // macOS Intel
}

func main() {
	start := time.Now()
	fmt.Println("🚀 Starting build process...")

	// 1. Get Version Info
	version := getGitTag()
	commit := getGitCommit()
	date := time.Now().Format("2006-01-02")

	fmt.Printf("📦 Version: %s\n", version)
	fmt.Printf("🔧 Commit:  %s\n", commit)
	fmt.Printf("📅 Date:    %s\n", date)

	// 2. Prepare LDFLAGS
	// -s -w: Strip debug symbols to reduce size
	ldflags := fmt.Sprintf("-s -w -X marsx/cmd.Version=%s -X marsx/cmd.Commit=%s -X marsx/cmd.Date=%s", version, commit, date)

	// 3. Clean/Create dist directory
	distDir := "dist"
	if err := os.RemoveAll(distDir); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
		panic(err)
	}

	// 4. Build for each platform
	for _, p := range platforms {
		targetName := fmt.Sprintf("marsx-%s-%s", p.OS, p.Arch)
		if p.OS == "windows" {
			targetName += ".exe"
		}
		outputPath := filepath.Join(distDir, targetName)

		fmt.Printf("🔨 Building for %s/%s...\n", p.OS, p.Arch)

		cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", outputPath, "main.go")
		cmd.Env = append(os.Environ(),
			"GOOS="+p.OS,
			"GOARCH="+p.Arch,
			"CGO_ENABLED=0", // Static build
		)

		// Stream output
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Build failed for %s/%s: %v\n", p.OS, p.Arch, err)
			os.Exit(1)
		}
	}

	fmt.Printf("\n✅ Build completed in %v\n", time.Since(start))
	fmt.Println("📂 Artifacts in ./dist:")
	listDist(distDir)
}

func getGitTag() string {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(out))
}

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func listDist(dir string) {
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		info, _ := f.Info()
		fmt.Printf("  - %-25s (%.2f MB)\n", f.Name(), float64(info.Size())/1024/1024)
	}
}

