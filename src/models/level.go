package models

import "strings"

// -----------------------------------------------------------------------------
// Level defines log levels
type Level uint8

// -----------------------------------------------------------------------------
const (
	LevelNotSet   Level = 0
	LevelDebug    Level = 1
	LevelStream   Level = 2
	LevelInfo     Level = 3
	LevelLogon    Level = 4
	LevelLogout   Level = 5
	LevelTrade    Level = 6
	LevelSchedule Level = 7
	LevelReport   Level = 8
	LevelWarning  Level = 9
	LevelError    Level = 10
	LevelCritical Level = 11
)

// -----------------------------------------------------------------------------
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelStream:
		return "STREAM"
	case LevelInfo:
		return "INFO"
	case LevelLogon:
		return "LOGON"
	case LevelLogout:
		return "LOGOUT"
	case LevelTrade:
		return "TRADE"
	case LevelSchedule:
		return "SCHEDULE"
	case LevelReport:
		return "REPORT"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelCritical:
		return "CRITICAL"
	default:
		return "INFO"
	}
}

// -----------------------------------------------------------------------------
func ParseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LevelDebug
	case "STREAM":
		return LevelStream
	case "INFO":
		return LevelInfo
	case "LOGON":
		return LevelLogon
	case "LOGOUT":
		return LevelLogout
	case "TRADE":
		return LevelTrade
	case "SCHEDULE":
		return LevelSchedule
	case "REPORT":
		return LevelReport
	case "WARNING", "WARN":
		return LevelWarning
	case "ERROR":
		return LevelError
	case "CRITICAL":
		return LevelCritical
	default:
		return LevelInfo
	}
}
