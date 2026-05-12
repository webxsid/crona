package app

import (
	sharedposthog "crona/shared/posthog"
	sharedtypes "crona/shared/types"
)

func captureDaemonStarted(telemetry sharedposthog.Client, transport string, runningChannel sharedtypes.UpdateChannel) error {
	return telemetry.Capture(sharedposthog.EventDaemonStarted, daemonLifecycleProperties(transport, runningChannel))
}

func captureDaemonStopped(telemetry sharedposthog.Client, transport string, runningChannel sharedtypes.UpdateChannel) error {
	return telemetry.Capture(sharedposthog.EventDaemonStopped, daemonLifecycleProperties(transport, runningChannel))
}

func daemonLifecycleProperties(transport string, runningChannel sharedtypes.UpdateChannel) sharedposthog.Properties {
	return sharedposthog.Properties{
		"entrypoint":      "daemon",
		"transport":       transport,
		"running_channel": string(runningChannel),
	}
}
