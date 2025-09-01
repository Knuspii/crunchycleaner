// ##################################
// CrunchyCleaner
// Made by: Knuspii (M)
// ##################################

package main

import (
	"bufio"         // For reading user input
	"context"       // For controlling goroutines (e.g., stopping the spinner)
	"fmt"           // Formatted input/output
	"os"            // General OS interactions (exit, files, etc.)
	"os/exec"       // Executing external commands
	"path/filepath" // Handling file paths in a cross-platform way
	"runtime"       // Info about the OS / architecture
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
