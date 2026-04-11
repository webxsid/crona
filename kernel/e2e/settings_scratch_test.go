//go:build e2e

package e2e

import (
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestSettingsAndScratchpadOverIPC(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	var ok shareddto.OKResponse
	kernel.call(t, protocol.MethodSettingsPatch, shareddto.PatchCoreSettingRequest{
		Key:   sharedtypes.CoreSettingsKeyBreaksEnabled,
		Value: false,
	}, &ok)
	if !ok.OK {
		t.Fatalf("expected settings patch ok")
	}

	var breaksEnabled bool
	kernel.call(t, protocol.MethodSettingsGet, shareddto.GetCoreSettingRequest{
		Key: sharedtypes.CoreSettingsKeyBreaksEnabled,
	}, &breaksEnabled)
	if breaksEnabled {
		t.Fatalf("expected breaksEnabled=false after patch")
	}

	var registered struct {
		OK       bool   `json:"ok"`
		FilePath string `json:"filePath"`
	}
	kernel.call(t, protocol.MethodScratchpadRegister, shareddto.RegisterScratchpadRequest{
		ID:   stringPtr("scratch-1"),
		Name: "Notes",
		Path: "notes/today",
	}, &registered)
	if !registered.OK || registered.FilePath == "" {
		t.Fatalf("unexpected scratchpad register response: %+v", registered)
	}

	var read sharedtypes.ScratchPadRead
	kernel.call(t, protocol.MethodScratchpadRead, shareddto.ScratchpadIDRequest{ID: "scratch-1"}, &read)
	if !read.OK || read.Meta == nil || read.Meta.Name != "Notes" {
		t.Fatalf("unexpected scratchpad read response: %+v", read)
	}

	kernel.call(t, protocol.MethodScratchpadDelete, shareddto.ScratchpadIDRequest{ID: "scratch-1"}, &ok)
	if !ok.OK {
		t.Fatalf("expected scratchpad delete ok")
	}

	var pads []sharedtypes.ScratchPadMeta
	kernel.call(t, protocol.MethodScratchpadList, shareddto.ListScratchpadsQuery{}, &pads)
	if len(pads) != 0 {
		t.Fatalf("expected scratchpads to be empty after delete, got %+v", pads)
	}
}
