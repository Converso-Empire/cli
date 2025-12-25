#!/usr/bin/env python3
"""
Converso CLI Python Bridge

This module provides the bridge between Go CLI and Python modules.
It handles JSON IPC communication and progress event streaming.
"""

import json
import sys
import os
import time
import signal
from typing import Dict, Any, Optional, Callable, Generator
from dataclasses import dataclass, asdict
from enum import Enum


class MessageType(Enum):
    """Message types for IPC communication"""
    REQUEST = "request"
    RESPONSE = "response"
    PROGRESS = "progress"
    ERROR = "error"


@dataclass
class ModuleRequest:
    """Request from Go CLI to Python module"""
    command: str
    args: Dict[str, Any]
    auth_token: str
    device_token: str
    timeout: int


@dataclass
class ModuleResponse:
    """Response from Python module to Go CLI"""
    success: bool
    data: Dict[str, Any]
    error: Optional[str] = None
    progress: Optional[Dict[str, Any]] = None


@dataclass
class ProgressEvent:
    """Progress event from Python module"""
    stage: str
    current: int
    total: int
    percentage: float
    message: str
    timestamp: float


class IPCBridge:
    """Handles JSON IPC communication between Go CLI and Python modules"""
    
    def __init__(self):
        self.auth_token = None
        self.device_token = None
        self.timeout = 300  # Default 5 minutes
        
    def read_request(self) -> ModuleRequest:
        """Read request from stdin"""
        try:
            line = sys.stdin.readline().strip()
            if not line:
                raise EOFError("No input received")
                
            data = json.loads(line)
            return ModuleRequest(
                command=data.get('command', ''),
                args=data.get('args', {}),
                auth_token=data.get('auth_token', ''),
                device_token=data.get('device_token', ''),
                timeout=data.get('timeout', 300)
            )
        except json.JSONDecodeError as e:
            self.send_error(f"Failed to parse JSON request: {e}")
            sys.exit(1)
        except Exception as e:
            self.send_error(f"Failed to read request: {e}")
            sys.exit(1)
    
    def send_response(self, response: ModuleResponse):
        """Send response to stdout"""
        try:
            output = json.dumps(asdict(response))
            sys.stdout.write(output + '\n')
            sys.stdout.flush()
        except Exception as e:
            self.send_error(f"Failed to send response: {e}")
            sys.exit(1)
    
    def send_progress(self, stage: str, current: int, total: int, message: str = ""):
        """Send progress event"""
        progress = ProgressEvent(
            stage=stage,
            current=current,
            total=total,
            percentage=(current / total * 100) if total > 0 else 0,
            message=message,
            timestamp=time.time()
        )
        
        response = ModuleResponse(
            success=True,
            data={},
            progress=asdict(progress)
        )
        
        self.send_response(response)
    
    def send_error(self, error: str):
        """Send error response"""
        response = ModuleResponse(
            success=False,
            data={},
            error=error
        )
        self.send_response(response)
    
    def validate_auth(self) -> bool:
        """Validate authentication tokens"""
        if not self.auth_token:
            self.send_error("Authentication required")
            return False
        
        # Basic token validation (in production, validate JWT signature)
        if not self.auth_token.startswith('Bearer '):
            self.send_error("Invalid token format")
            return False
        
        return True
    
    def handle_timeout(self, signum, frame):
        """Handle timeout signal"""
        self.send_error("Module execution timed out")
        sys.exit(1)
    
    def set_timeout(self, timeout_seconds: int):
        """Set execution timeout"""
        if timeout_seconds > 0:
            signal.signal(signal.SIGALRM, self.handle_timeout)
            signal.alarm(timeout_seconds)


class ModuleBase:
    """Base class for all Python modules"""
    
    def __init__(self):
        self.bridge = IPCBridge()
        self.commands = {}
    
    def register_command(self, name: str, handler: Callable):
        """Register a command handler"""
        self.commands[name] = handler
    
    def run(self):
        """Main execution loop"""
        try:
            # Read request
            request = self.bridge.read_request()
            
            # Set timeout
            self.bridge.set_timeout(request.timeout)
            
            # Store tokens
            self.bridge.auth_token = request.auth_token
            self.bridge.device_token = request.device_token
            
            # Validate authentication
            if not self.bridge.validate_auth():
                return
            
            # Handle command
            if request.command in self.commands:
                try:
                    result = self.commands[request.command](request.args)
                    response = ModuleResponse(success=True, data=result)
                except Exception as e:
                    response = ModuleResponse(success=False, data={}, error=str(e))
            else:
                response = ModuleResponse(
                    success=False, 
                    data={}, 
                    error=f"Unknown command: {request.command}"
                )
            
            # Send response
            self.bridge.send_response(response)
            
        except KeyboardInterrupt:
            self.bridge.send_error("Module execution interrupted")
        except Exception as e:
            self.bridge.send_error(f"Module execution failed: {e}")


def progress_generator(total: int, message: str = "") -> Generator[int, None, None]:
    """Generator for progress updates"""
    for i in range(total + 1):
        yield i
        if i % (total // 10) == 0 or i == total:  # Update every 10%
            sys.stdout.write(json.dumps({
                "stage": "processing",
                "current": i,
                "total": total,
                "percentage": (i / total * 100) if total > 0 else 0,
                "message": message,
                "timestamp": time.time()
            }) + '\n')
            sys.stdout.flush()


def validate_request(request: ModuleRequest) -> Optional[str]:
    """Validate module request"""
    if not request.command:
        return "Command is required"
    
    if request.timeout <= 0:
        return "Timeout must be positive"
    
    return None


def create_error_response(error: str) -> ModuleResponse:
    """Create error response"""
    return ModuleResponse(success=False, data={}, error=error)


def create_success_response(data: Dict[str, Any]) -> ModuleResponse:
    """Create success response"""
    return ModuleResponse(success=True, data=data)


# Utility functions for common operations
def get_auth_header(auth_token: str) -> Dict[str, str]:
    """Get authorization header"""
    return {"Authorization": f"Bearer {auth_token}"}


def format_size(bytes_size: int) -> str:
    """Format bytes to human-readable size"""
    if bytes_size <= 0:
        return "0 B"
    
    units = ["B", "KB", "MB", "GB", "TB"]
    i = 0
    while bytes_size >= 1024 and i < len(units) - 1:
        bytes_size /= 1024.0
        i += 1
    
    return f"{bytes_size:.1f} {units[i]}"


def check_ffmpeg() -> bool:
    """Check if FFmpeg is available"""
    import shutil
    return shutil.which("ffmpeg") is not None


if __name__ == "__main__":
    # Example usage
    module = ModuleBase()
    
    def example_command(args):
        """Example command handler"""
        return {"message": "Hello from Python module", "args": args}
    
    module.register_command("example", example_command)
    module.run()
