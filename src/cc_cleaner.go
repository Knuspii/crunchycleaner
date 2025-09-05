// CrunchyCleaner: Core Cleaner Functions

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

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

// cleanup orchestrates all cleanup tasks based on the mode (user/full/etc) and optionally a username.
func cleanup(mode string, username ...string) {
	// Define a cleanup task with a description and the command to execute
	type task struct {
		desc   string       // Short description of the task
		cmd    []string     // Command-line slice to execute if it's a cmd task
		goFunc func() error // Go function to execute if it's a Go-native task
	}

	var tasks []task
	switch mode {
	// ==================== USER PROFILE CLEANUP ==================== //
	case "user":
		var profiles []string
		switch goos {
		case "windows":
			userBase := os.Getenv("SystemDrive") + `\Users`
			files, err := os.ReadDir(userBase)
			if err != nil {
				printError("Reading Users folder: ")
				fmt.Printf("%s\n", err)
				return
			}
			for _, f := range files {
				if f.IsDir() {
					profiles = append(profiles, f.Name())
				}
			}

		default: // Linux / Unix-like
			files, err := os.ReadDir("/home")
			if err != nil {
				printError("Reading /home folder: ")
				fmt.Printf("%s\n", err)
				return
			}
			for _, f := range files {
				if f.IsDir() {
					profiles = append(profiles, f.Name())
				}
			}
		}

		// Bail out if we found no profiles
		if len(profiles) == 0 {
			printError("No user profiles found!")
			return
		}

		// Handle CLI argument for username or interactive selection
		if len(username) > 0 {
			selectedProfile = username[0]

			// Check if user exists
			valid := false
			for _, p := range profiles {
				if p == selectedProfile {
					valid = true
					break
				}
			}
			if !valid {
				printError("Invalid profile name provided")
				os.Exit(1)
			}

		} else {
			// Interactive mode
			fmt.Println("Available profiles:")
			for _, p := range profiles {
				fmt.Printf("  %s\n", p)
			}
			fmt.Printf("Select profile name to clean%s", PROMPT)
			choice, _ := reader.ReadString('\n')
			choice = strings.TrimSpace(choice)

			valid := false
			for _, p := range profiles {
				if p == choice {
					valid = true
					break
				}
			}
			if !valid {
				printError("Invalid profile name provided")
				os.Exit(1)
			}
			selectedProfile = choice
		}

		fmt.Printf("%sSelected profile: %s%s\n", YELLOW, selectedProfile, RC)

		switch goos {
		case "windows":
			userPath := os.Getenv("SystemDrive") + `\Users\` + selectedProfile
			tasks = []task{
				// ==================== WINDOWS USER-CLEANUP ====================
				{desc: "~\\...\\Windows\\Explorer (build-in)", goFunc: func() error { return cleanFolder(userPath + `\AppData\Local\Microsoft\Windows\Explorer`) }},
				{desc: "~\\...\\Local\\Temp (build-in)", goFunc: func() error { return cleanFolder(userPath + `\AppData\Local\Temp`) }},
			}
		default:
			userPath := "/home/" + selectedProfile
			tasks = []task{
				// ==================== LINUX USER-CLEANUP ====================
				{desc: "~/.../share/Trash (build-in)", goFunc: func() error { return cleanFolder(userPath + "/.local/share/Trash") }},
				{desc: "~/.cache (build-in)", goFunc: func() error { return cleanFolder(userPath + "/.cache") }},
				{desc: "~/.thumbnails (build-in)", goFunc: func() error { return cleanFolder(userPath + "/.thumbnails") }},
			}
		}
	// ==================== SYSTEM CLEANUP ==================== //
	default:
		switch goos {
		// ==================== WINDOWS CLEANUP ==================== //
		case "windows": // Windows
			switch mode {
			default:
				tasks = []task{
					// ==================== WINDOWS SAFE-CLEANUP ====================
					// ==================== BASICS ====================
					{desc: "%TEMP% (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("TEMP")) }},
					{desc: "\\...\\Windows\\Explorer (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("LocalAppData") + `\Microsoft\Windows\Explorer`) }},
					{desc: "\\...\\FontCache (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("LocalAppData") + `\FontCache`) }},
					// ==================== EXTRAS ====================
					{desc: "DNS Cache (shell)", cmd: []string{"ipconfig", "/flushdns"}},
				}
			case "full":
				tasks = []task{
					// ==================== WINDOWS FULL-CLEANUP ====================
					// ==================== BASICS ====================
					{desc: "%TEMP% (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("TEMP")) }},
					{desc: "\\...\\Windows\\Explorer (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("LocalAppData") + `\Microsoft\Windows\Explorer`) }},
					{desc: "\\...\\FontCache (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("LocalAppData") + `\FontCache`) }},
					{desc: "\\...\\winevt\\Logs (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\System32\winevt\Logs`) }},
					{desc: "\\...\\WindowsUpdate\\Logs (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("ProgramData") + `\Microsoft\Windows\WindowsUpdate\Logs`) }},
					{desc: "\\...\\Logs (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\Logs`) }},
					{desc: "\\...\\Temp (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\Temp`) }},
					{desc: "Admin Trash (shell) (ignore errors)", cmd: []string{"powershell", "-Command", "Clear-RecycleBin -Force -Confirm:$false -ErrorAction SilentlyContinue"}},
					{desc: "Windows Update Cache (shell)", cmd: []string{"powershell", "Remove-Item -Path $env:SystemRoot\\SoftwareDistribution\\Download\\* -Recurse -Force"}},
					// ==================== EXTRAS ====================
					{desc: "\\...\\Prefetch (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\Prefetch`) }},
					{desc: "DNS Cache (shell)", cmd: []string{"ipconfig", "/flushdns"}},
				}
			}

		// ==================== LINUX CLEANUP ==================== //
		default: // Linux
			switch mode {
			default:
				tasks = []task{
					//   ==================== LINUX SAFE-CLEANUP ====================
					//   ==================== BASICS ====================
					{desc: "/tmp (build-in)", goFunc: func() error { return cleanFolder("/tmp") }},
					{desc: "Journal Logs (100 days) (shell)", cmd: []string{"journalctl", "--vacuum-time=100d"}},
					// ==================== EXTRAS ====================
					{desc: "fc-cache (shell)", cmd: []string{"fc-cache", "-fr"}},
					{desc: "Apt Cache (shell)", cmd: []string{"apt-get", "clean"}},
					{desc: "Flatpak Cache (shell)", cmd: []string{"flatpak", "uninstall", "--unused", "-y"}},
					{desc: "Pacman Cache (shell)", cmd: []string{"pacman", "-Scc", "--noconfirm"}},
					{desc: "DNF Cache (shell)", cmd: []string{"sudo", "dnf", "clean", "all"}},
					{desc: "DNS Cache (shell)", cmd: []string{"systemd-resolve", "--flush-caches"}},
				}
			case "full":
				tasks = []task{
					//   ==================== LINUX FULL-CLEANUP ====================
					//   ==================== BASICS ====================
					{desc: "/tmp (build-in)", goFunc: func() error { return cleanFolder("/tmp") }},
					{desc: "/var/tmp (build-in)", goFunc: func() error { return cleanFolder("/var/tmp") }},
					{desc: "/var/cache (build-in)", goFunc: func() error { return cleanFolder("/var/cache") }},
					{desc: "Root Trash (build-in)", goFunc: func() error { return cleanFolder(os.Getenv("HOME") + "/.local/share/Trash") }},
					{desc: "All System Logs (10 days) (shell)", cmd: []string{"sh", "-c", "find /var/log -type f -mtime +10 -exec rm -f {} +"}},
					// ==================== EXTRAS ====================
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
			}
		}
	}

	// Ask user if verbose logging should be enabled
	if !skipPause {
		fmt.Print("Enable verbose logging? (yes/NO): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))
		if choice == "y" || choice == "yes" {
			verbose = true
		}
	}

	// Preview what will be executed
	printInfo("The following cleanup tasks will be executed:")
	for _, t := range tasks {
		cmdStr := strings.Join(t.cmd, " ")
		if t.goFunc != nil && cmdStr == "" {
			cmdStr = "(build-in)"
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

	fmt.Printf("%s#############################################%s\n", RED, RC)
	printInfo("*** CrunchyCleaner Cleanup STARTED ***\n")
	time.Sleep(2 * time.Second)

	// Execute all tasks
	startFree := getFreeMB()
	for _, t := range tasks {
		ctx, cancel := context.WithCancel(context.Background())
		go asyncSpinner(ctx, "Cleaning: "+t.desc)
		time.Sleep(CMDWAIT) // A little pause to actually see what the hell is going on

		var err error
		if t.goFunc != nil {
			_ = t.goFunc() // Execute Go cleanup
		} else if len(t.cmd) > 0 {
			_, err = runCommand(t.cmd) // Execute cmd command
		}
		cancel()
		fmt.Printf("\r\033[2K") // Clear spinner line

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
