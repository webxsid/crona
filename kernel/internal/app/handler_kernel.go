package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"crona/shared/config"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
	versionpkg "crona/shared/version"
)

func (h *Handler) handleKernelMethods(ctx context.Context, req protocol.Request) (protocol.Response, bool) {
	switch req.Method {
	case protocol.MethodHealthGet:
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		dbOK := h.pingDB == nil || h.pingDB(ctx) == nil
		return mustResult(req.ID, sharedtypes.Health{
			Status: map[bool]string{true: "ok", false: "degraded"}[dbOK],
			DB:     dbOK,
			OK:     map[bool]int{true: 1, false: 0}[dbOK],
			Uptime: time.Since(parseStartedAt(h.startedAt)).Seconds(),
		}), true
	case protocol.MethodKernelInfoGet:
		return mustResult(req.ID, h.info), true
	case protocol.MethodKernelShutdown:
		if h.shutdown != nil {
			h.shutdown()
		}
		return mustResult(req.ID, shareddto.OKResponse{OK: true}), true
	case protocol.MethodKernelRestart:
		if h.shutdown != nil {
			h.shutdown()
		}
		return mustResult(req.ID, shareddto.OKResponse{OK: true}), true
	case protocol.MethodKernelSeedDev:
		return h.handleNoParams(req, func() (any, error) {
			if !strings.EqualFold(h.envMode, config.ModeDev) {
				return nil, errors.New("kernel.dev.seed is only available in Dev mode")
			}
			return shareddto.OKResponse{OK: true}, h.seedDevData(ctx)
		}), true
	case protocol.MethodKernelClearDev:
		return h.handleNoParams(req, func() (any, error) {
			if !strings.EqualFold(h.envMode, config.ModeDev) {
				return nil, errors.New("kernel.dev.clear is only available in Dev mode")
			}
			return shareddto.OKResponse{OK: true}, h.clearDevData(ctx)
		}), true
	case protocol.MethodKernelPrepareLocalUpdate:
		return h.handleNoParams(req, func() (any, error) {
			if !strings.EqualFold(h.envMode, config.ModeDev) {
				return nil, errors.New("kernel.dev.prepare_local_update is only available in Dev mode")
			}
			if h.updater == nil {
				return nil, errors.New("update service is unavailable")
			}
			return h.updater.PrepareLocalRelease(ctx)
		}), true
	case protocol.MethodKernelWipeData:
		return handle(req, func(input shareddto.ConfirmDangerousActionRequest) (any, error) {
			if !input.Confirm {
				return nil, errors.New("kernel.data.wipe requires explicit confirmation")
			}
			return shareddto.OKResponse{OK: true}, h.wipeRuntimeData(ctx)
		}), true
	case protocol.MethodUpdateStatusGet:
		return h.handleNoParams(req, func() (any, error) {
			if h.updater == nil {
				return sharedtypes.UpdateStatus{
					CurrentVersion: "unknown",
					Enabled:        false,
					RunningChannel: versionpkg.RunningChannel(),
					RunningIsBeta:  versionpkg.IsBetaRelease(),
				}, nil
			}
			return h.updater.Status(), nil
		}), true
	case protocol.MethodUpdateCheck:
		return h.handleNoParams(req, func() (any, error) {
			if h.updater == nil {
				return sharedtypes.UpdateStatus{
					CurrentVersion: "unknown",
					Enabled:        false,
					RunningChannel: versionpkg.RunningChannel(),
					RunningIsBeta:  versionpkg.IsBetaRelease(),
				}, nil
			}
			return h.updater.CheckNow(ctx)
		}), true
	case protocol.MethodUpdateDismiss:
		return h.handleNoParams(req, func() (any, error) {
			if h.updater == nil {
				return sharedtypes.UpdateStatus{
					CurrentVersion: "unknown",
					Enabled:        false,
					RunningChannel: versionpkg.RunningChannel(),
					RunningIsBeta:  versionpkg.IsBetaRelease(),
				}, nil
			}
			return h.updater.DismissLatest()
		}), true
	case protocol.MethodAlertsStatusGet:
		return h.handleNoParams(req, func() (any, error) {
			if h.alerts == nil {
				return sharedtypes.AlertStatus{AvailableSoundPresets: sharedtypes.AvailableAlertSoundPresets()}, nil
			}
			return h.alerts.Status(), nil
		}), true
	case protocol.MethodAlertsTestNotification:
		return h.handleNoParams(req, func() (any, error) {
			if h.alerts == nil {
				return nil, errors.New("alerts service is unavailable")
			}
			return shareddto.OKResponse{OK: true}, h.alerts.TestNotification(ctx)
		}), true
	case protocol.MethodAlertsTestSound:
		return h.handleNoParams(req, func() (any, error) {
			if h.alerts == nil {
				return nil, errors.New("alerts service is unavailable")
			}
			return shareddto.OKResponse{OK: true}, h.alerts.TestSound(ctx)
		}), true
	case protocol.MethodAlertsNotify:
		return handle(req, func(input sharedtypes.AlertRequest) (any, error) {
			if h.alerts == nil {
				return nil, errors.New("alerts service is unavailable")
			}
			return shareddto.OKResponse{OK: true}, h.alerts.Notify(ctx, input)
		}), true
	case protocol.MethodAlertsRemindersList:
		return h.handleNoParams(req, func() (any, error) {
			if h.alerts == nil {
				return []sharedtypes.AlertReminder{}, nil
			}
			return h.alerts.ListReminders(ctx)
		}), true
	case protocol.MethodAlertsRemindersCreate:
		return handle(req, func(input shareddto.AlertReminderCreateRequest) (any, error) {
			if h.alerts == nil {
				return nil, errors.New("alerts service is unavailable")
			}
			return h.alerts.CreateReminder(ctx, input)
		}), true
	case protocol.MethodAlertsRemindersUpdate:
		return handle(req, func(input shareddto.AlertReminderUpdateRequest) (any, error) {
			if h.alerts == nil {
				return nil, errors.New("alerts service is unavailable")
			}
			return h.alerts.UpdateReminder(ctx, input)
		}), true
	case protocol.MethodAlertsRemindersDelete:
		return handle(req, func(input shareddto.AlertReminderIDRequest) (any, error) {
			if h.alerts == nil {
				return nil, errors.New("alerts service is unavailable")
			}
			return shareddto.OKResponse{OK: true}, h.alerts.DeleteReminder(ctx, input.ID)
		}), true
	case protocol.MethodAlertsRemindersToggle:
		return handle(req, func(input shareddto.AlertReminderToggleRequest) (any, error) {
			if h.alerts == nil {
				return nil, errors.New("alerts service is unavailable")
			}
			return h.alerts.ToggleReminder(ctx, input.ID, input.Enabled)
		}), true
	default:
		return protocol.Response{}, false
	}
}
