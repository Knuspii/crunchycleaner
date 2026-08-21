// ##################################################################
// CrunchyCleaner
// Here are all tech functions.
// ##################################################################

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
)

// getDiskMetrics uses gopsutil to fetch precise, platform-independent disk data.
func getDiskMetrics() (freeGB float64, totalStr, freeStr string) {
	const GB = 1024 * 1024 * 1024

	// On Linux, "/" is the root. On Windows, gopsutil handles the mapping to the system drive.
	usage, err := disk.Usage("/")
	if err != nil {
		// Fallback for specific Windows environments if "/" fails
		usage, err = disk.Usage("C:")
		if err != nil {
			return 0, "N/A", "N/A"
		}
	}

	freeGB = float64(usage.Free) / GB
	totalStr = fmt.Sprintf("%.2f GB", float64(usage.Total)/GB)
	freeStr = fmt.Sprintf("%.2f GB", freeGB)
	return
}

// formatMB converts bytes to a string representing Megabytes
func formatMB(bytes int64) string {
	mb := float64(bytes) / 1024 / 1024
	return fmt.Sprintf("%.2f MB", mb)
}

// getDirSize utilizes WalkDir (introduced in Go 1.16), which is significantly
// faster than the older filepath.Walk because it avoids unnecessary Lstat calls.
func getDirSize(path string) int64 {
	var size int64
	// Resolve glob patterns (e.g., paths containing '*')
	matches, _ := filepath.Glob(expandHome(path))

	for _, m := range matches {
		// WalkDir is the high-performance standard for scanning directories
		_ = filepath.WalkDir(m, func(_ string, d os.DirEntry, err error) error {
			if err != nil {
				// If a folder is restricted (e.g. System Cache), we just skip it
				return nil
			}
			if !d.IsDir() {
				info, err := d.Info()
				if err == nil {
					size += info.Size()
				}
			}
			return nil
		})
	}
	return size
}

// expandHome resolves the shorthand '~/ ' to the absolute user home directory
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// function to scan which programs actually exist on the disk
func scanForExisting() []Program {
	allPrograms := getPrograms()
	existing := []Program{}
	for _, p := range allPrograms {
		found := false
		var totalSize int64
		for _, path := range p.Paths {
			matches, _ := filepath.Glob(expandHome(path))
			if len(matches) > 0 {
				found = true
				totalSize += getDirSize(path)
			}
		}
		if found {
			// We format the name here so it's ready for both UI and Logs
			p.Name = fmt.Sprintf("%-30s %s(%s)%s", p.Name, YELLOW, formatMB(totalSize), RC)
			existing = append(existing, p)
		}
	}
	return existing
}

func deletePath(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	if !info.IsDir() {
		_ = os.Remove(path)
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Print(CLEARLINE)
		logWarn("Cannot read " + path + ": " + err.Error())
		return
	}

	for _, e := range entries {
		full := filepath.Join(path, e.Name())
		if err := os.RemoveAll(full); err != nil {
			fmt.Print(CLEARLINE)
			logWarn("Skipped " + e.Name() + ": " + err.Error())
		}
	}
}
