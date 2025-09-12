package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type PetSave struct {
	Level    int
	Points   int
	FoundAPs map[string]FoundAP // key = BSSID
	Created  time.Time
}

type FoundAP struct {
	SSID     string    `json:"ssid"`
	Security string    `json:"security"`
	Time     time.Time `json:"time"`
}

var pet PetSave

const saveFile = "cnws_save.json"

// Load/Save
func loadPet() {
	data, err := os.ReadFile(saveFile)
	if err != nil {
		pet = PetSave{
			Level:    0,
			Points:   0,
			FoundAPs: make(map[string]FoundAP),
			Created:  time.Now(),
		}
		return
	}
	json.Unmarshal(data, &pet)
}

func savePet() {
	data, _ := json.MarshalIndent(pet, "", "  ")
	os.WriteFile(saveFile, data, 0644)
}

// Find WiFi & log discovered SSIDs briefly
func findWiFi(networks []WiFiNetwork, ping int) {
	newFound := 0
	// Zuerst prüfen, wie viele APs neu sind
	for _, n := range networks {
		if _, ok := pet.FoundAPs[n.BSSID]; !ok {
			newFound++
		}
	}

	if newFound > 0 {
		multiplier := 1

		// Ping Bonus
		if ping < 20 {
			multiplier++
			fmt.Printf("[+] Ping bonus! (%d ms)\n", ping)
		}

		// AP Count Bonus
		if len(networks) > 10 {
			multiplier++
			fmt.Printf("[+] AP bonus! (%d visible networks)\n", len(networks))
		}
		if len(networks) > 20 {
			multiplier++
			fmt.Printf("[+] Big AP bonus! (%d visible networks)\n", len(networks))
		}

		// Punkte vergeben
		for _, n := range networks {
			if _, ok := pet.FoundAPs[n.BSSID]; !ok {
				points := 1 * multiplier
				pet.Points += points
				pet.FoundAPs[n.BSSID] = FoundAP{SSID: n.SSID, Security: n.Security, Time: time.Now()}
				fmt.Printf("[+] Found AP: %s (%s) (+%d exp)\n", n.SSID, n.BSSID, points)
			}
		}
	}

	// Level up
	var levelUpPoints int
	switch {
	case pet.Level <= 9:
		levelUpPoints = 5
	case pet.Level <= 19:
		levelUpPoints = 10
	case pet.Level <= 29:
		levelUpPoints = 20
	default:
		levelUpPoints = 30
	}
	for pet.Points >= levelUpPoints {
		pet.Level++
		pet.Points -= levelUpPoints
		fmt.Printf("[+] Leveled up! Now Level %d\n", pet.Level)
	}

	savePet()
}

// Display only Level & Points at top
func displayPetStatus() {
	fmt.Printf("Level: %d | Exp: %d\n", pet.Level, pet.Points)
}
