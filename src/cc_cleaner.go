// CrunchyCleaner: Core Cleaner Functions

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// task defines a cleanup action
// Either a Go-native function (goFunc) or an external command (cmd)
type task struct {
	desc   string       // Short description of the task
	cmd    []string     // Command-line slice to execute
	goFunc func() error // Go function to execute
}

// cleanFolder deletes **everything inside** a folder but keeps the folder itself alive
func cleanFolder(folder string) error {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := folder + string(os.PathSeparator) + entry.Name()
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

// cleanup orchestrates all cleanup tasks based on the mode (user/full/etc)
// and optionally a username.
func cleanup(mode string, username ...string) {
	var tasks []task

	switch mode {
	case "user":
		profile, ok := selectProfile(username)
		if !ok {
			return
		}
		tasks = buildUserTasks(goos, profile)
	default:
		tasks = buildSystemTasks(goos, mode)
	}

	askVerbose()
	previewTasks(tasks)
	runTasks(tasks)
}

// ==================== PROFILE HANDLING ==================== //

// selectProfile returns the chosen user profile (via CLI arg or interactive input)
func selectProfile(username []string) (string, bool) {
	profiles := getProfiles()
	if len(profiles) == 0 {
		printError("No user profiles found!")
		if !skipPause {
			pause()
		}
		return "", false
	}

	var selectedProfile string
	if len(username) > 0 {
		selectedProfile = username[0]
		if !contains(profiles, selectedProfile) {
			printError("Invalid profile name provided")
			if !skipPause {
				pause()
			}
			os.Exit(1)
		}
	} else {
		fmt.Println("Available profiles:")
		for _, p := range profiles {
			fmt.Printf("  %s\n", p)
		}
		fmt.Printf("Select profile name to clean%s", PROMPT)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		if !contains(profiles, choice) {
			printError("Invalid profile name provided")
			os.Exit(1)
		}
		selectedProfile = choice
	}

	fmt.Printf("%sSelected profile: %s%s\n", YELLOW, selectedProfile, RC)
	return selectedProfile, true
}

// getProfiles fetches user profiles depending on OS
func getProfiles() []string {
	var profiles []string
	var files []os.DirEntry
	var err error

	switch goos {
	case "windows":
		userBase := os.Getenv("SystemDrive") + `\\Users`
		files, err = os.ReadDir(userBase)
	default:
		files, err = os.ReadDir("/home")
	}

	if err != nil {
		printError("Reading profiles folder: ")
		fmt.Printf("%s\n", err)
		return nil
	}

	for _, f := range files {
		if f.IsDir() {
			profiles = append(profiles, f.Name())
		}
	}
	return profiles
}

// contains helper: check if slice contains string
func contains(list []string, val string) bool {
	for _, item := range list {
		if item == val {
			return true
		}
	}
	return false
}

// ==================== TASK BUILDERS ==================== //

// buildUserTasks builds cleanup tasks for a given user profile
func buildUserTasks(goos, profile string) []task {
	switch goos {
	case "windows":
		userPath := os.Getenv("SystemDrive") + `\\Users\\` + profile
		return []task{
			{desc: "~\\...\\Windows\\Explorer (built-in)", goFunc: func() error { return cleanFolder(userPath + `\\AppData\\Local\\Microsoft\\Windows\\Explorer`) }},
			{desc: "~\\...\\Local\\Temp (built-in)", goFunc: func() error { return cleanFolder(userPath + `\\AppData\\Local\\Temp`) }},
		}
	default:
		userPath := "/home/" + profile
		return []task{
			{desc: "~/.../share/Trash (built-in)", goFunc: func() error { return cleanFolder(userPath + "/.local/share/Trash") }},
			{desc: "~/.cache (built-in)", goFunc: func() error { return cleanFolder(userPath + "/.cache") }},
			{desc: "~/.thumbnails (built-in)", goFunc: func() error { return cleanFolder(userPath + "/.thumbnails") }},
		}
	}
}

// buildSystemTasks builds cleanup tasks depending on OS and mode
func buildSystemTasks(goos, mode string) []task {
	switch goos {
	case "windows":
		return buildWindowsTasks(mode)
	default:
		return buildLinuxTasks(mode)
	}
}

// buildWindowsTasks returns Windows cleanup tasks
func buildWindowsTasks(mode string) []task {
	switch mode {
	case "full":
		return []task{
			{desc: "%TEMP% (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("TEMP")) }},
			{desc: "\\..\\Windows\\Explorer (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("LocalAppData") + `\\Microsoft\\Windows\\Explorer`) }},
			{desc: "\\..\\FontCache (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("LocalAppData") + `\\FontCache`) }},
			{desc: "\\..\\winevt\\Logs (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\\System32\\winevt\\Logs`) }},
			{desc: "\\..\\WindowsUpdate\\Logs (built-in)", goFunc: func() error {
				return cleanFolder(os.Getenv("ProgramData") + `\\Microsoft\\Windows\\WindowsUpdate\\Logs`)
			}},
			{desc: "\\..\\Logs (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\\Logs`) }},
			{desc: "\\..\\Temp (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\\Temp`) }},
			{desc: "\\..\\Windows\\WER (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("ProgramData") + `\\Microsoft\\Windows\\WER`) }},
			{desc: "\\..\\..\\DeliveryOptimization (built-in)", goFunc: func() error {
				return cleanFolder(os.Getenv("SystemRoot") + `\\SoftwareDistribution\\DeliveryOptimization`)
			}},
			//{desc: "Admin Trash (shell)", cmd: []string{"powershell", "-Command", "Clear-RecycleBin -Force -Confirm:$false -ErrorAction SilentlyContinue"}},
			{desc: "Windows Update Cache (shell)", cmd: []string{"powershell", "Remove-Item -Path $env:SystemRoot\\SoftwareDistribution\\Download\\* -Recurse -Force"}},
			{desc: "\\..\\Prefetch (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\\Prefetch`) }},
			{desc: "DNS Cache (shell)", cmd: []string{"ipconfig", "/flushdns"}},
		}
	default:
		return []task{
			{desc: "%TEMP% (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("TEMP")) }},
			{desc: "\\..\\Windows\\Explorer (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("LocalAppData") + `\\Microsoft\\Windows\\Explorer`) }},
			{desc: "\\..\\FontCache (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("LocalAppData") + `\\FontCache`) }},
			{desc: "DNS Cache (shell)", cmd: []string{"ipconfig", "/flushdns"}},
		}
	}
}

// buildLinuxTasks returns Linux cleanup tasks
func buildLinuxTasks(mode string) []task {
	switch mode {
	case "full":
		return []task{
			{desc: "/tmp (built-in)", goFunc: func() error { return cleanFolder("/tmp") }},
			{desc: "/var/tmp (built-in)", goFunc: func() error { return cleanFolder("/var/tmp") }},
			{desc: "/var/cache (built-in)", goFunc: func() error { return cleanFolder("/var/cache") }},
			//{desc: "Root Trash (built-in)", goFunc: func() error { return cleanFolder(os.Getenv("HOME") + "/.local/share/Trash") }},
			{desc: "All System Logs (>10 days) (shell)", cmd: []string{"sh", "-c", "find /var/log -type f -mtime +10 -exec rm -f {} +"}},
			{desc: "/../systemd/coredump (built-in)", goFunc: func() error { return cleanFolder("/var/lib/systemd/coredump") }},
			{desc: "/var/crash (built-in)", goFunc: func() error { return cleanFolder("/var/crash") }},
			{desc: "fc-cache (shell)", cmd: []string{"fc-cache", "-fr"}},
			{desc: "Apt Cache (shell)", cmd: []string{"apt-get", "clean"}},
			{desc: "Flatpak Cache (shell)", cmd: []string{"flatpak", "uninstall", "--unused", "-y"}},
			{desc: "Pip Cache (shell)", cmd: []string{"pip", "cache", "purge"}},
			{desc: "Npm Cache (shell)", cmd: []string{"npm", "cache", "clean", "--force"}},
			{desc: "Yarn Cache (shell)", cmd: []string{"yarn", "cache", "clean"}},
			{desc: "DNF Cache (shell)", cmd: []string{"sudo", "dnf", "clean", "all"}},
			{desc: "Pacman Cache (shell)", cmd: []string{"sudo", "pacman", "-Scc", "--noconfirm"}},
			{desc: "Nix Garbage Collector (shell)", cmd: []string{"nix-collect-garbage", "-d"}},
			{desc: "Composer Cache (shell)", cmd: []string{"composer", "clear-cache"}},
			{desc: "Go Module Cache (shell)", cmd: []string{"go", "clean", "-modcache"}},
			{desc: "Rust Cargo Cache (shell)", cmd: []string{"cargo", "clean"}},
			{desc: "DNS Cache (shell)", cmd: []string{"systemd-resolve", "--flush-caches"}},
		}
	default:
		return []task{
			{desc: "/tmp (built-in)", goFunc: func() error { return cleanFolder("/tmp") }},
			{desc: "Journal Logs (100 days) (shell)", cmd: []string{"journalctl", "--vacuum-time=100d"}},
			{desc: "fc-cache (shell)", cmd: []string{"fc-cache", "-fr"}},
			{desc: "Apt Cache (shell)", cmd: []string{"apt-get", "clean"}},
			{desc: "Flatpak Cache (shell)", cmd: []string{"flatpak", "uninstall", "--unused", "-y"}},
			{desc: "Pacman Cache (shell)", cmd: []string{"pacman", "-Scc", "--noconfirm"}},
			{desc: "DNF Cache (shell)", cmd: []string{"sudo", "dnf", "clean", "all"}},
			{desc: "DNS Cache (shell)", cmd: []string{"systemd-resolve", "--flush-caches"}},
		}
	}
}

// ==================== EXECUTION ==================== //

// askVerbose asks the user if verbose logging should be enabled
func askVerbose() {
	if !skipPause {
		fmt.Print("Enable verbose logging? (yes/NO): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))
		if choice == "y" || choice == "yes" {
			verbose = true
		}
	}
}

// previewTasks prints all tasks before execution
func previewTasks(tasks []task) {
	printInfo("The following cleanup tasks will be executed:")
	for _, t := range tasks {
		cmdStr := strings.Join(t.cmd, " ")
		if t.goFunc != nil && cmdStr == "" {
			cmdStr = "(built-in)"
		}
		fmt.Printf("%s- Cleaning: %s%s\n  → %s\n", CYAN, t.desc, RC, cmdStr)
	}
	fmt.Printf("\n")
	printInfo("The above cleanup tasks will be executed")
	if verbose {
		printInfo("Verbose mode: showing all task details and errors")
	} else {
		printInfo("Verbose mode OFF: errors will be hidden")
	}
	printInfo("Press [CTRL+C] to cancel")
	printInfo("!!! You use this tool at your own risk !!!")
	if !skipPause {
		pause()
	}
}

// runTasks executes all cleanup tasks with spinner and logging
func runTasks(tasks []task) {
	fmt.Printf("%s#############################################%s\n", RED, RC)
	printInfo("*** CrunchyCleaner Cleanup STARTED ***\n")
	time.Sleep(2 * time.Second)

	startFree := getFreeMB()
	for _, t := range tasks {
		ctx, cancel := context.WithCancel(context.Background())
		go asyncSpinner(ctx, "Cleaning: "+t.desc)
		time.Sleep(CMDWAIT)

		var err error
		if t.goFunc != nil {
			_ = t.goFunc()
		} else if len(t.cmd) > 0 {
			_, err = runCommand(t.cmd)
		}
		cancel()
		fmt.Printf("\r\033[2K") // clear spinner line

		if verbose && err != nil {
			fmt.Printf("%sCleaning: %s%s FINISHED\n  → %s\n", CYAN, t.desc, RC, err)
		} else {
			fmt.Printf("%sCleaning: %s%s FINISHED\n", CYAN, t.desc, RC)
		}
	}

	endFree := getFreeMB()
	diff := endFree - startFree
	if diff < 0 {
		diff = 0
	}
	printInfo(fmt.Sprintf("Cleaned approx: %s%d MB%s disk space", YELLOW, diff, RC))
	printSuccess("*** CrunchyCleaner Cleanup FINISHED ***")
}
