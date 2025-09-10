// ##################################
// CrunchyCleaner
// Made by: Knuspii (M)
// ##################################

package main

import (
	"bufio"     // For reading user input
	"context"   // For controlling goroutines (e.g., stopping the spinner)
	"fmt"       // Formatted input/output
	"os"        // General OS interactions (exit, files, etc.)
	"os/exec"   // Executing external commands
	"os/signal" // Handling file paths in a cross-platform way
	"runtime"   // Info about the OS / architecture
	"strings"   // String manipulation (Trim, Split, Join, etc.)
	"syscall"   // System calls (for signal handling)
	"time"      // Time-related functions (sleep, timestamp, timeout)
)

const (
	CC_VERSION = "1.1"
	COLS       = 62
	LINES      = 30
	CMDWAIT    = 1 * time.Second        // Wait time running a command
	PROMPT     = (YELLOW + " >>:" + RC) // Prompt string displayed to the user
	RED        = "\033[31m"
	YELLOW     = "\033[33m"
	GREEN      = "\033[32m"
	BLUE       = "\033[34m"
	CYAN       = "\033[36m"
	RC         = "\033[0m" // Reset ANSI color
)

var (
	origCols, origLines int            // Original terminal size
	consoleRunning      = true         // Controls main loop
	verbose             = false        // If true, print all errors
	selectedProfile     = ""           // Username for user cleanup
	skipPause           = false        // If true, skip pause
	goos                = runtime.GOOS // Current OS
	reader              = bufio.NewReader(os.Stdin)
	//SPINNERFRAMES  = []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'} // Spinner animation frames
	SPINNERFRAMES = []rune{'|', '/', '-', '\\'} // Spinner animation frames
)

func adminCheck() {
	switch goos {
	// WINDOWS
	case "windows":
		// Try running a command that requires admin rights
		cmd := exec.Command("net", "session")
		if err := cmd.Run(); err != nil {
			printError("You need admin privileges")
			os.Exit(1)
		}
	default:
		if os.Geteuid() != 0 { // Check if current user is not root
			printError("You need root privileges")
			os.Exit(1)
		}
	}
}

func getAdmin() {
	switch goos {
	// WINDOWS
	case "windows":
		// Try running a command that requires admin privileges
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
	// LINUX
	default:
		if os.Geteuid() != 0 { // Check if current user is not root
			printInfo("Requesting root privileges...\n")

			// Get absolute path to the running executable
			scriptPath, err := exec.LookPath(os.Args[0])
			if err != nil {
				printError(fmt.Sprintf("Could not find executable: %v", err))
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
}

func getTermSize() (cols, lines int, err error) {
	switch goos {
	case "windows":
		// Windows: get with PowerShell
		cmd := exec.Command("powershell", "-Command",
			`$size = $Host.UI.RawUI.WindowSize; Write-Output "$($size.Width) $($size.Height)"`)
		out, err := cmd.Output()
		if err != nil {
			return 0, 0, err
		}
		fmt.Sscanf(string(out), "%d %d", &cols, &lines)
	default:
		// Unix: use stty
		cmd := exec.Command("stty", "size")
		cmd.Stdin = os.Stdin
		out, err := cmd.Output()
		if err != nil {
			return 0, 0, err
		}
		fmt.Sscanf(string(out), "%d %d", &lines, &cols)
	}
	return
}

func setTermSize(cols, lines int) {
	switch goos {
	case "windows":
		psCmd := fmt.Sprintf(
			`$Host.UI.RawUI.WindowSize = New-Object System.Management.Automation.Host.Size(%d, %d); $Host.UI.RawUI.BufferSize = New-Object System.Management.Automation.Host.Size(%d, 300)`,
			cols, lines, cols,
		)
		_ = exec.Command("powershell", "-Command", psCmd).Run()
	default:
		fmt.Printf("\033[8;%d;%dt", lines, cols)
	}
}

func init_term() {
	// Save original size
	cols, lines, err := getTermSize()
	if err == nil {
		origCols, origLines = cols, lines
	}

	// Resize to program defaults
	setTermSize(COLS, LINES)
}

func restoreTerm() {
	if origCols > 0 && origLines > 0 {
		setTermSize(origCols, origLines)
	}
}

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
	fmt.Printf(" %s[fullclean]%s - Does a Full-Cleanup\n", YELLOW, RC)
	fmt.Printf(" %s[safeclean]%s - Does a Safe-Cleanup\n", YELLOW, RC)
	fmt.Printf(" %s[userclean]%s - Does a User-Cleanup\n", YELLOW, RC)
	fmt.Printf(" %s[info]%s      - Shows some infos\n", YELLOW, RC)
	fmt.Printf(" %s[reset]%s     - Reset the TUI\n", YELLOW, RC)
	fmt.Printf(" %s[exit]%s      - Exit\n", RED, RC)
	line()
}

func usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("  %scrunchycleaner [option]%s\n\n", CYAN, RC)
	fmt.Printf("Options:\n")
	fmt.Printf("  %sNo option%s   Run with TUI (Text-UI)\n", YELLOW, RC)
	fmt.Printf("  %s-t%s          Run with TUI (Text-UI)\n", YELLOW, RC)
	fmt.Printf("  %s-s%s          Run Safe-Cleanup\n", YELLOW, RC)
	fmt.Printf("  %s-sy%s         Run Safe-Cleanup (non-interactive for scripts)\n", YELLOW, RC)
	fmt.Printf("  %s-f%s          Run Full-Cleanup\n", YELLOW, RC)
	fmt.Printf("  %s-fy%s         Run Full-Cleanup (non-interactive for scripts)\n", YELLOW, RC)
	fmt.Printf("  %s-u <user> %s  Run User-Cleanup\n", YELLOW, RC)
	fmt.Printf("  %s-uy <user>%s  Run User-Cleanup (non-interactive for scripts)\n", YELLOW, RC)
	fmt.Printf("  %s-v%s          Show version\n", YELLOW, RC)
	fmt.Printf("  %s-h%s          Show this help page\n", YELLOW, RC)
}

func normalstartup() {
	adminCheck()
	showBanner()
}

func skipstartup() {
	verbose = true
	skipPause = true
	adminCheck()
	showBanner()
}

func handleargs() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		// Safe-Cleanup
		case "-s":
			normalstartup()
			cleanup("safe")
			os.Exit(0)
		// Safe-Cleanup (non-interactive)
		case "-sy":
			skipstartup()
			cleanup("safe")
			os.Exit(0)
		// Full-Cleanup
		case "-f":
			normalstartup()
			cleanup("full")
			os.Exit(0)
		// Full-Cleanup (non-interactive)
		case "-fy":
			skipstartup()
			cleanup("full")
			os.Exit(0)
		// User-Cleanup
		case "-u":
			adminCheck()
			if len(os.Args) > 2 {
				showBanner()
				cleanup("user", os.Args[2])
			} else {
				printError("No profile name provided")
				os.Exit(1)
			}
			os.Exit(0)
		// User-Cleanup (non-interactive)
		case "-uy":
			skipPause = true
			verbose = true
			adminCheck()
			if len(os.Args) > 2 {
				showBanner()
				cleanup("user", os.Args[2])
			} else {
				printError("No profile name provided")
				os.Exit(1)
			}
			os.Exit(0)
		// Help
		case "-h", "--help":
			showBanner()
			usage()
			os.Exit(0)
		// Version
		case "-v", "--version":
			fmt.Printf("CrunchyCleaner %s\n", CC_VERSION)
			os.Exit(0)
		// TUI
		case "-t":
			// Just continue to TUI
		default:
			fmt.Printf("Unknown option: %s\n", os.Args[1])
			usage()
			os.Exit(1)
		}
	}
}

func handlecommands() {
	fmt.Print("Enter command" + PROMPT)
	cmd, _ := reader.ReadString('\n')
	cmd = strings.TrimSpace(cmd)
	cmdline()
	switch cmd {
	// FULL CLEAN
	case "fullclean", ",full clean", "full cleanup", "clean full", "cleanup full":
		cleanup("full")
		pause()
		consoleRunning = false

	// SAFE CLEAN
	case "safeclean", "safe clean", "safe cleanup", "clean safe", "cleanup safe":
		cleanup("safe")
		pause()
		consoleRunning = false

	// USER CLEAN
	case "userclean", "user clean", "user cleanup", "clean user", "cleanup user":
		cleanup("user")
		pause()
		consoleRunning = false

	// HELP
	case "help", "h":
		printInfo("Just type a command from the list above")
		pause()

	// INFO
	case "i", "info", "infos", "about", "version":
		usage()
		fmt.Printf(`
%sCrunchyCleaner Version: %s%s

DISCLAIMER:
MADE BY: Knuspii, (M)
Help by the World Wide Web.
A lightweight, cross-platform system cleanup tool.
You use this tool at your own risk.
I do not take any responsibilities.
This work is licensed under the:
Creative Commons Attribution-
NonCommercial 4.0 International License.
https://github.com/Knuspii/crunchycleaner
`, YELLOW, CC_VERSION, RC)
		pause()

	// RESET
	case "r", "reset", "refresh", "reload", "clear":
		ctx, cancel := context.WithCancel(context.Background())
		go asyncSpinner(ctx, "Reloading...")
		time.Sleep(CMDWAIT)
		cancel()
		clearScreen()
		getAdmin()
		init_term()
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

func main() {
	handleargs()
	// Catch [Ctrl+C] and restore
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		restoreTerm()
		os.Exit(0)
	}()
	getAdmin()
	init_term()
	showBanner()
	showCommands()

	for consoleRunning {
		handlecommands()
	}
	restoreTerm()
	printTask("EXITED\n")
	os.Exit(0)
}
