// ##################################
// CrunchyCleaner
// Made by: Knuspii (M)
// ##################################

package main

import (
	"bufio"         // For reading user input
	"context"       // For controlling goroutines (e.g., stopping the spinner)
	"errors"        // For creating custom errors
	"fmt"           // Formatted input/output
	"os"            // General OS interactions (exit, files, etc.)
	"os/exec"       // Executing external commands
	"path/filepath" // Handling file paths in a cross-platform way
	"runtime"       // Info about the OS / architecture
	"strconv"       // Converting strings to numbers and vice versa
	"strings"       // String manipulation (Trim, Split, Join, etc.)
	"time"          // Time-related functions (sleep, timestamp, timeout)
)

const (
	CC_VERSION = "0.3"
	COLS       = 62
	LINES      = 30
	CMDWAIT    = 2 * time.Second        // Wait time running a command
	PROMPT     = (YELLOW + " >>:" + RC) // Prompt string displayed to the user
	RED        = "\033[31m"
	YELLOW     = "\033[33m"
	GREEN      = "\033[32m"
	BLUE       = "\033[34m"
	CYAN       = "\033[36m"
	RC         = "\033[0m" // Reset ANSI color
)

var (
	consoleRunning  = true
	selectedProfile = ""
	skipPause       = false
	goos            = runtime.GOOS
	reader          = bufio.NewReader(os.Stdin)
	//SPINNERFRAMES  = []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}
	SPINNERFRAMES = []rune{'|', '/', '-', '\\'}
)

// ============================================================ HELPER FUNCTIONS ============================================================
func printInfo(msg string) {
	fmt.Printf("%s[INFO]%s %s\n", YELLOW, RC, msg)
}
func printError(msg string) {
	fmt.Printf("%s[ERROR]%s %s\n", RED, RC, msg)
}
func printSuccess(msg string) {
	fmt.Printf("\n%s[INFO]%s %s\n", GREEN, RC, msg)
}
func printTask(msg string) {
	fmt.Printf("%s[*]%s %s\n", GREEN, RC, msg)
}

func line() {
	fmt.Printf("%s#%s~%s\n", YELLOW, strings.Repeat("-", COLS-2), RC)
}
func cmdline() {
	fmt.Printf("%s#%s~%s\n", RED, strings.Repeat("-", COLS-2), RC)
}

// Pause waits for enter
func pause() {
	fmt.Printf("\nPress [ENTER] to continue: ")
	reader.ReadString('\n')
}

// asyncSpinner displays a spinning "loading" animation in the terminal.
// It runs asynchronously and stops when the provided context is canceled.
func asyncSpinner(ctx context.Context, text string) {
	i := 0 // Index for spinner frames
	for {
		select {
		// If the context is canceled, stop the spinner and return
		case <-ctx.Done():
			return

		// Default case: continue spinning
		default:
			// Print spinner line:
			// \r       -> Carriage return to overwrite the same line
			// [LOADING] -> Static label
			// YELLOW/RC -> Apply color and reset
			// text      -> Custom text passed to the spinner
			// SPINNERFRAMES[i%len(SPINNERFRAMES)] -> Rotate through spinner characters
			fmt.Printf("\r%s[LOADING]%s %s %c  ", YELLOW, RC, text, SPINNERFRAMES[i%len(SPINNERFRAMES)])
			time.Sleep(100 * time.Millisecond) // Wait a short time before next frame
			i++                                // Move to the next spinner frame
		}
	}
}

// runCommand executes an external command and returns its combined output (stdout + stderr).
// It takes a slice of strings, where the first element is the command and the rest are arguments.
func runCommand(cmd []string) (string, error) {
	// Check if the command slice is empty
	if len(cmd) == 0 {
		return "", errors.New("command is empty")
	}

	// Create an exec.Command object with the command and its arguments
	c := exec.Command(cmd[0], cmd[1:]...)

	// Run the command and capture both stdout and stderr
	outBytes, err := c.CombinedOutput()

	// Convert output bytes to string and trim whitespace/newlines
	out := strings.TrimSpace(string(outBytes))

	// If the command failed, return the output along with a formatted error
	if err != nil {
		return out, fmt.Errorf(
			"command '%s' failed: %v\nOutput: %s",
			strings.Join(cmd, " "), // Reconstruct the command for the error message
			err,
			out,
		)
	}

	// If successful, return the output and nil error
	return out, nil
}

func clearScreen() {
	var cmd *exec.Cmd
	if goos == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Fallback: ANSI for Clear + Reset Cursor
		fmt.Print("\033[H\033[2J")
	}
}

// getDiskInfo returns (total, free) disk space as human-readable strings
func getDiskInfo() (string, string) {
	var diskTotal, diskFree string

	if goos == "windows" {
		// Windows: PowerShell query
		sizeOut, _ := exec.Command("powershell", "-Command",
			"(Get-PSDrive -PSProvider FileSystem | Where-Object {$_.Name -eq 'C'}).Used, (Get-PSDrive -PSProvider FileSystem | Where-Object {$_.Name -eq 'C'}).Free").Output()
		parts := strings.Fields(string(sizeOut))
		if len(parts) >= 2 {
			used, _ := strconv.ParseFloat(parts[0], 64)
			free, _ := strconv.ParseFloat(parts[1], 64)
			total := used + free
			diskTotal = fmt.Sprintf("%.0f MB", total/1024/1024)
			diskFree = fmt.Sprintf("%.0f MB", free/1024/1024)
		}
	} else {
		// Linux / Unix: df
		dfOut, _ := exec.Command("sh", "-c", "df -BM --output=size,avail / | tail -1 | tr -d 'M'").Output()
		parts := strings.Fields(string(dfOut))
		if len(parts) >= 2 {
			diskTotal = parts[0] + " MB"
			diskFree = parts[1] + " MB"
		}
	}

	return diskTotal, diskFree
}

// getFreeMB returns free disk space in MB (numeric, for diff calculation)
func getFreeMB() int64 {
	if goos == "windows" {
		out, _ := exec.Command("powershell", "-Command",
			"(Get-PSDrive -PSProvider FileSystem | Where-Object {$_.Name -eq 'C'}).Free").Output()
		free, _ := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
		return free / 1024 / 1024
	}

	out, _ := exec.Command("sh", "-c", "df --output=avail / | tail -1").Output()
	free, _ := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	return free / 1024
}

// ============================================================ CORE FUNCTIONS ============================================================
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

	// Execute all tasks
	startFree := getFreeMB()

	// Preview what will be executed
	printInfo("The following cleanup tasks will be executed:")
	for _, t := range tasks {
		fmt.Printf("%s- %s%s\n  → %s\n", CYAN, t.desc, RC, strings.Join(t.cmd, " "))
	}
	fmt.Printf("\n")
	printInfo("The above cleanup tasks will be executed")
	printInfo("!!! You use this tool at your own risk !!!")
	printInfo("You can ignore most errors. Press [CTRL+C] to cancel")
	if !skipPause {
		pause()
	}

	fmt.Printf("%s------------------------------%s\n", RED, RC)
	printInfo("*** Cleanup STARTED ***\n")
	time.Sleep(2 * time.Second)

	// Now run them
	for _, t := range tasks {
		ctx, cancel := context.WithCancel(context.Background())
		go asyncSpinner(ctx, "Running: "+t.desc)
		time.Sleep(CMDWAIT)

		_, err := runCommand(t.cmd)
		cancel()
		fmt.Printf("\r\033[2K") // Clear spinner line

		if err != nil {
			printTask(t.desc + " FINISHED*")
			fmt.Printf(" Notice: %s\n", err)
		} else {
			printTask(t.desc + " FINISHED")
		}
		time.Sleep(200 * time.Millisecond)
	}
	endFree := getFreeMB()
	diff := endFree - startFree
	printInfo(fmt.Sprintf("Cleaned approx %s%d MB%s disk space", YELLOW, diff, RC))
	printSuccess("*** Cleanup FINISHED ***")
}

// ============================================================ CONSOLE / STARTUP ============================================================
func startup() {
	switch goos {
	case "windows":
		// Try running a command that requires admin rights
		cmd := exec.Command("net", "session")
		if err := cmd.Run(); err != nil {
			// If it fails, try to restart the program as admin
			printInfo("Restarting as admin...\n")

			elevate := exec.Command("powershell", "-Command", "Start-Process", os.Args[0], "-Verb", "RunAs")

			// Connect the elevated process output to the current console
			elevate.Stdout = os.Stdout
			elevate.Stderr = os.Stderr

			if err := elevate.Run(); err != nil {
				// Failed to elevate privileges
				printError(fmt.Sprintf("Failed to restart as admin: %v", err))
				fmt.Println("CrunchyCleaner might not work correctly without admin rights.")
				pause()
			} else {
				// Successfully elevated, exit current process
				os.Exit(0)
			}
		}

	default: // Unix-like systems
		if os.Geteuid() != 0 { // Check if current user is not root
			printInfo("Requesting root privileges...\n")

			// Get absolute path to the running executable
			scriptPath, err := filepath.Abs(os.Args[0])
			if err != nil {
				printError(fmt.Sprintf("Could not get own path: %v", err))
				return
			}

			// Prepare command with all original arguments
			args := append([]string{scriptPath}, os.Args[1:]...)
			cmd := exec.Command("sudo", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin // Needed to read sudo password

			if err := cmd.Run(); err != nil {
				// User canceled or sudo failed
				if strings.Contains(err.Error(), "interrupt") {
					fmt.Println("\nCancelled by user")
					os.Exit(0)
				}
				printError(fmt.Sprintf("Failed to restart with sudo: %v", err))
				fmt.Print("Press [ENTER] to continue anyway: ")
				reader.ReadString('\n')
			} else {
				// Successfully elevated, exit current process
				os.Exit(0)
			}
		}
	}

	switch goos {
	case "windows":
		// Set PowerShell window size and buffer
		psCmd := fmt.Sprintf(
			`$Host.UI.RawUI.WindowSize = New-Object System.Management.Automation.Host.Size(%d, %d); $Host.UI.RawUI.BufferSize = New-Object System.Management.Automation.Host.Size(%d, 300)`,
			COLS, LINES, COLS,
		)
		_, err := runCommand([]string{"powershell", "-Command", psCmd})
		if err != nil {
			printError("Failed to set window size: " + err.Error())
		}

	default:
		// Set Unix terminal size using ANSI escape codes
		fmt.Printf("\033[8;%d;%dt", LINES, COLS)
	}
}

// showBanner prints ASCII art and system info
func showBanner() {
	total, free := getDiskInfo()
	fmt.Printf(`%s  ____________________     .-.
 |  |              |  |    |_|
 |[]|              |[]|    | |
 |  |              |  |    |=|
 |  |              |  |  .=/I\=.
 |  |              |  | ////V\\\\
 |  |______________|  | |#######|
 |                    | |||||||||
 |     ____________   |
 |    | __      |  |  | CrunchyCleaner - Cleanup your system!
 |    ||  |     |  |  | Made by: Knuspii, (M)
 |    ||__|     |  |  | Version: %s
 |____|_________|__|__| Disk-Space: %s / %s%s
`, YELLOW, CC_VERSION, free, total, RC)
	line()
}

func showCommands() {
	fmt.Printf("Commands:\n")
	fmt.Printf(" %s[full clean]%s - Does a Full-Cleanup\n", YELLOW, RC)
	fmt.Printf(" %s[safe clean]%s - Does a Safe-Cleanup\n", YELLOW, RC)
	fmt.Printf(" %s[user clean]%s - Does a User-Cleanup\n", YELLOW, RC)
	fmt.Printf(" %s[info]%s       - Shows some infos\n", YELLOW, RC)
	fmt.Printf(" %s[reset]%s      - Reset the TUI\n", YELLOW, RC)
	fmt.Printf(" %s[exit]%s       - Exit\n", RED, RC)
	line()
}

func usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("  %scrunchycleaner [option]%s\n\n", CYAN, RC)
	fmt.Printf("Options:\n")
	fmt.Printf("  %sNo option%s   Run with TUI\n", YELLOW, RC)
	fmt.Printf("  %s-s%s          Run Safe-Cleanup\n", YELLOW, RC)
	fmt.Printf("  %s-f%s          Run Full-Cleanup\n", YELLOW, RC)
	fmt.Printf("  %s-u {user}%s   Run User-Cleanup\n", YELLOW, RC)
	fmt.Printf("  %s-f%s          Show version\n", YELLOW, RC)
	fmt.Printf("  %s-h%s          Show this help page\n", YELLOW, RC)
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-s":
			skipPause = true
			showBanner()
			cleanup("safe")
			os.Exit(0)
		case "-f":
			skipPause = true
			showBanner()
			cleanup("full")
			os.Exit(0)
		case "-u":
			skipPause = true
			if len(os.Args) > 2 {
				showBanner()
				cleanup("user", os.Args[2])
			} else {
				printError("No profile name provided")
				os.Exit(1)
			}
			os.Exit(0)
		case "-h", "--help":
			showBanner()
			usage()
			os.Exit(0)
		case "-v", "--version":
			fmt.Printf("CrunchyCleaner %s\n", CC_VERSION)
			os.Exit(0)
		default:
			fmt.Printf("Unknown option: %s\n", os.Args[1])
			usage()
			os.Exit(1)
		}
	}
	startup()
	showBanner()
	showCommands()

	for consoleRunning {
		fmt.Print("Enter command" + PROMPT)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		cmdline()

		switch cmd {
		// FULL CLEAN
		case "full clean", "full cleanup", "clean full", "cleanup full":
			cleanup("full")
			consoleRunning = false

		// SAFE CLEAN
		case "safe clean", "safe cleanup", "clean safe", "cleanup safe":
			cleanup("safe")
			consoleRunning = false

		// USER CLEAN
		case "user clean", "user cleanup", "clean user", "cleanup user":
			cleanup("user")
			consoleRunning = false

		// HELP
		case "help", "h":
			printInfo("Just type a command from the list above")
			pause()

		// INFO
		case "i", "info", "infos", "about", "version":
			fmt.Printf(`CrunchyCleaner Version: %s

DISCLAIMER:
MADE BY: Knuspii, (M)
Help by the World Wide Web.
Made with: Go, Bash, Powershell
Simple program with various functions.
You use this tool at your own risk.
I do not take any responsibilities.
https://github.com/Knuspii/crunchycleaner
`, CC_VERSION)
			consoleRunning = false

		// RESET
		case "r", "reset", "refresh", "reload", "clear":
			ctx, cancel := context.WithCancel(context.Background())
			go asyncSpinner(ctx, "Reloading...")
			time.Sleep(CMDWAIT)
			cancel()
			clearScreen()
			startup()
			showBanner()
			showCommands()

		// EXIT
		case "e", "q", "quit", "exit":
			ctx, cancel := context.WithCancel(context.Background())
			go asyncSpinner(ctx, "Exiting...")
			time.Sleep(CMDWAIT)
			cancel()
			fmt.Printf("\r\033[2K")
			consoleRunning = false

		// DEFAULT
		case "":
			printInfo("Input a command")
			pause()
		default:
			printInfo("Invalid command: " + cmd)
			pause()
		}
	}
	pause()
	printTask("EXITED\n")
	os.Exit(0)
}
