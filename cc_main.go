// ##################################################################
// CrunchyCleaner
// Made by: Knuspii, (M)
// Project: https://github.com/Knuspii/CrunchyCleaner
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
// ##################################################################

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
)

// Global constants for UI and Versioning
const (
	CC_VERSION = "2.8"
	COLS       = 56
	LINES      = 32
	GOOS       = runtime.GOOS
	CLEARLINE  = "\r\033[K"
	YELLOW     = "\033[33m"
	CYAN       = "\033[36m"
	GREEN      = "\033[32m"
	RC         = "\033[0m"
)

var (
	// CLI Flags
	Flagversion = flag.Bool("v", false, "Display version information")
	Flagdryrun  = flag.Bool("d", false, "Dry-run mode without deleting files (for testing)")
	Flagauto    = flag.Bool("a", false, "Automate cleaning (select all and start immediately)")
)

// Program represents a target application and its associated cache directories
type Program struct {
	Name    string
	Paths   []string // List of paths (supports wildcards/globbing)
	Checked bool     // Selection state in the menu
}

// ========================= HELPER FUNCTIONS =========================

// cc_exit provides a clean termination of the application
func cc_exit() {
	// Close keyboard
	keyboard.Close()

	// Enable cursor
	fmt.Print("\033[?25h")

	fmt.Printf("\nExiting CrunchyCleaner...\n")
	os.Exit(0)
}

func pause() {
	fmt.Printf("\nPress [ENTER] to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// line draws a formatted horizontal separator
func line() {
	fmt.Printf("%s#%s~%s\n", YELLOW, strings.Repeat("=", COLS-2), RC)
}

// spinner visualizes background tasks and cleans up properly
func spinner(text string, stop chan bool, ack chan bool) {
	frames := []string{"|", "/", "-", "\\"}
	i := 0
	for {
		select {
		case <-stop:
			// Clear the line and move cursor to start
			fmt.Print(CLEARLINE)
			// Signal back to main that we are done
			ack <- true
			return
		default:
			fmt.Printf("\r%s%s%s %s%s %s%s%s ", YELLOW, frames[i%len(frames)], RC, CYAN, text, YELLOW, frames[i%len(frames)], RC)
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}

// ========================= MENU UI LOGIC =========================

func showBanner() {
	_, total, free := getDiskMetrics()
	fmt.Printf(`%s  ____________________     .-.
 |   |  |       __ |  \    |_|
 |   |  |      |  ||  |    | |
 |   |  |      |__||  |    |=|
 |   |__|__________|  |  .=/I\=.
 |                    | ////V\\\\
 |   ______________   | |#######|
 |  |______________|  | |||||||||
 |  |              |  |
 |  |              |  | %sCrunchyCleaner%s
 |  |              |  | Made by: Knuspii, (M)
 |[]|              |[]| Version: %s
 |__|______________|__| Disk: %s / %s%s
`, YELLOW, RC, YELLOW, CC_VERSION, free, total, RC)
	line()
}

func logInfo(msg string) { fmt.Printf("%s[+] %s%s\n", CYAN, msg, RC) }
func logOK(msg string)   { fmt.Printf("%s[✓] %s%s\n", GREEN, msg, RC) }
func logWarn(msg string) { fmt.Printf("%s[!] %s%s\n", YELLOW, msg, RC) }

// renderMenu draws the interactive selection list
func renderMenu(existing []Program, idx int, fullRedraw bool) {
	if fullRedraw {
		showBanner()
		fmt.Printf("↑/↓ or W/S to navigate | ENTER to select | C to clean\n")
		fmt.Printf("Folders found: %d\n", len(existing))
	}

	// Render each detected program entry
	for i := range existing {
		cursor := "   "
		// Highlight the currently selected entry
		if i == idx {
			cursor = YELLOW + " >_" + RC
		}
		// Checkbox indicator for selection state
		check := "[ ]"
		if existing[i].Checked {
			check = "[" + GREEN + "X" + RC + "]"
		}
		// Clear the current line and print the menu entry
		fmt.Printf("%s%s%s %s\n", CLEARLINE, cursor, check, existing[i].Name)
	}
}

// handleMenu manages user input for navigation and selection
func handleMenu() {
	// Initial scan of the filesystem to find existing directories
	stop := make(chan bool)
	ack := make(chan bool)
	go spinner("Starting CrunchyCleaner & Scanning filesystem", stop, ack)
	time.Sleep(1 * time.Second)

	existing := scanForExisting()

	stop <- true // Tell spinner to stop
	<-ack        // WAIT for spinner to clear the line

	// Abort if nothing was detected
	if len(existing) == 0 {
		fmt.Printf("\nNo cache directories found on your system")
		pause()
		return
	}

	// Enable raw keyboard input mode
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer keyboard.Close()

	idx := 0
	renderMenu(existing, idx, true)
	// Main Input Loop
	// This might be fucked up but it works lol
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			break
		}
		updated := false

		// Navigation and selection controls
		if key == keyboard.KeyArrowUp || char == 'w' || char == 'W' {
			if idx > 0 {
				idx--
				updated = true
			}
		} else if key == keyboard.KeyArrowDown || char == 's' || char == 'S' {
			if idx < len(existing)-1 {
				idx++
				updated = true
			}
		} else if char == ' ' || key == keyboard.KeyEnter || key == keyboard.KeySpace {
			existing[idx].Checked = !existing[idx].Checked
			updated = true
		} else if char == 'a' || char == 'A' {
			// Toggle "Select All" logic
			allChecked := true
			for _, p := range existing {
				if !p.Checked {
					allChecked = false
					break
				}
			}
			for i := range existing {
				existing[i].Checked = !allChecked
			}
			updated = true
		} else if char == 'c' || char == 'C' {
			runCleanup(existing)
		} else if key == keyboard.KeyCtrlC || char == 'q' || char == 'Q' {
			cc_exit()
		}

		// Redraw menu entries in-place if state changed
		if updated {
			// Move cursor up to the start of the menu list
			fmt.Printf("\033[%dA", len(existing))
			renderMenu(existing, idx, false)
		}
	}
}

func runCleanup(programs []Program) {
	beforeFree, _, _ := getDiskMetrics()

	if *Flagdryrun {
		fmt.Printf("\n%sNOTE: Dry run active. No files will actually be deleted.%s", YELLOW, RC)
	} else {
		fmt.Printf("\nYou use this tool at your own risk!")
	}
	// Fetching the current system user
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("\nUsername: unknown")
	} else {
		name := usr.Username
		// Strip domain/machine name prefix on Windows
		if GOOS == "windows" && strings.Contains(name, "\\") {
			parts := strings.Split(name, "\\")
			name = parts[len(parts)-1]
		}
		fmt.Printf("\nUsername: %s", name)
	}
	//fmt.Printf("\nPress [CTRL+C] to cancel")
	fmt.Printf("\nCleaning caches started...\n")

	stop := make(chan bool)
	ack := make(chan bool)
	go spinner("Cleaning selected caches", stop, ack)
	time.Sleep(3 * time.Second)

	count := 0

	for _, p := range programs {
		if !p.Checked {
			continue
		}
		count++

		for _, path := range p.Paths {
			matches, _ := filepath.Glob(expandHome(path))

			for _, m := range matches {
				if *Flagdryrun {
					fmt.Print(CLEARLINE)
					logInfo("Would clean: " + m)
					continue
				}
				deletePath(m)
			}
		}

		// Cut the size part
		name := p.Name
		if idx := strings.Index(name, "("); idx != -1 {
			name = strings.TrimSpace(name[:idx])
		}
		fmt.Print(CLEARLINE)
		logOK(name)
	}

	stop <- true
	<-ack

	if count == 0 {
		fmt.Printf("\nNothing selected")
		time.Sleep(3 * time.Second)
		cc_exit()
	}

	if *Flagdryrun {
		logOK("Dry-Run finished")
	} else {
		logOK("Cleaning finished")
	}

	afterFree, _, _ := getDiskMetrics()
	cleaned := (afterFree - beforeFree) * 1024
	if cleaned < 0 || *Flagdryrun {
		cleaned = 0
	}

	line()
	if *Flagdryrun {
		fmt.Printf("CrunchyCleaner cleaned: NOTHING (DRY-RUN)\n")
	} else {
		fmt.Printf("CrunchyCleaner cleaned: %s%.2f MB%s\n", YELLOW, cleaned, RC)
	}

	if !*Flagauto {
		keyboard.Close()
		fmt.Printf("\nPress [ENTER] to exit")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
	cc_exit()
}

func main() {
	flag.Parse()

	// Capture OS Interrupts (like Ctrl+C) for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cc_exit()
	}()

	if *Flagversion {
		fmt.Printf("CrunchyCleaner %s\n", CC_VERSION)
		return
	}

	// AUTOMATION LOGIC
	if *Flagauto {
		showBanner()
		fmt.Printf("%sNOTE: Automation active. Scanning and selecting all caches...%s\n", YELLOW, RC)
		existing := scanForExisting()

		// Check all found items
		for i := range existing {
			existing[i].Checked = true
		}
		runCleanup(existing)
	}

	// Run interactive mode
	handleMenu()
}
