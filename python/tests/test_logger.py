import pytest
import os
import time
from flexible_logger import FlexibleLogger

def test_logger_init():
    """ Test that the logger can be initialized. """
    logger = FlexibleLogger(name="TestApp", profile="minimal")
    assert logger._handle >= 0
    logger.close()

def test_logging_methods():
    """ Test that logging methods don't crash. """
    logger = FlexibleLogger(name="TestApp", profile="minimal")
    
    logger.debug("This is a debug message")
    logger.info("This is an info message")
    logger.warning("This is a warning message")
    logger.error("This is an error message")
    
    # Wait a bit for the async sink to process (though we can't easily verify the output here)
    time.sleep(0.1)
    
    logger.close()

def test_invalid_profile():
    """ Test that an invalid profile raises a RuntimeError. """
    with pytest.raises(RuntimeError, match="invalid logger profile"):
        FlexibleLogger(name="TestApp", profile="non-existent")

if __name__ == "__main__":
    pytest.main([__file__])
