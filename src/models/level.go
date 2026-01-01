package models

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
