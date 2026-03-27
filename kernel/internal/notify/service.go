package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"crona/kernel/internal/core"
	"crona/kernel/internal/events"
	runtimepkg "crona/kernel/internal/runtime"
	sharedtypes "crona/shared/types"
)

type Service struct {
	core   *core.Context
	bus    *events.Bus
	logger *runtimepkg.Logger
	queue  chan sharedtypes.TimerBoundaryPayload
}

var (
	sendNotificationFn = sendNotification
	playSoundFn        = playSound
)

func Start(ctx context.Context, coreCtx *core.Context, bus *events.Bus, logger *runtimepkg.Logger) *Service {
	service := &Service{
		core:   coreCtx,
		bus:    bus,
		logger: logger,
		queue:  make(chan sharedtypes.TimerBoundaryPayload, 16),
	}
	unsubscribe := bus.Subscribe(func(event sharedtypes.KernelEvent) {
		if event.Type != sharedtypes.EventTypeTimerBoundary {
			return
		}
		var payload sharedtypes.TimerBoundaryPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			logger.Error("decode timer boundary payload", err)
			return
		}
		select {
		case service.queue <- payload:
		default:
			logger.Error("drop timer boundary notification", fmt.Errorf("notification queue full"))
		}
	})
	go func() {
		defer unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case payload := <-service.queue:
				service.dispatch(ctx, payload)
			}
		}
	}()
	return service
}

func (s *Service) dispatch(ctx context.Context, payload sharedtypes.TimerBoundaryPayload) {
	settings, err := s.core.CoreSettings.Get(ctx, s.core.UserID)
	if err != nil {
		s.logger.Error("load notification settings", err)
		return
	}
	if settings == nil || (!settings.BoundaryNotifications && !settings.BoundarySound) {
		return
	}
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		title = "Timer boundary reached"
	}
	message := strings.TrimSpace(payload.Message)
	if message == "" {
		message = "Structured timer boundary reached"
	}
	if settings.BoundaryNotifications {
		if err := sendNotificationFn(title, message); err != nil {
			s.logger.Error("send boundary notification", err)
		}
	}
	if settings.BoundarySound {
		if err := playSoundFn(); err != nil {
			s.logger.Error("play boundary sound", err)
		}
	}
}

func sendNotification(title, message string) error {
	switch runtime.GOOS {
	case "darwin":
		return runCommand("osascript", "-e", fmt.Sprintf(`display notification %q with title %q`, message, title))
	case "linux":
		if _, err := exec.LookPath("notify-send"); err != nil {
			return err
		}
		return runCommand("notify-send", title, message)
	case "windows":
		script := fmt.Sprintf(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; [Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] > $null; $xml = New-Object Windows.Data.Xml.Dom.XmlDocument; $xml.LoadXml("<toast><visual><binding template='ToastGeneric'><text>%s</text><text>%s</text></binding></visual></toast>"); $toast = [Windows.UI.Notifications.ToastNotification]::new($xml); [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Crona").Show($toast)`, escapePowerShell(title), escapePowerShell(message))
		return runCommand("powershell", "-NoProfile", "-Command", script)
	default:
		return nil
	}
}

func playSound() error {
	switch runtime.GOOS {
	case "darwin":
		return runCommand("osascript", "-e", "beep")
	case "linux":
		if _, err := exec.LookPath("canberra-gtk-play"); err == nil {
			return runCommand("canberra-gtk-play", "--id", "bell")
		}
		return nil
	case "windows":
		return runCommand("powershell", "-NoProfile", "-Command", "[console]::beep(880,250)")
	default:
		return nil
	}
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func escapePowerShell(value string) string {
	return strings.ReplaceAll(value, `'`, `''`)
}
