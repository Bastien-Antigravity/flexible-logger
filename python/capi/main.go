package main

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"
)

var (
	registry    sync.Map
	nextHandle  int32
	lastError   [1024]byte
	registryMut sync.Mutex
)

// setError sets the global last error string for C access.
func setError(err error) {
	if err != nil {
		s := err.Error()
		if len(s) > 1023 {
			s = s[:1023]
		}
		copy(lastError[:], s)
		lastError[len(s)] = 0
	} else {
		lastError[0] = 0
	}
}

//export GetLastError
func GetLastError() *C.char {
	return (*C.char)(unsafe.Pointer(&lastError[0]))
}

//export CreateLogger
func CreateLogger(name *C.char, profile *C.char, configProfile *C.char) int32 {
	n := C.GoString(name)
	p := C.GoString(profile)
	cp := C.GoString(configProfile)

	// 1. Create Config
	distConf := distributed_config.New(cp)
	if distConf == nil {
		setError(fmt.Errorf("failed to create config for profile: %s", cp))
		return -1
	}

	// 2. Create Logger based on profile
	var logger interfaces.Logger
	switch p {
	case "high-perf":
		logger = profiles.NewHighPerfLogger(n, distConf)
	case "standard":
		logger = profiles.NewStandardLogger(n, distConf)
	case "no-lock":
		logger = profiles.NewNoLockLogger(n, distConf)
	case "devel":
		logger = profiles.NewDevelLogger(n)
	case "minimal":
		logger = profiles.NewMinimalLogger(n)
	case "notif":
		logger = profiles.NewNotifLogger(n, distConf)
	default:
		setError(fmt.Errorf("invalid logger profile: %s", p))
		return -1
	}

	if logger == nil {
		setError(fmt.Errorf("failed to create logger for profile: %s", p))
		return -1
	}

	registryMut.Lock()
	defer registryMut.Unlock()
	nextHandle++
	registry.Store(nextHandle, logger)
	return nextHandle
}

//export Log
func Log(handle int32, level int32, message *C.char) int32 {
	val, ok := registry.Load(handle)
	if !ok {
		setError(fmt.Errorf("invalid handle: %d", handle))
		return -1
	}

	logger, ok := val.(interfaces.Logger)
	if !ok {
		setError(fmt.Errorf("handle %d is not a valid Logger instance", handle))
		return -1
	}

	msg := C.GoString(message)

	logger.Log(models.Level(level), msg)

	return 0
}

//export CloseLogger
func CloseLogger(handle int32) {
	val, ok := registry.Load(handle)
	if ok {
		if logger, ok := val.(interfaces.Logger); ok {
			logger.Close()
		}
		registry.Delete(handle)
	}
}

//export FreeString
func FreeString(s *C.char) {
	C.free(unsafe.Pointer(s))
}

func main() {}
