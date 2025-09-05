[![Golang](https://img.shields.io/static/v1?label=Made%20with&message=Go&logo=go&color=007ACC)](https://go.dev/)
[![Go version](https://img.shields.io/github/go-mod/go-version/knuspii/crunchycleaner)](https://github.com/knuspii/crunchycleaner)
[![Go Report Card](https://goreportcard.com/report/github.com/knuspii/crunchycleaner)](https://goreportcard.com/report/github.com/knuspii/crunchycleaner)
[![GitHub Issues](https://img.shields.io/github/issues/knuspii/crunchycleaner)](https://github.com/knuspii/crunchycleaner/issues)
[![GitHub Stars](https://img.shields.io/github/stars/knuspii/crunchycleaner?style=social)](https://github.com/knuspii/crunchycleaner/stargazers)

<h1>CrunchyCleaner 🧹</h1>

<p align="center">
<pre>
  ____________________     .-.
 |  |              |  |    |_|
 |[]|              |[]|    | |
 |  |              |  |    |=|
 |  |              |  |  .=/I\=.
 |  |              |  | ////V\\\\
 |  |______________|  | |#######|
 |                    | |||||||||
 |     ____________   |
 |    | __      |  |  |
 |    ||  |     |  |  |
 |    ||__|     |  |  | CrunchyCleaner - Cleanup your system!
 |____|_________|__|__| Made by: Knuspii, (M)
</pre>
</p>

### ✨ A lightweight, cross-platform system cleanup tool
CrunchyCleaner is made to be simple, easy and *very crunchy indeed!*\
It helps you clear out caches, temp files, logs, and more — without confusing menus and 100+ options.


## 📥 [[Download here]](https://github.com/Knuspii/crunchycleaner/releases) <- Click here to download CrunchyCleaner!

---

## 🔑 Key features:

- 💻 **Cross-Platform**: Works on both **Windows** and **Linux**
- ⚡ **Lightweight**: Single binary, no dependencies (just download and run it)
- 🎨 **TUI (Text-UI)**: Simple, minimalist interface, no confusing menus
- 🧹 **Multiple Modes**:
  - Safe Clean (harmless cache cleanup)
  - Full Clean (deep cleanup of system junk)
  - User Clean (profile-specific cleanup)

---

## 📥 Download and install as command:
- **You need root/admin privileges!**
- Open your terminal and input this command, based on your operating system.

**Linux:**
```
sudo curl -L https://github.com/Knuspii/crunchycleaner/releases/latest/download/crunchycleaner -o /usr/local/bin/crunchycleaner && sudo chmod +x /usr/local/bin/crunchycleaner && echo "CrunchyCleaner installed at /usr/local/bin/crunchycleaner. Restart terminal to use 'crunchycleaner'"
```
**Windows (Powershell as admin):**
```
$ip="C:\Program Files\CrunchyCleaner"; if(-not (Test-Path $ip)){New-Item -ItemType Directory -Path $ip -Force | Out-Null}; $ep=Join-Path $ip "crunchycleaner.exe"; Invoke-WebRequest "https://github.com/Knuspii/crunchycleaner/releases/latest/download/crunchycleaner.exe" -OutFile $ep -UseBasicParsing; $envp=[System.Environment]::GetEnvironmentVariable("Path",[System.EnvironmentVariableTarget]::Machine); if($envp -notlike "*$ip*"){[System.Environment]::SetEnvironmentVariable("Path","$envp;$ip",[System.EnvironmentVariableTarget]::Machine)}; Write-Host "CrunchyCleaner installed at $ep. Restart terminal to use 'crunchycleaner'"
```
- **Now restart your terminal or reboot.**
- After that just type "crunchycleaner --version" into your terminal and it should output the current version.

---

## ⚙️ Start options:
```
Usage:
  crunchycleaner [option]

Options:
  No option     Run with TUI (Text-UI)
  -t            Run with TUI (Text-UI)
  -s            Run Safe-Cleanup
  -sy           Run Safe-Cleanup (non-interactive for scripts)
  -f            Run Full-Cleanup
  -fy           Run Full-Cleanup (non-interactive for scripts)
  -u [<user>]}  Run User-Cleanup
  -uy [<user>]  Run User-Cleanup (non-interactive for scripts)
  -v            Show version
  -h            Show this help page
