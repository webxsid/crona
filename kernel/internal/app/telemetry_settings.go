package app

import sharedtypes "crona/shared/types"

func telemetryUsageEnabled(settings *sharedtypes.CoreSettings) bool {
	return settings != nil && settings.OnboardingCompleted && settings.UsageTelemetryEnabled
}

func telemetryErrorReportingEnabled(settings *sharedtypes.CoreSettings) bool {
	return settings != nil && settings.OnboardingCompleted && settings.ErrorReportingEnabled
}
