package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	CNWS_VERSION       = "0.5"
	COLS               = 51
	barMaxLen          = 8  // number of blocks
	pingHistoryLen     = 20 // number of ping points
	RED                = "\033[31m"
	YELLOW             = "\033[33m"
	GREEN              = "\033[32m"
	BLUE               = "\033[34m"
	CYAN               = "\033[36m"
	RC                 = "\033[0m" // Reset ANSI color
	DefaultRefresh     = 10 * time.Second
	DefaultRefreshGame = 60 * time.Second
)

var (
	signalBlocks  = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	emptyBlock    = ' '
	goos          = runtime.GOOS // Current OS
	pingHistory   []int
	SPINNERFRAMES = []rune{'|', '/', '-', '\\'} // Spinner animation frames
	refreshrate   = DefaultRefresh
	enableGame    = false
)

// ---------------- Data Structures ----------------
type WiFiNetwork struct {
	SSID     string
	BSSID    string
	Security string
	Signal   int // %
}

func usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("  %scrunchycleaner [option]%s\n\n", CYAN, RC)
	fmt.Printf("Options:\n")
	fmt.Printf("  %s-r <seconds>%s  Set refresh rate in seconds\n", YELLOW, RC)
	fmt.Printf("  %s-g%s            Enable EXP / Level-up system\n", YELLOW, RC)
	fmt.Printf("  %s-v%s            Show version and exit\n", YELLOW, RC)
	fmt.Printf("  %s-h%s            Show this help\n", YELLOW, RC)
}

func handleargs() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		// Refresh rate
		case "-r":
			if len(os.Args) > 2 {
				seconds, err := strconv.Atoi(os.Args[2])
				if err != nil {
					fmt.Println("Invalid refreshrate:", err)
					return
				}
				refreshrate = time.Duration(seconds) * time.Second
			} else {
				fmt.Printf("No profile name provided\n")
				os.Exit(1)
			}
		// Enable Game
		case "-g":
			enableGame = true
			refreshrate = DefaultRefreshGame
		// Help
		case "-h", "--help":
			usage()
			os.Exit(0)
		// Version
		case "-v", "--version":
			fmt.Printf("CrunchyNWS %s\n", CNWS_VERSION)
			os.Exit(0)
		default:
			fmt.Printf("Unknown option: %s\n", os.Args[1])
			usage()
			os.Exit(1)
		}
	}
}

// ---------------- Main ----------------
func main() {
	handleargs()
	pingctx, pingcancel := context.WithCancel(context.Background())
	scanctx, scancancel := context.WithCancel(context.Background())
	if enableGame {
		loadPet()
	}
	for {
		clearScreen()
		fmt.Printf(`
     \ | /       
     - * -       
      /|\        
     /\|/\       
    /  |  \      
   /\/\|/\/\    CrunchyNWS %s
  /    |    \   CrunchyNetworkScanner
 -     -     -	Made by: Knuspii, (M)
`, CNWS_VERSION)

		line()
		if enableGame {
			displayPetStatus()
		}
		// Ping
		go asyncSpinner(pingctx, "Pinging...")
		ms := pingCloudflare()
		if len(pingHistory) >= pingHistoryLen {
			pingHistory = pingHistory[1:]
		}
		pingHistory = append(pingHistory, ms)
		pingcancel()
		fmt.Printf("\r\033[2K") // Clear spinner line
		fmt.Printf("PING History: [%s] %d ms\n\n", pingBarGraph(pingHistory), ms)

		// Scan Wi-Fi
		fmt.Printf("[BSSID   [SSID           [SECURITY   [%%  [STRENGTH\n")
		go asyncSpinner(scanctx, "Networks...")
		wifi := scanWiFi()
		scancancel()
		if len(wifi) == 0 {
			fmt.Printf("No Wi-Fi networks found.\n")
		} else {
			for _, n := range wifi {
				if n.SSID == "" {
					continue // skip unnamed SSIDs
				}
				sec := n.Security
				if sec == "" {
					sec = "-"
				}
				bssidSuffix := n.BSSID[len(n.BSSID)-5:]
				fmt.Printf("\r\033[2K") // Clear spinner line
				fmt.Printf("...%-5s %-15s %-11s %2d%% [%s]\n", bssidSuffix, n.SSID, truncate(sec, 11), n.Signal, signalBar(n.Signal))
			}
		}
		scancancel()
		if enableGame {
			fmt.Printf("\nExp-Logs:\n")
			findWiFi(wifi, ms)
		}
		time.Sleep(refreshrate)
	}
}
