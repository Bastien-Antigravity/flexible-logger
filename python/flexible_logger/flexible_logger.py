import ctypes
import os
import platform
import sys
from typing import Optional

class FlexibleLogger:
    """ Python wrapper for the Go Flexible Logger library. """

    # Log Levels
    DEBUG = 0
    INFO = 1
    WARNING = 2
    ERROR = 3

    _lib = None

    def __init__(self, name: str, profile: str = "standard", config_profile: str = "standalone"):
        self.name = name
        self.profile = profile
        self.config_profile = config_profile
        self._handle = -1

        if FlexibleLogger._lib is None:
            FlexibleLogger._lib = self._load_library()

        self._create_logger()

    def _load_library(self):
        system = platform.system().lower()
        if system == "linux":
            lib_name = "libflexible_logger.so"
        elif system == "darwin":
            lib_name = "libflexible_logger.dylib"
        elif system == "windows":
            lib_name = "libflexible_logger.dll"
        else:
            raise RuntimeError(f"Unsupported platform: {system}")

        # Search in the package directory
        lib_path = os.path.join(os.path.dirname(__file__), lib_name)
        
        # Fallback to current directory for local development
        if not os.path.exists(lib_path):
            lib_path = lib_name

        try:
            lib = ctypes.CDLL(lib_path)
            
            # Configure function signatures
            lib.CreateLogger.argtypes = [ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
            lib.CreateLogger.restype = ctypes.c_int32
            
            lib.Log.argtypes = [ctypes.c_int32, ctypes.c_int32, ctypes.c_char_p]
            lib.Log.restype = ctypes.c_int32
            
            lib.CloseLogger.argtypes = [ctypes.c_int32]
            lib.CloseLogger.restype = None
            
            lib.GetLastError.argtypes = []
            lib.GetLastError.restype = ctypes.c_char_p
            
            return lib
        except Exception as e:
            raise RuntimeError(f"Failed to load shared library {lib_path}: {e}")

    def _create_logger(self):
        h = self._lib.CreateLogger(
            self.name.encode('utf-8'),
            self.profile.encode('utf-8'),
            self.config_profile.encode('utf-8')
        )
        if h < 0:
            err = self._lib.GetLastError()
            raise RuntimeError(f"Failed to create logger: {err.decode('utf-8') if err else 'unknown error'}")
        self._handle = h

    def log(self, level: int, message: str):
        if self._handle < 0:
            return
        
        res = self._lib.Log(self._handle, level, message.encode('utf-8'))
        if res < 0:
            err = self._lib.GetLastError()
            print(f"Logging error: {err.decode('utf-8') if err else 'unknown error'}", file=sys.stderr)

    def debug(self, message: str):
        self.log(self.DEBUG, message)

    def info(self, message: str):
        self.log(self.INFO, message)

    def warning(self, message: str):
        self.log(self.WARNING, message)

    def error(self, message: str):
        self.log(self.ERROR, message)

    def close(self):
        if self._handle >= 0:
            self._lib.CloseLogger(self._handle)
            self._handle = -1

    def __del__(self):
        self.close()
