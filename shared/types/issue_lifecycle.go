package types

func NormalizeIssueStatus(status IssueStatus) IssueStatus {
	switch status {
	case "todo":
		return IssueStatusBacklog
	case "active":
		return IssueStatusInProgress
	case IssueStatusBacklog,
		IssueStatusPlanned,
		IssueStatusReady,
		IssueStatusInProgress,
		IssueStatusBlocked,
		IssueStatusInReview,
		IssueStatusDone,
		IssueStatusAbandoned:
		return status
	default:
		return status
	}
}

func AllIssueStatuses() []IssueStatus {
	return []IssueStatus{
		IssueStatusBacklog,
		IssueStatusPlanned,
		IssueStatusReady,
		IssueStatusInProgress,
		IssueStatusBlocked,
		IssueStatusInReview,
		IssueStatusDone,
		IssueStatusAbandoned,
	}
}

func AllowedIssueStatusTransitions(status IssueStatus) []IssueStatus {
	switch NormalizeIssueStatus(status) {
	case IssueStatusBacklog:
		return []IssueStatus{IssueStatusPlanned, IssueStatusReady, IssueStatusAbandoned}
	case IssueStatusPlanned:
		return []IssueStatus{
			IssueStatusBacklog,
			IssueStatusReady,
			IssueStatusInProgress,
			IssueStatusBlocked,
			IssueStatusAbandoned,
		}
	case IssueStatusReady:
		return []IssueStatus{
			IssueStatusPlanned,
			IssueStatusInProgress,
			IssueStatusBlocked,
			IssueStatusAbandoned,
		}
	case IssueStatusInProgress:
		return []IssueStatus{
			IssueStatusPlanned,
			IssueStatusReady,
			IssueStatusBlocked,
			IssueStatusInReview,
			IssueStatusDone,
			IssueStatusAbandoned,
		}
	case IssueStatusBlocked:
		return []IssueStatus{IssueStatusPlanned, IssueStatusReady, IssueStatusAbandoned}
	case IssueStatusInReview:
		return []IssueStatus{
			IssueStatusPlanned,
			IssueStatusReady,
			IssueStatusInProgress,
			IssueStatusDone,
			IssueStatusAbandoned,
		}
	case IssueStatusDone:
		return []IssueStatus{IssueStatusPlanned}
	case IssueStatusAbandoned:
		return []IssueStatus{IssueStatusBacklog, IssueStatusPlanned}
	default:
		return nil
	}
}

func IsValidIssueTransition(from, to IssueStatus) bool {
	from = NormalizeIssueStatus(from)
	to = NormalizeIssueStatus(to)
	for _, candidate := range AllowedIssueStatusTransitions(from) {
		if candidate == to {
			return true
		}
	}
	return false
}

func CanStartFocus(status IssueStatus) bool {
	switch NormalizeIssueStatus(status) {
	case IssueStatusPlanned, IssueStatusReady, IssueStatusInProgress:
		return true
	default:
		return false
	}
}

func AutoStatusOnFocusStart(status IssueStatus) IssueStatus {
	switch NormalizeIssueStatus(status) {
	case IssueStatusPlanned, IssueStatusReady:
		return IssueStatusInProgress
	default:
		return NormalizeIssueStatus(status)
	}
}

func AutoStatusOnTodoAssigned(status IssueStatus) IssueStatus {
	switch NormalizeIssueStatus(status) {
	case IssueStatusBacklog:
		return IssueStatusPlanned
	default:
		return NormalizeIssueStatus(status)
	}
}
