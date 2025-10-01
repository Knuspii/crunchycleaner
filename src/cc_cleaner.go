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
	path   string
}

// cleanFolder deletes **everything inside** a folder but keeps the folder itself alive
func cleanFolder(folder string) error {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := folder + string(os.PathSeparator) + entry.Name()
		os.RemoveAll(path)
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
			pause()
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
			{
				desc:   "Windows Explorer Cache",
				path:   userPath + `\AppData\Local\Microsoft\Windows\Explorer`,
				goFunc: func() error { return cleanFolder(userPath + `\AppData\Local\Microsoft\Windows\Explorer`) },
			},
			{
				desc:   "Local Crash Dumps",
				path:   userPath + `\AppData\Local\CrashDumps`,
				goFunc: func() error { return cleanFolder(userPath + `\AppData\Local\CrashDumps`) },
			},
			{
				desc:   "Temp Folder",
				path:   userPath + `\AppData\Local\Temp`,
				goFunc: func() error { return cleanFolder(userPath + `\AppData\Local\Temp`) },
			},
		}
	default:
		userPath := "/home/" + profile
		return []task{
			{
				desc:   "Cache Folder",
				path:   userPath + "/.cache",
				goFunc: func() error { return cleanFolder(userPath + "/.cache") },
			},
			{
				desc:   "Thumbnails",
				path:   userPath + "/.thumbnails",
				goFunc: func() error { return cleanFolder(userPath + "/.thumbnails") },
			},
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
			{
				desc:   "Windows Event Logs",
				path:   os.Getenv("SystemRoot") + `\System32\winevt\Logs`,
				goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\System32\winevt\Logs`) },
			},
			{
				desc:   "Windows Update Logs",
				path:   os.Getenv("ProgramData") + `\Microsoft\Windows\WindowsUpdate\Logs`,
				goFunc: func() error { return cleanFolder(os.Getenv("ProgramData") + `\Microsoft\Windows\WindowsUpdate\Logs`) },
			},
			{
				desc:   "Windows Logs",
				path:   os.Getenv("SystemRoot") + `\Logs`,
				goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\Logs`) },
			},
			{
				desc:   "Temp Folder",
				path:   os.Getenv("SystemRoot") + `\Temp`,
				goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\Temp`) },
			},
			{
				desc:   "Prefetch Folder",
				path:   os.Getenv("SystemRoot") + `\Prefetch`,
				goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\Prefetch`) },
			},
			{
				desc:   "Windows WER",
				path:   os.Getenv("ProgramData") + `\Microsoft\Windows\WER`,
				goFunc: func() error { return cleanFolder(os.Getenv("ProgramData") + `\Microsoft\Windows\WER`) },
			},
			{
				desc:   "WDI Log Files",
				path:   os.Getenv("SystemRoot") + `\System32\WDI\LogFiles`,
				goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\System32\WDI\LogFiles`) },
			},
			{
				desc: "Defender CacheManager",
				path: os.Getenv("ProgramData") + `\Microsoft\Windows Defender\Scans\History\CacheManager`,
				goFunc: func() error {
					return cleanFolder(os.Getenv("ProgramData") + `\Microsoft\Windows Defender\Scans\History\CacheManager`)
				},
			},
			{
				desc:   "CBS Logs",
				path:   os.Getenv("SystemRoot") + `\Logs\CBS`,
				goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\Logs\CBS`) },
			},
			{
				desc: "Delivery Optimization",
				path: os.Getenv("SystemRoot") + `\SoftwareDistribution\DeliveryOptimization`,
				goFunc: func() error {
					return cleanFolder(os.Getenv("SystemRoot") + `\SoftwareDistribution\DeliveryOptimization`)
				},
			},
			{
				desc:   "Windows.old",
				path:   `C:\Windows.old`,
				goFunc: func() error { return cleanFolder(`C:\Windows.old`) },
			},
			{
				desc:   "SoftwareDistribution Download",
				path:   os.Getenv("SystemRoot") + `\SoftwareDistribution\Download`,
				goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\SoftwareDistribution\Download`) },
			},
			{
				desc:   "Driver FileRepository",
				path:   os.Getenv("SystemRoot") + `\System32\DriverStore\FileRepository`,
				goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\System32\DriverStore\FileRepository`) },
			},
		}
	default:
		return []task{
			{
				desc:   "Temp Folder",
				path:   os.Getenv("SystemRoot") + `\Temp`,
				goFunc: func() error { return cleanFolder(os.Getenv("SystemRoot") + `\Temp`) },
			},
			{
				desc:   "Windows Update Logs",
				path:   os.Getenv("ProgramData") + `\Microsoft\Windows\WindowsUpdate\Logs`,
				goFunc: func() error { return cleanFolder(os.Getenv("ProgramData") + `\Microsoft\Windows\WindowsUpdate\Logs`) },
			},
			{
				desc: "Defender CacheManager",
				path: os.Getenv("ProgramData") + `\Microsoft\Windows Defender\Scans\History\CacheManager`,
				goFunc: func() error {
					return cleanFolder(os.Getenv("ProgramData") + `\Microsoft\Windows Defender\Scans\History\CacheManager`)
				},
			},
		}
	}
}

// buildLinuxTasks returns Linux cleanup tasks
func buildLinuxTasks(mode string) []task {
	switch mode {
	case "full":
		return []task{
			{
				desc:   "Temp",
				path:   "/tmp",
				goFunc: func() error { return cleanFolder("/tmp") },
			},
			{
				desc:   "Var Temp",
				path:   "/var/tmp",
				goFunc: func() error { return cleanFolder("/var/tmp") },
			},
			{
				desc:   "Var Cache",
				path:   "/var/cache",
				goFunc: func() error { return cleanFolder("/var/cache") },
			},
			{
				desc:   "Systemd Coredump",
				path:   "/var/lib/systemd/coredump",
				goFunc: func() error { return cleanFolder("/var/lib/systemd/coredump") },
			},
			{
				desc:   "Var Crash",
				path:   "/var/crash",
				goFunc: func() error { return cleanFolder("/var/crash") },
			},
			{
				desc: "All System Logs (>10 days)",
				cmd:  []string{"sh", "-c", "find /var/log -type f -mtime +10 -exec rm -f {} +"},
			},
			{
				desc: "fc-cache",
				cmd:  []string{"fc-cache", "-fr"},
			},
			{
				desc: "Apt Cache",
				cmd:  []string{"apt-get", "clean"},
			},
			{
				desc: "Flatpak Cache",
				cmd:  []string{"flatpak", "uninstall", "--unused", "-y"},
			},
			{
				desc: "Pip Cache",
				cmd:  []string{"pip", "cache", "purge"},
			},
			{
				desc: "Npm Cache",
				cmd:  []string{"npm", "cache", "clean", "--force"},
			},
			{
				desc: "Yarn Cache",
				cmd:  []string{"yarn", "cache", "clean"},
			},
			{
				desc: "DNF Cache",
				cmd:  []string{"dnf", "clean", "all"},
			},
			{
				desc: "Pacman Cache",
				cmd:  []string{"pacman", "-Scc", "--noconfirm"},
			},
			{
				desc: "Nix Garbage Collector",
				cmd:  []string{"nix-collect-garbage", "-d"},
			},
			{
				desc: "Composer Cache",
				cmd:  []string{"composer", "clear-cache"},
			},
			{
				desc: "Go Module Cache",
				cmd:  []string{"go", "clean", "-modcache"},
			},
			{
				desc: "Rust Cargo Cache",
				cmd:  []string{"cargo", "clean"},
			},
			{
				desc: "Docker System Prune",
				cmd:  []string{"docker", "system", "prune", "-af"},
			},
			{
				desc: "Podman System Prune",
				cmd:  []string{"podman", "system", "prune", "-af"},
			},
			{
				desc: "Systemd-Tmpfiles",
				cmd:  []string{"systemd-tmpfiles", "--clean"},
			},
		}
	default:
		return []task{
			{
				desc:   "Temp",
				path:   "/tmp",
				goFunc: func() error { return cleanFolder("/tmp") },
			},
			{
				desc: "Journal Logs (>100 days)",
				cmd:  []string{"journalctl", "--vacuum-time=100d"},
			},
			{
				desc: "fc-cache",
				cmd:  []string{"fc-cache", "-fr"},
			},
			{
				desc: "Apt Cache",
				cmd:  []string{"apt-get", "clean"},
			},
			{
				desc: "Flatpak Cache",
				cmd:  []string{"flatpak", "uninstall", "--unused", "-y"},
			},
			{
				desc: "Pacman Cache",
				cmd:  []string{"pacman", "-Scc", "--noconfirm"},
			},
			{
				desc: "DNF Cache",
				cmd:  []string{"dnf", "clean", "all"},
			},
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
		var detail string
		if t.goFunc != nil {
			if t.path != "" {
				detail = t.path // zeigt den Pfad automatisch an
			} else {
				detail = "(built-in)"
			}
		} else if len(t.cmd) > 0 {
			detail = strings.Join(t.cmd, " ")
		}

		fmt.Printf("%s- Cleaning: %s → %s%s\n", CYAN, t.desc, detail, RC)
	}

	fmt.Println()
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
