// CrunchyCleaner: Core Cleaner Function

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

func cleanup(mode string, username ...string) {
	// Define a cleanup task with a description and the command to execute
	type task struct {
		desc string   // Description of the task (for logging/spinner)
		cmd  []string // Command and arguments to run
	}

	var tasks []task
	switch mode {
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

		if len(profiles) == 0 {
			printError("No user profiles found!")
			return
		}

		// If cli-argument
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
				return
			}

		} else {
			// Interaktive Auswahl
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
				return
			}
			selectedProfile = choice
		}

		fmt.Printf("%sSelected profile: %s%s\n", YELLOW, selectedProfile, RC)

		switch goos {
		case "windows":
			userPath := os.Getenv("SystemDrive") + `\Users\` + selectedProfile
			tasks = []task{
				{"Cleaning Temp Files", []string{"powershell", "-Command", fmt.Sprintf("Get-ChildItem -Path '%s\\AppData\\Local\\Temp' | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue", userPath)}},
				{"Cleaning Thumbnail Cache", []string{"powershell", "-Command", fmt.Sprintf("Remove-Item -Path '%s\\AppData\\Local\\Microsoft\\Windows\\Explorer\\thumbcache_*' -Force -ErrorAction SilentlyContinue", userPath)}},
				{"Cleaning Recycle Bin", []string{"powershell", "(New-Object -ComObject Shell.Application).NameSpace(10).Items() | ForEach-Object { Remove-Item $_.Path -Force -Recurse -ErrorAction SilentlyContinue }"}},
			}

		default:
			userPath := "/home/" + selectedProfile
			tasks = []task{
				{"Cleaning Temp Files", []string{"sh", "-c", fmt.Sprintf("rm -rf %s/tmp/*", userPath)}},
				{"Cleaning Thumbnail Cache", []string{"sh", "-c", fmt.Sprintf("rm -rf %s/.cache/thumbnails/*", userPath)}},
				{"Cleaning Trash", []string{"sh", "-c", fmt.Sprintf("rm -rf %s/.local/share/Trash/*", userPath)}},
			}
		}
	default:
		switch goos {
		case "windows": // Windows
			switch mode {
			default:
				tasks = []task{
					// Windows Safe-Cleanup
					// Basic
					{"Cleaning Delivery Optimization", []string{"powershell", "Remove-Item", "$env:SystemDrive\\ProgramData\\Microsoft\\Network\\Downloader\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Cleaning Explorer MRU", []string{"powershell", "Remove-Item", "$env:APPDATA\\Microsoft\\Windows\\Recent\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					// Windows
					{"Cleaning Windows Update Cache", []string{"powershell", "Remove-Item", "C:\\Windows\\SoftwareDistribution\\Download\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					// Extra
					{"(ipconfig) Flushing DNS Cache", []string{"ipconfig", "/flushdns"}},
				}
			case "full":
				tasks = []task{
					// Windows Full-Cleanup
					// Basic
					{"Stopping Windows Update service", []string{"powershell", "Stop-Service", "-Name", "wuauserv", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Stopping BITS service", []string{"powershell", "Stop-Service", "-Name", "bits", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Cleaning Recycle Bin", []string{"powershell", "(New-Object -ComObject Shell.Application).NameSpace(10).Items() | ForEach-Object { Remove-Item $_.Path -Force -Recurse -ErrorAction SilentlyContinue }"}},
					{"Cleaning Temp Files", []string{"powershell", "-Command", "Get-ChildItem -Path $env:TEMP | ForEach-Object { Remove-Item $_.FullName -Recurse -Force -ErrorAction SilentlyContinue }"}},
					{"Cleaning Thumbnail Cache", []string{"powershell", "Remove-Item", "$env:LOCALAPPDATA\\Microsoft\\Windows\\Explorer\\thumbcache_*", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Cleaning Explorer MRU", []string{"powershell", "Remove-Item", "$env:APPDATA\\Microsoft\\Windows\\Recent\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Cleaning Delivery Optimization", []string{"powershell", "Remove-Item", "$env:SystemDrive\\ProgramData\\Microsoft\\Network\\Downloader\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Cleaning Error Reporting", []string{"powershell", "Remove-Item", "$env:ProgramData\\Microsoft\\Windows\\WER\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Cleaning Prefetch", []string{"powershell", "Remove-Item", "C:\\Windows\\Prefetch\\*", "-Force", "-Recurse", "-ErrorAction", "SilentlyContinue"}},
					// Windows
					{"Cleaning Windows Update Cache", []string{"powershell", "Remove-Item", "C:\\Windows\\SoftwareDistribution\\Download\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Cleaning Windows Installer Cache", []string{"powershell", "Remove-Item", "C:\\Windows\\Installer\\$PatchCache$\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Cleaning Windows Temp (WinDir)", []string{"powershell", "Remove-Item", "C:\\Windows\\Temp\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					{"Cleaning Windows Temp Files", []string{"powershell", "Remove-Item", "$env:LOCALAPPDATA\\Microsoft\\Windows\\Caches\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
					// Extra
					{"(dism) Cleaning Old Windows Updates", []string{"dism", "/Online", "/Cleanup-Image", "/StartComponentCleanup", "/Quiet"}},
					{"(ipconfig) Flushing DNS Cache", []string{"ipconfig", "/flushdns"}},
				}
			}

		default: // Linux / Unix-like
			switch mode {
			default:
				tasks = []task{
					// Linux Safe Cleanup
					// Basic
					{"Cleaning System Logs (older than 90 days)", []string{"journalctl", "--vacuum-time=90d"}},
					// Extra
					{"Cleaning Apt Cache", []string{"sudo", "apt-get", "clean"}},
					{"Cleaning Flatpak Cache", []string{"flatpak", "uninstall", "--unused", "-y"}},
					{"Cleaning Pip Cache", []string{"pip", "cache", "purge"}},
					{"Cleaning Npm Cache", []string{"npm", "cache", "clean", "--force"}},
					{"Cleaning Yarn Cache", []string{"yarn", "cache", "clean"}},
					{"Cleaning Snap Cache", []string{"sudo", "rm", "-rf", "/var/cache/snapd/*"}},
					{"Cleaning DNF Cache", []string{"sh", "-c", "rm -rf /var/cache/dnf/*"}},
					{"Cleaning Pacman Cache", []string{"sh", "-c", "rm -rf /var/cache/pacman/pkg/*"}},
				}
			case "full":
				tasks = []task{
					// Linux Full Cleanup
					// Basic
					{"Cleaning Thumbnail Cache", []string{"sh", "-c", "rm -rf ~/.cache/thumbnails/*"}},
					{"Cleaning System Logs (all)", []string{"sh", "-c", "rm -rf /var/log/journal/*"}},
					{"Cleaning Trash", []string{"sh", "-c", "rm -rf ~/.local/share/Trash/*"}},
					{"Cleaning Temp Files", []string{"sh", "-c", "rm -rf /tmp/*"}},
					{"Cleaning Var Temp Files", []string{"sh", "-c", "rm -rf /var/tmp/*"}},
					// Extra
					{"Cleaning Apt Cache", []string{"sudo", "apt-get", "clean"}},
					{"Cleaning Flatpak Cache", []string{"flatpak", "uninstall", "--unused", "-y"}},
					{"Cleaning Pip Cache", []string{"pip", "cache", "purge"}},
					{"Cleaning Npm Cache", []string{"npm", "cache", "clean", "--force"}},
					{"Cleaning Yarn Cache", []string{"yarn", "cache", "clean"}},
					{"Cleaning Snap Cache", []string{"sudo", "rm", "-rf", "/var/cache/snapd/*"}},
					{"Cleaning DNF Cache", []string{"sh", "-c", "rm -rf /var/cache/dnf/*"}},
					{"Cleaning Pacman Cache", []string{"sh", "-c", "rm -rf /var/cache/pacman/pkg/*"}},
					{"Running Nix Garbage Collector", []string{"nix-collect-garbage", "-d"}},
				}
			}
		}
	}

	// Preview what will be executed
	printInfo("The following cleanup tasks will be executed:")
	for _, t := range tasks {
		fmt.Printf("%s- %s%s\n  → %s\n", CYAN, t.desc, RC, strings.Join(t.cmd, " "))
	}
	fmt.Printf("\n")
	printInfo("The above cleanup tasks will be executed")
	printInfo("!!! You use this tool at your own risk !!!")
	if !skipPause {
		pause()
	}

	fmt.Printf("%s##############################%s\n", RED, RC)
	printInfo("*** Cleanup STARTED ***\n")
	time.Sleep(2 * time.Second)

	// Execute all tasks
	startFree := getFreeMB()
	for _, t := range tasks {
		ctx, cancel := context.WithCancel(context.Background())
		go asyncSpinner(ctx, "- "+t.desc)
		time.Sleep(CMDWAIT)

		_, err := runCommand(t.cmd)
		cancel()
		fmt.Printf("\r\033[2K") // Clear spinner line

		if err != nil {
			printTask(t.desc + " FINISHED*")
		} else {
			printTask(t.desc + " FINISHED")
		}
		time.Sleep(200 * time.Millisecond)
	}
	endFree := getFreeMB()
	diff := endFree - startFree
	if diff < 0 {
		diff = 0
	}
	printInfo(fmt.Sprintf("Cleaned approx %s%d MB%s disk space", YELLOW, diff, RC))
	printSuccess("*** Cleanup FINISHED ***")
}
