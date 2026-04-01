from .flexible_logger import FlexibleLogger

# Log Levels
NOT_SET = FlexibleLogger.NOT_SET
DEBUG = FlexibleLogger.DEBUG
STREAM = FlexibleLogger.STREAM
INFO = FlexibleLogger.INFO
LOGON = FlexibleLogger.LOGON
LOGOUT = FlexibleLogger.LOGOUT
TRADE = FlexibleLogger.TRADE
SCHEDULE = FlexibleLogger.SCHEDULE
REPORT = FlexibleLogger.REPORT
WARNING = FlexibleLogger.WARNING
ERROR = FlexibleLogger.ERROR
CRITICAL = FlexibleLogger.CRITICAL

__all__ = [
    'FlexibleLogger',
    'NOT_SET',
    'DEBUG',
    'STREAM',
    'INFO',
    'LOGON',
    'LOGOUT',
    'TRADE',
    'SCHEDULE',
    'REPORT',
    'WARNING',
    'ERROR',
    'CRITICAL'
]
