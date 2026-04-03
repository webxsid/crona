package app

import (
	"context"
	"errors"

	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/scratchfile"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func (h *Handler) handleRuntimeMethods(ctx context.Context, req protocol.Request) (protocol.Response, bool) {
	switch req.Method {
	case protocol.MethodContextGet:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.GetActiveContext(ctx, h.core)
		}), true
	case protocol.MethodContextSet:
		return h.handleNoParams(req, func() (any, error) {
			raw, err := decodeObject(req.Params)
			if err != nil {
				return nil, err
			}
			repoID, repoSet, err := decodeOptionalInt64FromMap(raw, "repoId")
			if err != nil {
				return nil, err
			}
			streamID, streamSet, err := decodeOptionalInt64FromMap(raw, "streamId")
			if err != nil {
				return nil, err
			}
			issueID, issueSet, err := decodeOptionalInt64FromMap(raw, "issueId")
			if err != nil {
				return nil, err
			}
			return corecommands.SetContext(ctx, h.core, corecommands.ContextPatch{
				RepoSet:   repoSet,
				RepoID:    repoID,
				StreamSet: streamSet,
				StreamID:  streamID,
				IssueSet:  issueSet,
				IssueID:   issueID,
			})
		}), true
	case protocol.MethodContextSwitchRepo:
		return handle(req, func(input shareddto.SwitchRepoRequest) (any, error) {
			return corecommands.SwitchRepo(ctx, h.core, input.RepoID)
		}), true
	case protocol.MethodContextSwitchStream:
		return handle(req, func(input shareddto.SwitchStreamRequest) (any, error) {
			return corecommands.SwitchStream(ctx, h.core, input.StreamID)
		}), true
	case protocol.MethodContextSwitchIssue:
		return handle(req, func(input shareddto.SwitchIssueRequest) (any, error) {
			return corecommands.SwitchIssue(ctx, h.core, input.IssueID)
		}), true
	case protocol.MethodContextClearIssue:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ClearIssue(ctx, h.core)
		}), true
	case protocol.MethodContextClear:
		return h.handleNoParams(req, func() (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.ClearContext(ctx, h.core)
		}), true
	case protocol.MethodSessionListByIssue:
		return handle(req, func(input shareddto.ListSessionsQuery) (any, error) {
			if input.IssueID == nil || *input.IssueID == 0 {
				return nil, errors.New("issueId is required")
			}
			return h.core.Sessions.ListByIssue(ctx, *input.IssueID, h.core.UserID)
		}), true
	case protocol.MethodSessionGet:
		return handle(req, func(input shareddto.SessionIDRequest) (any, error) {
			return h.core.Sessions.GetByID(ctx, input.ID, h.core.UserID)
		}), true
	case protocol.MethodSessionDetail:
		return handle(req, func(input shareddto.SessionIDRequest) (any, error) {
			return corecommands.GetSessionDetail(ctx, h.core, input.ID)
		}), true
	case protocol.MethodSessionGetActive:
		return h.handleNoParams(req, func() (any, error) {
			return h.core.Sessions.GetActiveSession(ctx, h.core.UserID)
		}), true
	case protocol.MethodSessionStart:
		return handle(req, func(input shareddto.StartSessionRequest) (any, error) {
			return corecommands.StartSession(ctx, h.core, input.IssueID)
		}), true
	case protocol.MethodSessionPause:
		return h.handleNoParams(req, func() (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.PauseSession(ctx, h.core, sharedtypes.SessionSegmentRest)
		}), true
	case protocol.MethodSessionResume:
		return h.handleNoParams(req, func() (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.ResumeSession(ctx, h.core)
		}), true
	case protocol.MethodSessionEnd:
		return handle(req, func(input shareddto.EndSessionRequest) (any, error) {
			return corecommands.StopSession(ctx, h.core, corecommands.SessionEndInput{
				CommitMessage: input.CommitMessage,
				WorkedOn:      input.WorkedOn,
				Outcome:       input.Outcome,
				NextStep:      input.NextStep,
				Blockers:      input.Blockers,
				Links:         input.Links,
			})
		}), true
	case protocol.MethodSessionLogManual:
		return handle(req, func(input shareddto.ManualSessionLogRequest) (any, error) {
			return corecommands.LogManualSession(ctx, h.core, corecommands.ManualSessionInput{
				IssueID:              input.IssueID,
				Date:                 input.Date,
				WorkDurationSeconds:  input.WorkDurationSeconds,
				BreakDurationSeconds: input.BreakDurationSeconds,
				StartTime:            input.StartTime,
				EndTime:              input.EndTime,
				CommitMessage:        input.CommitMessage,
				Notes:                input.Notes,
			})
		}), true
	case protocol.MethodSessionAmendNote:
		return handle(req, func(input shareddto.AmendSessionNoteRequest) (any, error) {
			return corecommands.AmendSessionNotes(ctx, h.core, input.Note, input.ID)
		}), true
	case protocol.MethodSessionHistory:
		return handle(req, func(input shareddto.SessionHistoryQuery) (any, error) {
			useContext := input.Context != nil && *input.Context
			return corecommands.ListSessionHistory(ctx, h.core, struct {
				RepoID   *int64
				StreamID *int64
				IssueID  *int64
				Since    *string
				Until    *string
				Limit    *int
				Offset   *int
			}{
				RepoID:   input.RepoID,
				StreamID: input.StreamID,
				IssueID:  input.IssueID,
				Since:    input.Since,
				Until:    input.Until,
				Limit:    input.Limit,
				Offset:   input.Offset,
			}, useContext)
		}), true
	case protocol.MethodTimerGetState:
		return h.handleNoParams(req, func() (any, error) {
			return h.timer.GetState(ctx)
		}), true
	case protocol.MethodTimerStart:
		return handle(req, func(input shareddto.TimerStartRequest) (any, error) {
			return h.timer.Start(ctx, input.IssueID)
		}), true
	case protocol.MethodTimerPause:
		return h.handleNoParams(req, func() (any, error) {
			return h.timer.Pause(ctx)
		}), true
	case protocol.MethodTimerResume:
		return h.handleNoParams(req, func() (any, error) {
			return h.timer.Resume(ctx)
		}), true
	case protocol.MethodTimerEnd:
		return handle(req, func(input shareddto.EndSessionRequest) (any, error) {
			return h.timer.End(ctx, corecommands.SessionEndInput{
				CommitMessage: input.CommitMessage,
				WorkedOn:      input.WorkedOn,
				Outcome:       input.Outcome,
				NextStep:      input.NextStep,
				Blockers:      input.Blockers,
				Links:         input.Links,
			})
		}), true
	case protocol.MethodStashList:
		return h.handleNoParams(req, func() (any, error) {
			return corecommands.ListStashes(ctx, h.core)
		}), true
	case protocol.MethodStashGet:
		return handle(req, func(input shareddto.StashIDRequest) (any, error) {
			return corecommands.GetStash(ctx, h.core, input.ID)
		}), true
	case protocol.MethodStashPush:
		return handle(req, func(input shareddto.CreateStashRequest) (any, error) {
			return corecommands.StashPush(ctx, h.core, input.StashNote)
		}), true
	case protocol.MethodStashApply:
		return handle(req, func(input shareddto.StashIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.StashPop(ctx, h.core, h.timer, input.ID)
		}), true
	case protocol.MethodStashDrop:
		return handle(req, func(input shareddto.StashIDRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.StashDrop(ctx, h.core, input.ID)
		}), true
	case protocol.MethodScratchpadList:
		return handle(req, func(input shareddto.ListScratchpadsQuery) (any, error) {
			return corecommands.ListScratchpads(ctx, h.core, input.PinnedOnly != nil && *input.PinnedOnly)
		}), true
	case protocol.MethodScratchpadRegister:
		return handle(req, func(input shareddto.RegisterScratchpadRequest) (any, error) {
			pinned := false
			if input.Pinned != nil {
				pinned = *input.Pinned
			}
			lastOpenedAt := ""
			if input.LastOpenedAt != nil {
				lastOpenedAt = *input.LastOpenedAt
			}
			id := ""
			if input.ID != nil {
				id = *input.ID
			}
			filePath, err := corecommands.RegisterScratchpad(ctx, h.core, sharedtypes.ScratchPadMeta{
				ID:           id,
				Name:         input.Name,
				Path:         input.Path,
				Pinned:       pinned,
				LastOpenedAt: lastOpenedAt,
			})
			if err != nil {
				return nil, err
			}
			if _, err := scratchfile.Create(h.core.ScratchDir, filePath, input.Name); err != nil {
				_ = corecommands.RemoveScratchpad(ctx, h.core, id)
				return nil, err
			}
			return map[string]any{"ok": true, "filePath": filePath}, nil
		}), true
	case protocol.MethodScratchpadGetMeta:
		return handle(req, func(input shareddto.ScratchpadIDRequest) (any, error) {
			meta, err := corecommands.GetScratchpad(ctx, h.core, input.ID)
			if err != nil {
				return nil, err
			}
			if meta == nil {
				return shareddto.ErrorResponse{OK: false, Error: "Scratchpad not found"}, nil
			}
			return map[string]any{"ok": true, "meta": meta}, nil
		}), true
	case protocol.MethodScratchpadRead:
		return handle(req, func(input shareddto.ScratchpadIDRequest) (any, error) {
			meta, err := corecommands.GetScratchpad(ctx, h.core, input.ID)
			if err != nil {
				return nil, err
			}
			if meta == nil {
				return sharedtypes.ScratchPadRead{OK: false, Error: ptrTo("Scratchpad not found")}, nil
			}
			content, err := scratchfile.Read(h.core.ScratchDir, meta.Path)
			if err != nil {
				return nil, err
			}
			return sharedtypes.ScratchPadRead{OK: true, Meta: meta, Content: &content}, nil
		}), true
	case protocol.MethodScratchpadPin:
		return handle(req, func(input shareddto.PinScratchpadRequest) (any, error) {
			return shareddto.OKResponse{OK: true}, corecommands.PinScratchpad(ctx, h.core, input.ID, input.Pinned)
		}), true
	case protocol.MethodScratchpadDelete:
		return handle(req, func(input shareddto.ScratchpadIDRequest) (any, error) {
			meta, err := corecommands.GetScratchpad(ctx, h.core, input.ID)
			if err != nil {
				return nil, err
			}
			if err := corecommands.RemoveScratchpad(ctx, h.core, input.ID); err != nil {
				return nil, err
			}
			if meta != nil {
				if err := scratchfile.Delete(h.core.ScratchDir, meta.Path); err != nil {
					return nil, err
				}
			}
			return shareddto.OKResponse{OK: true}, nil
		}), true
	case protocol.MethodSettingsGetAll:
		return h.handleNoParams(req, func() (any, error) {
			return h.core.CoreSettings.GetAllSettings(ctx)
		}), true
	case protocol.MethodSettingsGet:
		return handle(req, func(input shareddto.GetCoreSettingRequest) (any, error) {
			return h.core.CoreSettings.GetSetting(ctx, h.core.UserID, input.Key)
		}), true
	case protocol.MethodSettingsPatch:
		return handle(req, func(input shareddto.PatchCoreSettingRequest) (any, error) {
			if err := h.core.CoreSettings.SetSetting(ctx, h.core.UserID, input.Key, input.Value); err != nil {
				return nil, err
			}
			return shareddto.OKResponse{OK: true}, nil
		}), true
	case protocol.MethodSettingsPut:
		return handle(req, func(input shareddto.PutCoreSettingsRequest) (any, error) {
			updated := map[string]any{}
			for key, value := range input {
				if err := h.core.CoreSettings.SetSetting(ctx, h.core.UserID, key, value); err != nil {
					return nil, err
				}
				updated[string(key)] = value
			}
			return updated, nil
		}), true
	case protocol.MethodOpsLatest:
		return handle(req, func(input shareddto.ListLatestOpsQuery) (any, error) {
			limit := 50
			if input.Limit != nil {
				limit = *input.Limit
			}
			return corecommands.ListLatestOps(ctx, h.core, limit)
		}), true
	case protocol.MethodOpsSince:
		return handle(req, func(input shareddto.ListOpsSinceQuery) (any, error) {
			return corecommands.ListOpsSince(ctx, h.core, input.Since)
		}), true
	case protocol.MethodOpsList:
		return handle(req, func(input shareddto.ListOpsQuery) (any, error) {
			if input.Entity == nil || input.EntityID == nil {
				return nil, errors.New("entity and entityId are required")
			}
			limit := 100
			if input.Limit != nil {
				limit = *input.Limit
			}
			return corecommands.ListOpsByEntity(ctx, h.core, *input.Entity, *input.EntityID, limit)
		}), true
	default:
		return protocol.Response{}, false
	}
}
