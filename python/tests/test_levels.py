import pytest
import flexible_logger
from flexible_logger import FlexibleLogger

def test_levels_exposed():
    """ Test that log levels are exposed at the package level. """
    assert hasattr(flexible_logger, "DEBUG")
    assert hasattr(flexible_logger, "INFO")
    assert hasattr(flexible_logger, "WARNING")
    assert hasattr(flexible_logger, "ERROR")
    assert hasattr(flexible_logger, "CRITICAL")
    assert hasattr(flexible_logger, "TRADE")
    
    assert flexible_logger.DEBUG == 1
    assert flexible_logger.INFO == 3
    assert flexible_logger.WARNING == 9
    assert flexible_logger.ERROR == 10
    assert flexible_logger.CRITICAL == 11
    assert flexible_logger.TRADE == 6

def test_convenience_methods():
    """ Test that all convenience methods work. """
    logger = FlexibleLogger(name="TestLevels", profile="minimal")
    
    # These should not raise exceptions
    logger.debug("debug")
    logger.stream("stream")
    logger.info("info")
    logger.logon("logon")
    logger.logout("logout")
    logger.trade("trade")
    logger.schedule("schedule")
    logger.report("report")
    logger.warning("warning")
    logger.error("error")
    logger.critical("critical")
    
    logger.close()

if __name__ == "__main__":
    pytest.main([__file__])
