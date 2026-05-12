package main

import sharedposthog "crona/shared/posthog"

func captureTUIStarted(telemetry sharedposthog.Client) error {
	return telemetry.Capture(sharedposthog.EventTUIStarted, sharedposthog.Properties{
		"entrypoint": "tui",
	})
}
