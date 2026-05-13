package main

import sharedtypes "crona/shared/types"

func telemetryOnboardingCompleted(settings *sharedtypes.CoreSettings) bool {
	return settings != nil && settings.OnboardingCompleted
}

func telemetryUsageEnabled(settings *sharedtypes.CoreSettings) bool {
	return telemetryOnboardingCompleted(settings) && settings.UsageTelemetryEnabled
}

func telemetryErrorReportingEnabled(settings *sharedtypes.CoreSettings) bool {
	return telemetryOnboardingCompleted(settings) && settings.ErrorReportingEnabled
}
