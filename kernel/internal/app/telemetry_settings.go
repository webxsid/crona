package app

import sharedtypes "crona/shared/types"

func telemetryUsageEnabled(settings *sharedtypes.CoreSettings) bool {
	return settings != nil && settings.UsageTelemetryEnabled
}

func telemetryErrorReportingEnabled(settings *sharedtypes.CoreSettings) bool {
	return settings != nil && settings.ErrorReportingEnabled
}
