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
	CC_VERSION = "0.1"
	COLS       = 70
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
	consoleRunning = true
	goos           = runtime.GOOS
	reader         = bufio.NewReader(os.Stdin)
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
	fmt.Printf("\nPress [Enter] to continue: ")
	reader.ReadString('\n')
}

// yesNo asks question, returns true if yes
func yesNo(question string) bool {
	for {
		fmt.Printf("%s (yes/no)%s", question, PROMPT)
		answer, _ := reader.ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		switch answer {
		case "y", "yes", "ye", "yup", "ja":
			return true
		case "n", "no", "nope", "nah", "ne", "nein":
			return false
		}
	}
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

// getUptime returns the system uptime as a formatted string, e.g., "12H:34M".
// Supports Windows and Linux/Unix systems.
func getUptime() string {
	// Helper function to format a duration into "H:MM" format
	formatUptime := func(d time.Duration) string {
		h := int(d.Hours())          // Total hours
		m := int(d.Minutes()) - h*60 // Remaining minutes
		if h > 999 {                 // Cap hours at 999 for display
			return "+999H"
		}
		return fmt.Sprintf("%dH:%02dM", h, m)
	}

	switch goos {
	case "windows":
		// Windows: get last boot time using PowerShell
		out, err := runCommand([]string{
			"powershell", "-Command",
			"(Get-CimInstance Win32_OperatingSystem).LastBootUpTime.ToString('s')",
		})
		if err != nil {
			return "unknown" // Unable to retrieve boot time
		}

		bootTimeStr := strings.TrimSpace(out)
		layout := "2006-01-02T15:04:05"      // PowerShell ISO-like format
		loc, _ := time.LoadLocation("Local") // Local timezone
		bootTime, err := time.ParseInLocation(layout, bootTimeStr, loc)
		if err != nil {
			return "unknown" // Failed to parse boot time
		}

		// Calculate uptime and format
		return formatUptime(time.Since(bootTime))

	default:
		// Linux/Unix: read boot time from /proc/stat
		data, err := os.ReadFile("/proc/stat")
		if err != nil {
			return "unknown" // Unable to read file
		}

		// Search for "btime" line which contains boot timestamp
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "btime ") {
				parts := strings.Fields(line)
				if len(parts) == 2 {
					sec, err := strconv.ParseInt(parts[1], 10, 64)
					if err != nil {
						return "unknown" // Failed to parse timestamp
					}
					bootTime := time.Unix(sec, 0)
					return formatUptime(time.Since(bootTime))
				}
			}
		}
		return "unknown" // btime not found
	}
}

// ============================================================ CORE FUNCTIONS ============================================================
func cleanup(mode string) {
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
		case "windows": // Windows
			userBase := os.Getenv("SystemDrive") + `\Users`
			files, err := os.ReadDir(userBase)
			if err != nil {
				printError("Reading Users folder: ")
				fmt.Printf("%s\n", err)
				return
			}

			for _, f := range files {
				//if f.IsDir() && f.Name() != "Public" && f.Name() != "Default" && f.Name() != "Default User" {
				if f.IsDir() {
					profiles = append(profiles, f.Name())
				}
			}

		default: // Linux / Unix-like
			// Usually /home for user folders
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

		fmt.Println("Available profiles:")
		fmt.Printf("%s[0]%s %s\n", YELLOW, RC, "Cancel")
		for i, p := range profiles {
			fmt.Printf("%s[%d]%s %s\n", YELLOW, i+1, RC, p)
		}

		fmt.Printf("Select profile number to clean%s", PROMPT)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		choiceInt, err := strconv.Atoi(choice)
		if err != nil {
			fmt.Printf("Invalid input\n")
			return
		}
		//
		if choiceInt == 0 {
			fmt.Printf("User-Cleanup cancelled\n")
			return
		}
		if choiceInt < 1 || choiceInt > len(profiles) {
			fmt.Printf("Invalid choice\n")
			return
		}

		selectedProfile := profiles[choiceInt-1]
		fmt.Printf("%sSelected profile: %s%s\n", YELLOW, selectedProfile, RC)

		switch goos {
		case "windows":
			userPath := os.Getenv("SystemDrive") + `\Users\` + selectedProfile
			tasks = []task{
				{"Cleaning Temp Files", []string{"powershell", "-Command", fmt.Sprintf("Get-ChildItem -Path '%s\\AppData\\Local\\Temp' | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue", userPath)}},
				{"Cleaning Thumbnail Cache", []string{"powershell", "-Command", fmt.Sprintf("Remove-Item -Path '%s\\AppData\\Local\\Microsoft\\Windows\\Explorer\\thumbcache_*' -Force -ErrorAction SilentlyContinue", userPath)}},
			}

		case "linux":
			userPath := "/home/" + selectedProfile
			tasks = []task{
				{"Cleaning Temp Files", []string{"sh", "-c", fmt.Sprintf("rm -rf %s/tmp/*", userPath)}},
				{"Cleaning Thumbnail Cache", []string{"sh", "-c", fmt.Sprintf("rm -rf %s/.cache/thumbnails/*", userPath)}},
				{"Cleaning Trash", []string{"sh", "-c", fmt.Sprintf("rm -rf %s/.local/share/Trash/*", userPath)}},
			}
		}
	default:
		// Fallthrough to safe/full cleanup
	}

	switch goos {
	case "windows": // Windows
		switch mode {
		default:
			tasks = []task{
				// Windows Safe Cleanup
				{"Cleaning Windows Update Cache", []string{"powershell", "Remove-Item", "C:\\Windows\\SoftwareDistribution\\Download\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
				{"Cleaning Delivery Optimization", []string{"powershell", "Remove-Item", "$env:SystemDrive\\ProgramData\\Microsoft\\Network\\Downloader\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
				{"Flushing DNS Cache", []string{"ipconfig", "/flushdns"}},
			}
		case "full":
			tasks = []task{
				// Windows Full Cleanup
				{"Stopping Windows Update service", []string{"powershell", "Stop-Service", "-Name", "wuauserv", "-Force", "-ErrorAction", "SilentlyContinue"}},
				{"Stopping BITS service", []string{"powershell", "Stop-Service", "-Name", "bits", "-Force", "-ErrorAction", "SilentlyContinue"}},
				{"Cleaning Prefetch", []string{"powershell", "Remove-Item", "C:\\Windows\\Prefetch\\*", "-Force", "-Recurse", "-ErrorAction", "SilentlyContinue"}},
				{"Cleaning Error Reporting", []string{"powershell", "Remove-Item", "$env:ProgramData\\Microsoft\\Windows\\WER\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
				{"Cleaning Windows Update Cache", []string{"powershell", "Remove-Item", "C:\\Windows\\SoftwareDistribution\\Download\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
				{"Cleaning Delivery Optimization", []string{"powershell", "Remove-Item", "$env:SystemDrive\\ProgramData\\Microsoft\\Network\\Downloader\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
				{"Cleaning Thumbnail Cache", []string{"powershell", "Remove-Item", "$env:LOCALAPPDATA\\Microsoft\\Windows\\Explorer\\thumbcache_*", "-Force", "-ErrorAction", "SilentlyContinue"}},
				{"Cleaning Recycle Bin", []string{"powershell", "(New-Object -ComObject Shell.Application).NameSpace(10).Items() | ForEach-Object { Remove-Item $_.Path -Force -Recurse -ErrorAction SilentlyContinue }"}},
				{"Cleaning Temp Files", []string{"powershell", "-Command", "Get-ChildItem -Path $env:TEMP | ForEach-Object { Remove-Item $_.FullName -Recurse -Force -ErrorAction SilentlyContinue }"}},
				{"Cleaning Windows Temp Files", []string{"powershell", "Remove-Item", "$env:LOCALAPPDATA\\Microsoft\\Windows\\Caches\\*", "-Recurse", "-Force", "-ErrorAction", "SilentlyContinue"}},
				{"Flushing DNS Cache", []string{"ipconfig", "/flushdns"}},
			}
		}

	default: // Linux / Unix-like
		switch mode {
		default:
			tasks = []task{
				// Linux Safe Cleanup
				{"Cleaning Apt Cache", []string{"sudo", "apt-get", "clean"}},
				{"Cleaning Flatpak Cache", []string{"flatpak", "uninstall", "--unused", "-y"}},
				{"Cleaning Snap Cache", []string{"sudo", "rm", "-rf", "/var/cache/snapd/*"}},
				{"Cleaning DNF Cache", []string{"sh", "-c", "rm -rf /var/cache/dnf/*"}},
				{"Cleaning Pacman Cache", []string{"sh", "-c", "rm -rf /var/cache/pacman/pkg/*"}},
				{"Cleaning System Logs (older than 90 days)", []string{"journalctl", "--vacuum-time=90d"}},
			}
		case "full":
			tasks = []task{
				// Linux Full Cleanup
				{"Cleaning Apt Cache", []string{"sudo", "apt-get", "clean"}},
				{"Cleaning Flatpak Cache", []string{"flatpak", "uninstall", "--unused", "-y"}},
				{"Cleaning Snap Cache", []string{"sudo", "rm", "-rf", "/var/cache/snapd/*"}},
				{"Cleaning DNF Cache", []string{"sh", "-c", "rm -rf /var/cache/dnf/*"}},
				{"Cleaning Pacman Cache", []string{"sh", "-c", "rm -rf /var/cache/pacman/pkg/*"}},
				{"Running Nix Garbage Collector", []string{"nix-collect-garbage", "-d"}},
				{"Cleaning Thumbnail Cache", []string{"sh", "-c", "rm -rf ~/.cache/thumbnails/*"}},
				{"Cleaning System Logs (older than 7 days)", []string{"journalctl", "--vacuum-time=7d"}},
				{"Cleaning Trash", []string{"sh", "-c", "rm -rf ~/.local/share/Trash/*"}},
				{"Cleaning Temp Files", []string{"sh", "-c", "rm -rf /tmp/*"}},
			}
		}
	}

	// Execute all tasks
	for _, t := range tasks {
		ctx, cancel := context.WithCancel(context.Background())
		go asyncSpinner(ctx, "Running: "+t.desc)
		time.Sleep(CMDWAIT)

		_, err := runCommand(t.cmd)
		cancel()
		fmt.Printf("\r\033[2K") // Clear spinner line

		if err != nil {
			printError(t.desc + " FAILED")
			fmt.Printf("  Error: %s\n", err)
		} else {
			printTask(t.desc + " FINISHED")
		}

		time.Sleep(200 * time.Millisecond)
	}
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

			elevate := exec.Command("powershell", "-Command",
				"Start-Process", os.Args[0], "-Verb", "RunAs")

			// Connect the elevated process output to the current console
			elevate.Stdout = os.Stdout
			elevate.Stderr = os.Stderr

			if err := elevate.Run(); err != nil {
				// Failed to elevate privileges
				printError(fmt.Sprintf("Failed to restart as admin: %v", err))
				fmt.Println("CrunchyCleaner might not work correctly without admin rights.")
				fmt.Print("Press [Enter] to continue anyway: ")
				reader.ReadString('\n')
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
	uptime := getUptime()
	if uptime == "unknown" || uptime == "" {
		printError("Could not determine system uptime, using fallback")
		uptime = "0H:00M"
	}
	line()
	fmt.Print(YELLOW + `░░      ░░░       ░░░  ░░░░  ░░   ░░░  ░░░      ░░░  ░░░░  ░░  ░░░░  ░
▒  ▒▒▒▒  ▒▒  ▒▒▒▒  ▒▒  ▒▒▒▒  ▒▒    ▒▒  ▒▒  ▒▒▒▒  ▒▒  ▒▒▒▒  ▒▒▒  ▒▒  ▒▒
▓  ▓▓▓▓▓▓▓▓       ▓▓▓  ▓▓▓▓  ▓▓  ▓  ▓  ▓▓  ▓▓▓▓▓▓▓▓        ▓▓▓▓    ▓▓▓
█  ████  ██  ███  ███  ████  ██  ██    ██  ████  ██  ████  █████  ████
██      ███  ████  ███      ███  ███   ███      ███  ████  █████  ████
  ______   __        ________   ______   ___   __  ________  _______
░░      ░░░  ░░░░░░░░        ░░░      ░░░   ░░░  ░░        ░░       ░░
▒  ▒▒▒▒  ▒▒  ▒▒▒▒▒▒▒▒  ▒▒▒▒▒▒▒▒  ▒▒▒▒  ▒▒    ▒▒  ▒▒  ▒▒▒▒▒▒▒▒  ▒▒▒▒  ▒
▓  ▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓      ▓▓▓▓  ▓▓▓▓  ▓▓  ▓  ▓  ▓▓      ▓▓▓▓       ▓▓
█  ████  ██  ████████  ████████        ██  ██    ██  ████████  ███  ██
██      ███        ██        ██  ████  ██  ███   ██        ██  ████  █
` + RC)
	line()
	fmt.Printf(" | CrunchyCleaner - Cleanup your system!\n")
	fmt.Printf(" | Made by: Knuspii, (M)\n")
	fmt.Printf(" | Version: %s\n", CC_VERSION)
	fmt.Printf(" | Uptime:  %s\n", uptime)
	line()
	fmt.Printf("Commands:\n")
	fmt.Printf(" %s[full clean]%s - Does a Full-Cleanup\n", YELLOW, RC)
	fmt.Printf(" %s[safe clean]%s - Does a Safe-Cleanup\n", YELLOW, RC)
	fmt.Printf(" %s[user clean]%s - Does a User-Cleanup\n", YELLOW, RC)
	fmt.Printf(" %s[info]%s       - Shows some infos\n", YELLOW, RC)
	fmt.Printf(" %s[reset]%s      - Reset the TUI\n", YELLOW, RC)
	fmt.Printf(" %s[exit]%s       - Exit\n", RED, RC)
	line()
}

func main() {
	startup()
	showBanner()
	for consoleRunning {
		fmt.Print("Enter command" + PROMPT)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)

		switch cmd {
		// FULL CLEAN
		case "full clean", "full cleanup", "clean full", "cleanup full":
			cmdname := "Full-Cleanup"
			disclaimer := "!!! You use this tool at your own risk !!!"
			yesnoquestion := "Are you sure you want to do a " + cmdname + "?"
			startinfo := cmdname + " STARTED"
			finishinfo := cmdname + " FINISHED"
			cancelinfo := cmdname + " cancelled"

			printInfo(disclaimer)
			if yesNo(yesnoquestion) {
				cmdline()
				printInfo(startinfo)
				printInfo("You can ignore most errors")
				time.Sleep(CMDWAIT)
				cleanup("full")
				printSuccess(finishinfo)
			} else {
				fmt.Printf("%s\n", cancelinfo)
			}
			consoleRunning = false

		// SAFE CLEAN
		case "safe clean", "safe cleanup", "clean safe", "cleanup safe":
			cmdname := "Safe-Cleanup"
			disclaimer := "!!! You use this tool at your own risk !!!"
			yesnoquestion := "Are you sure you want to do a " + cmdname + "?"
			startinfo := cmdname + " STARTED"
			finishinfo := cmdname + " FINISHED"
			cancelinfo := cmdname + " cancelled"

			printInfo(disclaimer)
			if yesNo(yesnoquestion) {
				cmdline()
				printInfo(startinfo)
				printInfo("You can ignore most errors")
				time.Sleep(CMDWAIT)
				cleanup("safe")
				printSuccess(finishinfo)
			} else {
				fmt.Printf("%s\n", cancelinfo)
			}
			consoleRunning = false

		// USER CLEAN
		case "user clean", "user cleanup", "clean user", "cleanup user":
			cmdname := "User-Cleanup"
			disclaimer := "!!! You use this tool at your own risk !!!"
			yesnoquestion := "Are you sure you want to do a " + cmdname + "?"
			startinfo := cmdname + " STARTED"
			finishinfo := cmdname + " FINISHED"
			cancelinfo := cmdname + " cancelled"

			printInfo(disclaimer)
			if yesNo(yesnoquestion) {
				cmdline()
				printInfo(startinfo)
				printInfo("You can ignore most errors")
				cleanup("user")
				printSuccess(finishinfo)
			} else {
				fmt.Printf("%s\n", cancelinfo)
			}
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
