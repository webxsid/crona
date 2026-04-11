package notify

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	runtimepkg "crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
)

func detectAlertStatus(paths runtimepkg.Paths) sharedtypes.AlertStatus {
	status := sharedtypes.AlertStatus{
		AvailableSoundPresets: AvailableSoundPresets(),
		IconPath:              alertIconPath(paths),
	}
	switch runtime.GOOS {
	case "darwin":
		status.NotificationOptions = []string{"terminal-notifier", "osascript"}
		status.SoundOptions = []string{"afplay"}
		if _, err := exec.LookPath("terminal-notifier"); err == nil {
			status.NotificationsAvailable = true
			status.NotificationBackend = "terminal-notifier"
			status.SubtitleSupported = true
			status.IconSupported = status.IconPath != ""
		} else {
			if _, err := exec.LookPath("osascript"); err == nil {
				status.NotificationsAvailable = true
				status.NotificationBackend = "osascript"
				status.SubtitleSupported = true
			}
		}
		if _, err := exec.LookPath("afplay"); err == nil {
			status.SoundAvailable = true
			status.SoundBackend = "afplay"
			status.BundledSoundSupported = true
		}
	case "linux":
		status.NotificationOptions = []string{"notify-send"}
		status.SoundOptions = []string{"paplay", "aplay", "play", "canberra-gtk-play"}
		if _, err := exec.LookPath("notify-send"); err == nil {
			status.NotificationsAvailable = true
			status.NotificationBackend = "notify-send"
			status.UrgencySupported = true
			status.IconSupported = status.IconPath != ""
		}
		switch {
		case commandAvailable("paplay"):
			status.SoundAvailable = true
			status.SoundBackend = "paplay"
			status.BundledSoundSupported = true
		case commandAvailable("aplay"):
			status.SoundAvailable = true
			status.SoundBackend = "aplay"
			status.BundledSoundSupported = true
		case commandAvailable("play"):
			status.SoundAvailable = true
			status.SoundBackend = "play"
			status.BundledSoundSupported = true
		case commandAvailable("canberra-gtk-play"):
			status.SoundAvailable = true
			status.SoundBackend = "canberra-gtk-play"
		}
	case "windows":
		status.NotificationOptions = []string{"burnttoast", "powershell_toast"}
		status.SoundOptions = []string{"powershell_soundplayer"}
		status.NotificationsAvailable = true
		if burntToastAvailable() {
			status.NotificationBackend = "burnttoast"
		} else {
			status.NotificationBackend = "powershell_toast"
		}
		status.SubtitleSupported = true
		status.IconSupported = status.IconPath != ""
		status.SoundAvailable = true
		status.SoundBackend = "powershell_soundplayer"
		status.BundledSoundSupported = true
	}
	return status
}

func AvailableSoundPresets() []sharedtypes.AlertSoundPreset {
	return sharedtypes.AvailableAlertSoundPresets()
}

func sendAlertNotification(status sharedtypes.AlertStatus, req sharedtypes.AlertRequest) error {
	switch runtime.GOOS {
	case "darwin":
		if status.NotificationBackend == "terminal-notifier" {
			args := []string{"-title", req.Title, "-message", req.Body}
			if req.Subtitle != "" {
				args = append(args, "-subtitle", req.Subtitle)
			}
			if req.IconEnabled && status.IconSupported && status.IconPath != "" {
				args = append(args, "-appIcon", fileURI(status.IconPath))
			}
			return runCommand("terminal-notifier", args...)
		}
		script := fmt.Sprintf(`display notification %q with title %q`, req.Body, req.Title)
		if req.Subtitle != "" {
			script += fmt.Sprintf(` subtitle %q`, req.Subtitle)
		}
		return runCommand("osascript", "-e", script)
	case "linux":
		if !status.NotificationsAvailable {
			return nil
		}
		body := req.Body
		if req.Subtitle != "" {
			body = req.Subtitle + "\n" + req.Body
		}
		args := []string{}
		if status.UrgencySupported {
			args = append(args, "-u", linuxUrgency(req.Urgency))
		}
		if req.IconEnabled && status.IconSupported && status.IconPath != "" {
			args = append(args, "-i", status.IconPath)
		}
		args = append(args, req.Title, body)
		return runCommand("notify-send", args...)
	case "windows":
		if status.NotificationBackend == "burnttoast" {
			lines := []string{req.Title}
			if strings.TrimSpace(req.Subtitle) != "" {
				lines = append(lines, req.Subtitle)
			}
			if strings.TrimSpace(req.Body) != "" {
				lines = append(lines, req.Body)
			}
			var script strings.Builder
			script.WriteString("Import-Module BurntToast -ErrorAction Stop; ")
			script.WriteString("$builder = New-BTContentBuilder; ")
			fmt.Fprintf(&script, "$builder = Add-BTText -ContentBuilder $builder -Text @(%s) -PassThru; ", powerShellStringArray(lines))
			if req.IconEnabled && status.IconSupported && status.IconPath != "" {
				fmt.Fprintf(&script, "$builder = Add-BTImage -ContentBuilder $builder -Source '%s' -AppLogoOverride -PassThru; ", escapePowerShell(filepath.Clean(status.IconPath)))
			}
			script.WriteString("Show-BTNotification -ContentBuilder $builder")
			return runCommand(powerShellExecutable(), "-NoProfile", "-Command", script.String())
		}
		textNodes := []string{
			fmt.Sprintf("<text>%s</text>", escapeXML(req.Title)),
		}
		if req.Subtitle != "" {
			textNodes = append(textNodes, fmt.Sprintf("<text>%s</text>", escapeXML(req.Subtitle)))
		}
		if req.Body != "" {
			textNodes = append(textNodes, fmt.Sprintf("<text>%s</text>", escapeXML(req.Body)))
		}
		logo := ""
		if req.IconEnabled && status.IconSupported && status.IconPath != "" {
			logo = fmt.Sprintf("<image placement='appLogoOverride' hint-crop='none' src='%s'/>", escapeXML(fileURI(status.IconPath)))
		}
		xml := "<toast><visual><binding template='ToastGeneric'>" + logo + strings.Join(textNodes, "") + "</binding></visual></toast>"
		script := fmt.Sprintf(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; [Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] > $null; $xml = New-Object Windows.Data.Xml.Dom.XmlDocument; $xml.LoadXml('%s'); $toast = [Windows.UI.Notifications.ToastNotification]::new($xml); [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Crona').Show($toast)`, escapePowerShell(xml))
		return runCommand(powerShellExecutable(), "-NoProfile", "-Command", script)
	default:
		return nil
	}
}

func playAlertSound(status sharedtypes.AlertStatus, soundPath string) error {
	switch runtime.GOOS {
	case "darwin":
		if status.SoundBackend == "afplay" {
			return runCommand("afplay", soundPath)
		}
	case "linux":
		switch status.SoundBackend {
		case "paplay":
			return runCommand("paplay", soundPath)
		case "aplay":
			return runCommand("aplay", soundPath)
		case "play":
			return runCommand("play", "-q", soundPath)
		case "canberra-gtk-play":
			return runCommand("canberra-gtk-play", "--id", "bell")
		}
	case "windows":
		script := fmt.Sprintf(`$player = New-Object System.Media.SoundPlayer '%s'; $player.PlaySync()`, escapePowerShell(filepath.Clean(soundPath)))
		return runCommand(powerShellExecutable(), "-NoProfile", "-Command", script)
	}
	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func burntToastAvailable() bool {
	powershell := powerShellExecutable()
	if powershell == "" {
		return false
	}
	cmd := exec.Command(powershell, "-NoProfile", "-Command", "Get-Module -ListAvailable -Name BurntToast | Select-Object -First 1 | ForEach-Object { $_.Name }")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(string(output))), "burnttoast")
}

func powerShellExecutable() string {
	if commandAvailable("pwsh") {
		return "pwsh"
	}
	if commandAvailable("powershell") {
		return "powershell"
	}
	return "powershell"
}

func linuxUrgency(value sharedtypes.AlertUrgency) string {
	switch sharedtypes.NormalizeAlertUrgency(value) {
	case sharedtypes.AlertUrgencyLow:
		return "low"
	case sharedtypes.AlertUrgencyHigh:
		return "critical"
	default:
		return "normal"
	}
}

func fileURI(path string) string {
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return "file:///" + filepath.ToSlash(path)
	}
	return "file://" + path
}

func escapePowerShell(value string) string {
	return strings.ReplaceAll(value, `'`, `''`)
}

func powerShellStringArray(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("'%s'", escapePowerShell(value)))
	}
	return strings.Join(quoted, ", ")
}

func escapeXML(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(value)
}
