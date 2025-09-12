package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// clearScreen clears the terminal screen for neat output
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

func line() {
	fmt.Printf("%s#%s~%s\n", YELLOW, strings.Repeat("-", COLS-2), RC)
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
			// \r        -> Carriage return to overwrite the same line
			// [LOADING] -> Static label
			// YELLOW/RC -> Apply color and reset
			// text      -> Custom text passed to the spinner
			// SPINNERFRAMES[i%len(SPINNERFRAMES)] -> Rotate through spinner characters
			fmt.Printf("\r%s[LOADING]%s %s %s%c%s  ", YELLOW, RC, text, YELLOW, SPINNERFRAMES[i%len(SPINNERFRAMES)], RC)
			time.Sleep(100 * time.Millisecond) // Wait a short time before next frame
			i++                                // Move to the next spinner frame
		}
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func signalBar(percent int) string {
	idx := percent * barMaxLen / 100
	if percent > 90 {
		idx = barMaxLen
	}
	if idx > barMaxLen {
		idx = barMaxLen
	}

	bar := ""
	for i := 0; i < idx; i++ {
		bar += string(signalBlocks[i])
	}
	for i := idx; i < barMaxLen; i++ {
		bar += string(emptyBlock)
	}

	color := RED
	if percent > 70 {
		color = GREEN
	} else if percent > 30 {
		color = YELLOW
	}
	return color + bar + RC
}

func pingBarGraph(history []int) string {
	bar := ""
	maxLen := 20
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	if len(history) > maxLen {
		history = history[len(history)-maxLen:]
	}

	for _, val := range history {
		if val > 300 {
			val = 300
		}
		idx := val * len(blocks) / 300
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		block := blocks[idx]

		color := GREEN
		switch {
		case val >= 300:
			color = RED
		case val >= 90:
			color = YELLOW
		}
		bar += string(color) + string(block) + RC
	}

	for i := len(history); i < maxLen; i++ {
		bar += " "
	}
	return bar
}

func pingCloudflare() int {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows Ping: 1 Ping, nur Zeilen mit TTL
		cmd = exec.Command("cmd", "/C", `ping -n 1 1.1.1.1 | findstr TTL=`)
	} else {
		// Linux/macOS Ping: 1 Ping
		cmd = exec.Command("bash", "-c", `ping -c 1 1.1.1.1 | grep 'icmp_seq=1' | sed -E 's/.* ([0-9.]+) ms$/\1/'`)
	}

	out, err := cmd.Output()
	if err != nil {
		return 999
	}

	line := strings.TrimSpace(string(out))
	if line == "" {
		return 999
	}

	if runtime.GOOS == "windows" {
		// Windows: vorletztes Feld ist "Zeit=XXms"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			val := parts[len(parts)-2]             // z.B. "Zeit=31ms"
			val = strings.TrimPrefix(val, "Zeit=") // nur "31ms"
			val = strings.TrimSuffix(val, "ms")    // nur "31"
			ms, err := strconv.Atoi(val)
			if err == nil {
				return ms
			}
		}
		return 999
	}
	// Linux/macOS: einfach alle Zahlen vor "ms" extrahieren, egal welche Sprache
	re := regexp.MustCompile(`([0-9]+(\.[0-9]+)?)\s*ms`)
	match := re.FindStringSubmatch(line)
	if len(match) >= 2 {
		msFloat, err := strconv.ParseFloat(match[1], 64)
		if err == nil {
			return int(msFloat + 0.5)
		}
	}
	return 999
}

func splitEscaped(line string) []string {
	var parts []string
	current := ""
	escape := false

	for _, r := range line {
		if escape {
			current += string(r)
			escape = false
			continue
		}
		if r == '\\' {
			escape = true
			continue
		}
		if r == ':' {
			parts = append(parts, current)
			current = ""
			continue
		}
		current += string(r)
	}
	parts = append(parts, current)
	return parts
}

func scanWiFi() []WiFiNetwork {
	var networks []WiFiNetwork

	if runtime.GOOS == "windows" {
		// Windows: netsh WLAN Scan
		out, err := exec.Command("cmd", "/C", "netsh wlan show networks mode=Bssid").Output()
		if err != nil {
			return networks
		}
		lines := strings.Split(string(out), "\n")
		var n WiFiNetwork
		for _, line := range lines {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "SSID") && !strings.HasPrefix(line, "SSID-Broadcast") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					ssid := strings.TrimSpace(parts[1])
					if ssid != "" {
						n.SSID = truncate(ssid, 15)
					}
				}
			} else if strings.HasPrefix(line, "Aut") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					n.Security = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "BSSID") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					n.BSSID = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "Sig") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					val, _ := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(parts[1]), "%"))
					n.Signal = val
					if n.SSID != "" && n.BSSID != "" {
						networks = append(networks, n)
					}
					n = WiFiNetwork{} // Reset für nächsten Eintrag
				}
			}
		}

	} else {
		// Linux/macOS mit nmcli
		out, err := exec.Command("bash", "-c", `nmcli -t -f SSID,BSSID,SECURITY,SIGNAL dev wifi`).Output()
		if err != nil {
			return networks
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := splitEscaped(line)
			if len(parts) < 4 {
				continue
			}
			ssid := strings.TrimSpace(parts[0])
			bssid := strings.TrimSpace(parts[1])
			security := strings.TrimSpace(parts[2])
			sig, _ := strconv.Atoi(parts[3])

			if ssid == "" || bssid == "" {
				continue
			}

			networks = append(networks, WiFiNetwork{
				SSID:     truncate(ssid, 15),
				BSSID:    bssid,
				Security: truncate(security, 10),
				Signal:   sig,
			})
		}
	}
	return networks
}
