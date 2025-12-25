#!/usr/bin/env python3
"""
Converso YouTube Downloader
A Product of Converso Empire
World-class video & audio downloader with beautiful CLI interface
"""

import sys
import os
import shutil
import argparse
import math
import json
from pathlib import Path
from typing import List, Dict, Optional, Tuple
import yt_dlp

try:
    from rich.console import Console
    from rich.table import Table
    from rich.panel import Panel
    from rich.progress import (
        Progress, 
        SpinnerColumn, 
        BarColumn, 
        TextColumn, 
        DownloadColumn, 
        TransferSpeedColumn, 
        TimeRemainingColumn,
        TaskProgressColumn
    )
    from rich.prompt import Prompt, Confirm
    from rich.text import Text
    from rich.align import Align
    from rich import box
    from rich.layout import Layout
    from rich.live import Live
    RICH_AVAILABLE = True
except ImportError:
    RICH_AVAILABLE = False
    print("⚠️  Warning: 'rich' library not found. Install with: pip install rich")
    print("    Falling back to basic output...\n")

# Initialize console
console = Console() if RICH_AVAILABLE else None

# ==================== CONFIGURATION ====================

BRAND_NAME = "Converso YouTube Downloader"
BRAND_TAGLINE = "A Product of Converso Empire"
VERSION = "3.0.0"

CONFIG_DIR = Path.home() / ".converso_yt"
CONFIG_FILE = CONFIG_DIR / "config.json"
CACHE_DIR = CONFIG_DIR / "cache"

# Default configuration
DEFAULT_CONFIG = {
    "default_output_dir": str(Path.home() / "Downloads" / "Converso_YT"),
    "default_quality": "best",
    "default_format": "mp4",
    "concurrency": 10,
    "show_thumbnails": True
}

# ==================== BRANDING & STYLING ====================

LOGO = """
 ██████╗ ██████╗ ███╗   ██╗██╗   ██╗███████╗██████╗ ███████╗  ██████╗ 
██╔════╝██╔═══██╗████╗  ██║██║   ██║██╔════╝██╔══██╗██╔════╝ ██╔═══██╗
██║     ██║   ██║██╔██╗ ██║██║   ██║█████╗  ██████╔╝███████╗ ██║   ██║
██║     ██║   ██║██║╚██╗██║╚██╗ ██╔╝██╔══╝  ██╔══██╗╚════██║ ██║   ██║
╚██████╗╚██████╔╝██║ ╚████║ ╚████╔╝ ███████╗██║  ██║███████║ ╚██████╔╝
 ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝╚══════╝  ╚═════╝ 
"""

def show_header(show_tagline: bool = True):
    """Display the branded header"""
    if RICH_AVAILABLE:
        logo_text = Text(LOGO, style="bold cyan", justify="center")
        console.print()
        console.print(logo_text)
        console.print(
            Align.center(Text(f"YouTube Downloader • v{VERSION}", style="bold white")),
        )
        if show_tagline:
            console.print(
                Align.center(Text(BRAND_TAGLINE, style="dim italic cyan"))
            )
            console.print(
                Align.center(Text("🎬 World-class video & audio downloader", style="dim"))
            )
        console.print()
    else:
        print(LOGO)
        print(f"YouTube Downloader • v{VERSION}")
        if show_tagline:
            print(BRAND_TAGLINE)
            print("🎬 World-class video & audio downloader\n")

def print_success(msg: str):
    if RICH_AVAILABLE:
        console.print(f"[bold green]✓[/] {msg}")
    else:
        print(f"✓ {msg}")

def print_error(msg: str):
    if RICH_AVAILABLE:
        console.print(f"[bold red]✗[/] {msg}")
    else:
        print(f"✗ {msg}")

def print_warning(msg: str):
    if RICH_AVAILABLE:
        console.print(f"[bold yellow]⚠[/] {msg}")
    else:
        print(f"⚠ {msg}")

def print_info(msg: str):
    if RICH_AVAILABLE:
        console.print(f"[bold cyan]ℹ[/] {msg}")
    else:
        print(f"ℹ {msg}")

# ==================== CONFIGURATION MANAGEMENT ====================

def load_config() -> Dict:
    """Load configuration from file"""
    if CONFIG_FILE.exists():
        try:
            with open(CONFIG_FILE, 'r') as f:
                return {**DEFAULT_CONFIG, **json.load(f)}
        except:
            pass
    return DEFAULT_CONFIG.copy()

def save_config(config: Dict):
    """Save configuration to file"""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    with open(CONFIG_FILE, 'w') as f:
        json.dump(config, f, indent=2)

def ensure_output_dir(path: str) -> Path:
    """Ensure output directory exists"""
    output_path = Path(path)
    output_path.mkdir(parents=True, exist_ok=True)
    return output_path

# ==================== UTILITIES ====================

def check_ffmpeg() -> bool:
    """Check if FFmpeg is available"""
    return shutil.which("ffmpeg") is not None

def human_size(n: Optional[int]) -> str:
    """Convert bytes to human-readable size"""
    if not n or n <= 0:
        return "Unknown"
    units = ["B", "KB", "MB", "GB", "TB"]
    i = int(math.floor(math.log(n, 1024)))
    i = min(i, len(units) - 1)
    val = n / (1024 ** i)
    return f"{val:.1f} {units[i]}"

def format_duration(seconds: Optional[int]) -> str:
    """Format duration in seconds to MM:SS"""
    if not seconds:
        return "Unknown"
    m, s = divmod(seconds, 60)
    h, m = divmod(m, 60)
    if h > 0:
        return f"{h}:{m:02d}:{s:02d}"
    return f"{m}:{s:02d}"

def fetch_info(url: str) -> Dict:
    """Fetch video information"""
    if RICH_AVAILABLE:
        with console.status("[bold cyan]Fetching video information...", spinner="dots"):
            ydl_opts = {
                'quiet': True, 
                'skip_download': True, 
                'no_warnings': True,
                'extract_flat': False
            }
            with yt_dlp.YoutubeDL(ydl_opts) as ydl:
                return ydl.extract_info(url, download=False)
    else:
        print("Fetching video information...")
        ydl_opts = {
            'quiet': True, 
            'skip_download': True, 
            'no_warnings': True,
            'extract_flat': False
        }
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            return ydl.extract_info(url, download=False)

# ==================== FORMAT HANDLING ====================

class FormatManager:
    """Manages video/audio format selection and categorization"""
    
    @staticmethod
    def categorize_formats(formats: List[Dict]) -> Tuple[List[Dict], List[Dict], List[Dict]]:
        """Categorize formats into audio-only, video-only, and progressive"""
        audio_only = []
        video_only = []
        progressive = []
        
        for f in formats:
            vcodec = f.get('vcodec', 'none')
            acodec = f.get('acodec', 'none')
            
            if acodec not in (None, 'none') and vcodec in (None, 'none'):
                audio_only.append(f)
            elif vcodec not in (None, 'none') and acodec in (None, 'none'):
                video_only.append(f)
            elif vcodec not in (None, 'none') and acodec not in (None, 'none'):
                progressive.append(f)
        
        return audio_only, video_only, progressive
    
    @staticmethod
    def sort_formats(formats: List[Dict], format_type: str) -> List[Dict]:
        """Sort formats by quality"""
        if format_type == "audio":
            return sorted(formats, 
                         key=lambda f: (f.get('abr', 0), f.get('asr', 0)), 
                         reverse=True)
        elif format_type == "video":
            return sorted(formats, 
                         key=lambda f: (f.get('height', 0), f.get('fps', 0), f.get('tbr', 0)), 
                         reverse=True)
        else:  # progressive
            return sorted(formats, 
                         key=lambda f: (f.get('height', 0), f.get('fps', 0), f.get('tbr', 0)), 
                         reverse=True)
    
    @staticmethod
    def get_best_video(formats: List[Dict]) -> Optional[Dict]:
        """Select best video format with codec preference"""
        videos = [f for f in formats 
                 if f.get('vcodec', 'none') not in (None, 'none') 
                 and f.get('acodec', 'none') in (None, 'none')]
        
        if not videos:
            return None
        
        # Codec preference order
        codec_order = ['av01', 'vp9', 'avc1', 'h264']
        
        for codec in codec_order:
            candidates = [f for f in videos 
                         if (f.get('vcodec', '') or '').lower().startswith(codec)]
            if candidates:
                return max(candidates, 
                          key=lambda f: (f.get('height', 0), f.get('fps', 0), f.get('tbr', 0)))
        
        return max(videos, 
                  key=lambda f: (f.get('height', 0), f.get('fps', 0), f.get('tbr', 0)))
    
    @staticmethod
    def get_best_audio(formats: List[Dict]) -> Optional[Dict]:
        """Select best audio format"""
        audios = [f for f in formats 
                 if f.get('acodec', 'none') not in (None, 'none') 
                 and f.get('vcodec', 'none') in (None, 'none')]
        
        if not audios:
            return None
        
        return max(audios, key=lambda f: (f.get('abr', 0), f.get('asr', 0)))
    
    @staticmethod
    def suggest_container(fmt: Dict, for_audio: bool = False) -> str:
        """Suggest best container format"""
        ext = (fmt.get('ext', '') or '').lower()
        
        if for_audio:
            acodec = (fmt.get('acodec', '') or '').lower()
            if 'opus' in acodec or ext == 'webm':
                return 'opus'
            if ext == 'm4a' or 'mp4a' in acodec:
                return 'm4a'
            return 'mp3'
        else:
            if ext in ('mp4', 'webm', 'mkv'):
                return ext
            return 'mp4'

# ==================== DISPLAY FUNCTIONS ====================

class DisplayManager:
    """Manages all display operations"""
    
    @staticmethod
    def show_video_info(info: Dict):
        """Display video metadata"""
        title = info.get('title', 'Unknown')
        duration = info.get('duration')
        uploader = info.get('uploader', 'Unknown')
        view_count = info.get('view_count')
        upload_date = info.get('upload_date')
        
        if RICH_AVAILABLE:
            table = Table(show_header=False, box=None, padding=(0, 2), border_style="cyan")
            table.add_column("Key", style="bold cyan", width=15, no_wrap=True)
            table.add_column("Value", style="white")
            
            # Truncate long titles
            display_title = title if len(title) <= 70 else title[:67] + "..."
            table.add_row("📹 Title", display_title)
            table.add_row("👤 Uploader", uploader)
            
            if duration:
                table.add_row("⏱️  Duration", format_duration(duration))
            
            if view_count:
                table.add_row("👁️  Views", f"{view_count:,}")
            
            if upload_date:
                # Format date nicely
                try:
                    date_str = f"{upload_date[:4]}-{upload_date[4:6]}-{upload_date[6:]}"
                    table.add_row("📅 Uploaded", date_str)
                except:
                    table.add_row("📅 Uploaded", upload_date)
            
            console.print(Panel(table, title="[bold cyan]Video Information[/]", 
                              border_style="cyan", box=box.ROUNDED))
            console.print()
        else:
            print("\n" + "="*50)
            print("VIDEO INFORMATION")
            print("="*50)
            print(f"Title: {title}")
            print(f"Uploader: {uploader}")
            if duration:
                print(f"Duration: {format_duration(duration)}")
            if view_count:
                print(f"Views: {view_count:,}")
            if upload_date:
                print(f"Uploaded: {upload_date}")
            print("="*50 + "\n")
    
    @staticmethod
    def show_formats_table(formats: List[Dict], title: str, format_type: str):
        """Display formats in a table"""
        if not formats:
            print_info(f"No {title.lower()} available")
            return
        
        if RICH_AVAILABLE:
            table = Table(title=f"[bold cyan]{title}[/]", 
                         box=box.ROUNDED, 
                         border_style="cyan")
            
            table.add_column("#", style="bold blue", justify="center", width=4)
            table.add_column("ID", style="yellow", justify="center", width=8)
            table.add_column("Ext", style="green", justify="center", width=6)
            
            if format_type == "audio":
                table.add_column("Codec", style="magenta", width=12)
                table.add_column("Bitrate", style="cyan", justify="right", width=10)
                table.add_column("Sample Rate", style="white", justify="right", width=11)
                table.add_column("Size", style="green", justify="right", width=10)
                
                for idx, f in enumerate(formats):
                    fid = str(f.get('format_id', 'N/A'))
                    ext = f.get('ext', 'N/A')
                    codec = f.get('acodec', 'N/A')[:12]
                    bitrate = f"{int(f.get('abr', 0))} kbps" if f.get('abr') else "N/A"
                    asr = f"{int(f.get('asr', 0))} Hz" if f.get('asr') else "N/A"
                    size = human_size(f.get('filesize') or f.get('filesize_approx'))
                    table.add_row(str(idx), fid, ext, codec, bitrate, asr, size)
                    
            elif format_type == "video":
                table.add_column("Codec", style="magenta", width=14)
                table.add_column("Resolution", style="cyan", justify="right", width=10)
                table.add_column("FPS", style="white", justify="right", width=5)
                table.add_column("Bitrate", style="yellow", justify="right", width=10)
                table.add_column("Size", style="green", justify="right", width=10)
                
                for idx, f in enumerate(formats):
                    fid = str(f.get('format_id', 'N/A'))
                    ext = f.get('ext', 'N/A')
                    codec = f.get('vcodec', 'N/A')[:14]
                    height = f"{f.get('height', 0)}p" if f.get('height') else "N/A"
                    fps = f"{int(f.get('fps', 0))}" if f.get('fps') else "N/A"
                    bitrate = f"{int(f.get('tbr', 0))} kbps" if f.get('tbr') else "N/A"
                    size = human_size(f.get('filesize') or f.get('filesize_approx'))
                    table.add_row(str(idx), fid, ext, codec, height, fps, bitrate, size)
                    
            else:  # progressive
                table.add_column("Video", style="magenta", width=14)
                table.add_column("Audio", style="magenta", width=12)
                table.add_column("Resolution", style="cyan", justify="right", width=10)
                table.add_column("Size", style="green", justify="right", width=10)
                
                for idx, f in enumerate(formats):
                    fid = str(f.get('format_id', 'N/A'))
                    ext = f.get('ext', 'N/A')
                    vcodec = f.get('vcodec', 'N/A')[:14]
                    acodec = f.get('acodec', 'N/A')[:12]
                    height = f"{f.get('height', 0)}p" if f.get('height') else "N/A"
                    size = human_size(f.get('filesize') or f.get('filesize_approx'))
                    table.add_row(str(idx), fid, ext, vcodec, acodec, height, size)
            
            console.print(table)
            console.print()
        else:
            print(f"\n{title}")
            print("-" * 50)
            for idx, f in enumerate(formats):
                fid = f.get('format_id', 'N/A')
                ext = f.get('ext', 'N/A')
                print(f"[{idx}] ID: {fid}, Ext: {ext}")
            print()

# ==================== DOWNLOAD MANAGER ====================

class DownloadManager:
    """Manages download operations with progress tracking"""
    
    def __init__(self):
        self.progress = None
        self.live = None
        self.current_task = None
        
    def create_progress_hook(self):
        """Create progress hook for yt-dlp"""
        def hook(d):
            if not RICH_AVAILABLE or not self.progress or self.current_task is None:
                return
            
            status = d.get('status')
            
            if status == 'downloading':
                downloaded = d.get('downloaded_bytes', 0)
                total = d.get('total_bytes') or d.get('total_bytes_estimate', 0)
                
                if total > 0:
                    self.progress.update(
                        self.current_task,
                        completed=downloaded,
                        total=total
                    )
            
            elif status == 'finished':
                if self.progress and self.current_task is not None:
                    self.progress.update(self.current_task, completed=100, total=100)
        
        return hook
    
    def execute_download(self, url: str, ydl_opts: Dict, description: str = "Downloading") -> bool:
        """Execute download with progress tracking"""
        try:
            if RICH_AVAILABLE:
                # Setup progress display
                self.progress = Progress(
                    SpinnerColumn(),
                    TextColumn("[bold blue]{task.description}"),
                    BarColumn(bar_width=40),
                    TaskProgressColumn(),
                    DownloadColumn(),
                    TransferSpeedColumn(),
                    TimeRemainingColumn(),
                    console=console
                )
                
                with self.progress:
                    self.current_task = self.progress.add_task(
                        f"[cyan]{description}", 
                        total=100
                    )
                    
                    # Add progress hook
                    ydl_opts['progress_hooks'] = [self.create_progress_hook()]
                    
                    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
                        ydl.download([url])
                
                console.print()  # Add newline after progress
            else:
                print(f"\n{description}...")
                with yt_dlp.YoutubeDL(ydl_opts) as ydl:
                    ydl.download([url])
                print()
            
            return True
            
        except Exception as e:
            console.print() if RICH_AVAILABLE else print()
            print_error(f"Download failed: {str(e)}")
            return False
        
        finally:
            self.progress = None
            self.current_task = None
            self.live = None

# ==================== DOWNLOAD OPTIONS BUILDER ====================

class DownloadOptionsBuilder:
    """Builds yt-dlp download options"""
    
    @staticmethod
    def common_opts(output_path: str, concurrency: int = 10) -> Dict:
        """Common yt-dlp options"""
        return {
            'outtmpl': str(Path(output_path) / '%(title)s [%(id)s].%(ext)s'),
            'restrictfilenames': True,
            'windowsfilenames': True,
            'prefer_ffmpeg': True,
            'quiet': True,
            'no_warnings': True,
            'retries': 10,
            'fragment_retries': 10,
            'skip_unavailable_fragments': True,
            'socket_timeout': 30,
            'continuedl': True,
            'concurrent_fragment_downloads': max(1, min(concurrency, 16)),
            'noprogress': True,  # We handle progress ourselves
        }
    
    @staticmethod
    def build_audio_opts(fmt: Dict, container: str, output_path: str, concurrency: int) -> Dict:
        """Build options for audio download"""
        opts = DownloadOptionsBuilder.common_opts(output_path, concurrency)
        opts['format'] = f"{fmt['format_id']}/bestaudio/best"
        opts['postprocessors'] = [{
            'key': 'FFmpegExtractAudio',
            'preferredcodec': container,
            'preferredquality': '0'
        }]
        return opts
    
    @staticmethod
    def build_video_opts(fmt: Dict, container: Optional[str], output_path: str, concurrency: int) -> Dict:
        """Build options for video download"""
        opts = DownloadOptionsBuilder.common_opts(output_path, concurrency)
        opts['format'] = f"{fmt['format_id']}/bestvideo/best"
        if container:
            opts['postprocessors'] = [{
                'key': 'FFmpegVideoRemuxer',
                'preferedformat': container
            }]
        return opts
    
    @staticmethod
    def build_merge_opts(vfmt: Dict, afmt: Dict, container: str, output_path: str, concurrency: int) -> Dict:
        """Build options for merged video+audio download"""
        opts = DownloadOptionsBuilder.common_opts(output_path, concurrency)
        opts['format'] = f"{vfmt['format_id']}+{afmt['format_id']}/bestvideo+bestaudio/best"
        opts['merge_output_format'] = container
        return opts

# ==================== INTERACTIVE MODE ====================

class InteractiveMode:
    """Handles interactive download workflow"""
    
    def __init__(self, config: Dict):
        self.config = config
        self.fm = FormatManager()
        self.dm = DisplayManager()
        self.dlm = DownloadManager()
        self.opts = DownloadOptionsBuilder()
    
    def prompt_choice(self, message: str, choices: List[str], default: int = 0) -> int:
        """Prompt user for choice"""
        if RICH_AVAILABLE:
            console.print(f"\n[bold cyan]{message}[/]")
            for i, choice in enumerate(choices):
                console.print(f"[bold blue][{i}][/] {choice}")
            console.print()
            
            while True:
                sel = Prompt.ask("[bold green]Enter selection[/]", default=str(default))
                if sel.isdigit() and 0 <= int(sel) < len(choices):
                    return int(sel)
                print_warning(f"Invalid selection. Enter 0-{len(choices)-1}")
        else:
            print(f"\n{message}")
            for i, choice in enumerate(choices):
                print(f"[{i}] {choice}")
            
            while True:
                sel = input(f"Enter selection (default {default}): ").strip() or str(default)
                if sel.isdigit() and 0 <= int(sel) < len(choices):
                    return int(sel)
                print(f"Invalid selection. Enter 0-{len(choices)-1}")
    
    def prompt_format(self, formats: List[Dict], best_fmt: Optional[Dict]) -> Optional[Dict]:
        """Prompt user to select a format"""
        if not formats:
            return None
        
        if RICH_AVAILABLE:
            console.print("[bold green][[b]][/] Select best quality automatically")
            console.print("[bold yellow][[q]][/] Cancel")
            console.print()
            
            while True:
                sel = Prompt.ask("[bold cyan]Select format[/]", default="b").strip().lower()
                if sel == 'q':
                    return None
                if sel == 'b' and best_fmt:
                    return best_fmt
                if sel.isdigit() and 0 <= int(sel) < len(formats):
                    return formats[int(sel)]
                print_warning("Invalid selection")
        else:
            print("[b] Best quality (default)")
            print("[q] Cancel")
            
            while True:
                sel = input("Select format (b/q/number): ").strip().lower() or "b"
                if sel == 'q':
                    return None
                if sel == 'b' and best_fmt:
                    return best_fmt
                if sel.isdigit() and 0 <= int(sel) < len(formats):
                    return formats[int(sel)]
                print("Invalid selection")
    
    def run(self, url: str):
        """Run interactive mode"""
        if RICH_AVAILABLE:
            console.clear()
        
        show_header()
        
        # Fetch video info
        try:
            info = fetch_info(url)
        except Exception as e:
            print_error(f"Failed to fetch video info: {str(e)}")
            return
        
        self.dm.show_video_info(info)
        
        # Get and categorize formats
        all_formats = info.get('formats', [])
        if not all_formats:
            print_error("No formats available for this video")
            return
        
        audio_only, video_only, progressive = self.fm.categorize_formats(all_formats)
        audio_only = self.fm.sort_formats(audio_only, "audio")
        video_only = self.fm.sort_formats(video_only, "video")
        progressive = self.fm.sort_formats(progressive, "progressive")
        
        # Select download mode
        mode = self.prompt_choice(
            "What would you like to download?",
            [
                "🎵 Audio only (extract audio)",
                "🎬 Video only (no audio)",
                "🎭 Best video + audio (merged)",
                "📹 Progressive (video with audio)"
            ]
        )
        
        output_path = ensure_output_dir(self.config['default_output_dir'])
        
        # Handle each mode
        if mode == 0:  # Audio only
            self._download_audio(url, audio_only, all_formats, output_path)
        elif mode == 1:  # Video only
            self._download_video(url, video_only, all_formats, output_path)
        elif mode == 2:  # Merged
            self._download_merged(url, video_only, audio_only, all_formats, output_path)
        else:  # Progressive
            self._download_progressive(url, progressive, output_path)
    
    def _download_audio(self, url: str, audio_formats: List[Dict], 
                        all_formats: List[Dict], output_path: Path):
        """Handle audio download"""
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        
        if not audio_formats:
            print_error("No audio formats available")
            return
        
        self.dm.show_formats_table(audio_formats, "Available Audio Formats", "audio")
        
        best_audio = self.fm.get_best_audio(all_formats)
        selected = self.prompt_format(audio_formats, best_audio)
        
        if not selected:
            print_warning("No format selected")
            return
        
        if not check_ffmpeg():
            print_error("FFmpeg is required for audio conversion")
            print_info("Install FFmpeg: https://ffmpeg.org/download.html")
            return
        
        # Select container
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        
        suggested = self.fm.suggest_container(selected, for_audio=True)
        container_idx = self.prompt_choice(
            "Choose audio format:",
            ['m4a', 'mp3', 'opus', 'aac', 'flac', 'wav'],
            default=['m4a', 'mp3', 'opus'].index(suggested) if suggested in ['m4a', 'mp3', 'opus'] else 1
        )
        containers = ['m4a', 'mp3', 'opus', 'aac', 'flac', 'wav']
        container = containers[container_idx]
        
        # Download
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        console.print(f"[bold cyan]🎵 Downloading Audio[/]") if RICH_AVAILABLE else print("Downloading Audio...")
        console.print() if RICH_AVAILABLE else print()
        
        opts = self.opts.build_audio_opts(
            selected, container, str(output_path), self.config['concurrency']
        )
        
        if self.dlm.execute_download(url, opts, "Downloading audio"):
            print_success(f"Audio downloaded to: {output_path}")
    
    def _download_video(self, url: str, video_formats: List[Dict], 
                        all_formats: List[Dict], output_path: Path):
        """Handle video download"""
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        
        if not video_formats:
            print_error("No video formats available")
            return
        
        self.dm.show_formats_table(video_formats, "Available Video Formats", "video")
        
        best_video = self.fm.get_best_video(all_formats)
        selected = self.prompt_format(video_formats, best_video)
        
        if not selected:
            print_warning("No format selected")
            return
        
        # Select container
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        
        container_idx = self.prompt_choice(
            "Choose output format:",
            ['mp4', 'mkv', 'webm'],
            default=0
        )
        containers = ['mp4', 'mkv', 'webm']
        container = containers[container_idx] if check_ffmpeg() else None
        
        if not check_ffmpeg():
            print_warning("FFmpeg not found - using original container")
        
        # Download
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        console.print(f"[bold cyan]🎬 Downloading Video[/]") if RICH_AVAILABLE else print("Downloading Video...")
        console.print() if RICH_AVAILABLE else print()
        
        opts = self.opts.build_video_opts(
            selected, container, str(output_path), self.config['concurrency']
        )
        
        if self.dlm.execute_download(url, opts, "Downloading video"):
            print_success(f"Video downloaded to: {output_path}")
    
    def _download_merged(self, url: str, video_formats: List[Dict], 
                         audio_formats: List[Dict], all_formats: List[Dict], 
                         output_path: Path):
        """Handle merged download"""
        if not video_formats or not audio_formats:
            print_error("Insufficient formats for merging")
            return
        
        if not check_ffmpeg():
            print_error("FFmpeg is required for merging")
            print_info("Install FFmpeg: https://ffmpeg.org/download.html")
            return
        
        # Select video
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        
        self.dm.show_formats_table(video_formats, "Available Video Formats", "video")
        best_video = self.fm.get_best_video(all_formats)
        selected_video = self.prompt_format(video_formats, best_video)
        
        if not selected_video:
            print_warning("No video format selected")
            return
        
        # Select audio
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        
        self.dm.show_formats_table(audio_formats, "Available Audio Formats", "audio")
        best_audio = self.fm.get_best_audio(all_formats)
        selected_audio = self.prompt_format(audio_formats, best_audio)
        
        if not selected_audio:
            print_warning("No audio format selected")
            return
        
        # Select container
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        
        container_idx = self.prompt_choice(
            "Choose output format:",
            ['mp4', 'mkv', 'webm'],
            default=0
        )
        containers = ['mp4', 'mkv', 'webm']
        container = containers[container_idx]
        
        # Download
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        console.print(f"[bold cyan]🎭 Merging Video + Audio[/]") if RICH_AVAILABLE else print("Merging...")
        console.print() if RICH_AVAILABLE else print()
        
        opts = self.opts.build_merge_opts(
            selected_video, selected_audio, container, 
            str(output_path), self.config['concurrency']
        )
        
        if self.dlm.execute_download(url, opts, "Downloading and merging"):
            print_success(f"Merged video downloaded to: {output_path}")
    
    def _download_progressive(self, url: str, progressive_formats: List[Dict], 
                              output_path: Path):
        """Handle progressive download"""
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        
        if not progressive_formats:
            print_error("No progressive formats available")
            return
        
        self.dm.show_formats_table(progressive_formats, "Available Progressive Formats", "progressive")
        
        best_prog = progressive_formats[0] if progressive_formats else None
        selected = self.prompt_format(progressive_formats, best_prog)
        
        if not selected:
            print_warning("No format selected")
            return
        
        # Select container
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        
        container_idx = self.prompt_choice(
            "Choose output format:",
            ['mp4', 'mkv', 'webm'],
            default=0
        )
        containers = ['mp4', 'mkv', 'webm']
        container = containers[container_idx] if check_ffmpeg() else None
        
        if not check_ffmpeg():
            print_warning("FFmpeg not found - using original container")
        
        # Download
        if RICH_AVAILABLE:
            console.clear()
        show_header(False)
        console.print(f"[bold cyan]📹 Downloading Progressive Format[/]") if RICH_AVAILABLE else print("Downloading...")
        console.print() if RICH_AVAILABLE else print()
        
        opts = self.opts.build_video_opts(
            selected, container, str(output_path), self.config['concurrency']
        )
        
        if self.dlm.execute_download(url, opts, "Downloading"):
            print_success(f"Video downloaded to: {output_path}")

# ==================== CLI MODE ====================

def cli_list(url: str):
    """List all formats"""
    show_header()
    
    try:
        info = fetch_info(url)
    except Exception as e:
        print_error(f"Failed to fetch info: {str(e)}")
        return
    
    dm = DisplayManager()
    fm = FormatManager()
    
    dm.show_video_info(info)
    
    all_formats = info.get('formats', [])
    audio_only, video_only, progressive = fm.categorize_formats(all_formats)
    
    audio_only = fm.sort_formats(audio_only, "audio")
    video_only = fm.sort_formats(video_only, "video")
    progressive = fm.sort_formats(progressive, "progressive")
    
    dm.show_formats_table(audio_only, "Audio-only Formats", "audio")
    dm.show_formats_table(video_only, "Video-only Formats", "video")
    dm.show_formats_table(progressive, "Progressive Formats", "progressive")

def cli_download(args, config: Dict):
    """CLI download mode"""
    show_header()
    
    try:
        info = fetch_info(args.url)
    except Exception as e:
        print_error(f"Failed to fetch info: {str(e)}")
        return
    
    dm = DisplayManager()
    fm = FormatManager()
    dlm = DownloadManager()
    opts_builder = DownloadOptionsBuilder()
    
    dm.show_video_info(info)
    
    all_formats = info.get('formats', [])
    audio_only, video_only, progressive = fm.categorize_formats(all_formats)
    
    audio_only = fm.sort_formats(audio_only, "audio")
    video_only = fm.sort_formats(video_only, "video")
    progressive = fm.sort_formats(progressive, "progressive")
    
    output_path = ensure_output_dir(args.output or config['default_output_dir'])
    
    # Handle different modes
    if args.mode == 'audio':
        if not audio_only:
            print_error("No audio formats available")
            return
        
        fmt = fm.get_best_audio(all_formats) if args.best else (
            audio_only[args.a_index] if args.a_index is not None and 0 <= args.a_index < len(audio_only) else None
        )
        
        if not fmt:
            print_error("Invalid format selection")
            return
        
        if not check_ffmpeg():
            print_error("FFmpeg required for audio conversion")
            return
        
        container = args.container or fm.suggest_container(fmt, for_audio=True)
        opts = opts_builder.build_audio_opts(fmt, container, str(output_path), config['concurrency'])
        
        console.print(f"\n[bold cyan]🎵 Downloading Audio[/]\n") if RICH_AVAILABLE else print("\nDownloading Audio...\n")
        
        if dlm.execute_download(args.url, opts, "Downloading audio"):
            print_success(f"Downloaded to: {output_path}")
    
    elif args.mode == 'video':
        if not video_only:
            print_error("No video formats available")
            return
        
        fmt = fm.get_best_video(all_formats) if args.best else (
            video_only[args.v_index] if args.v_index is not None and 0 <= args.v_index < len(video_only) else None
        )
        
        if not fmt:
            print_error("Invalid format selection")
            return
        
        container = args.container if check_ffmpeg() else None
        opts = opts_builder.build_video_opts(fmt, container, str(output_path), config['concurrency'])
        
        console.print(f"\n[bold cyan]🎬 Downloading Video[/]\n") if RICH_AVAILABLE else print("\nDownloading Video...\n")
        
        if dlm.execute_download(args.url, opts, "Downloading video"):
            print_success(f"Downloaded to: {output_path}")
    
    elif args.mode == 'merge':
        if not video_only or not audio_only:
            print_error("Insufficient formats for merging")
            return
        
        if not check_ffmpeg():
            print_error("FFmpeg required for merging")
            return
        
        vfmt = fm.get_best_video(all_formats) if args.best else (
            video_only[args.v_index] if args.v_index is not None and 0 <= args.v_index < len(video_only) else None
        )
        afmt = fm.get_best_audio(all_formats) if args.best else (
            audio_only[args.a_index] if args.a_index is not None and 0 <= args.a_index < len(audio_only) else None
        )
        
        if not vfmt or not afmt:
            print_error("Invalid format selection")
            return
        
        container = args.container or 'mp4'
        opts = opts_builder.build_merge_opts(vfmt, afmt, container, str(output_path), config['concurrency'])
        
        console.print(f"\n[bold cyan]🎭 Merging Video + Audio[/]\n") if RICH_AVAILABLE else print("\nMerging...\n")
        
        if dlm.execute_download(args.url, opts, "Downloading and merging"):
            print_success(f"Downloaded to: {output_path}")
    
    elif args.mode == 'progressive':
        if not progressive:
            print_error("No progressive formats available")
            return
        
        fmt = progressive[0] if args.best else (
            progressive[args.index] if args.index is not None and 0 <= args.index < len(progressive) else None
        )
        
        if not fmt:
            print_error("Invalid format selection")
            return
        
        container = args.container if check_ffmpeg() else None
        opts = opts_builder.build_video_opts(fmt, container, str(output_path), config['concurrency'])
        
        console.print(f"\n[bold cyan]📹 Downloading Progressive[/]\n") if RICH_AVAILABLE else print("\nDownloading...\n")
        
        if dlm.execute_download(args.url, opts, "Downloading"):
            print_success(f"Downloaded to: {output_path}")

# ==================== ARGUMENT PARSER ====================

def build_parser() -> argparse.ArgumentParser:
    """Build argument parser"""
    parser = argparse.ArgumentParser(
        prog="converso-yt",
        description=f"{BRAND_NAME} v{VERSION} - {BRAND_TAGLINE}",
        epilog="🎬 World-class video & audio downloader"
    )
    
    parser.add_argument("--version", action="version", version=f"{BRAND_NAME} v{VERSION}")
    parser.add_argument("--concurrency", type=int, help="Concurrent fragment downloads")
    parser.add_argument("--output", "-o", type=str, help="Output directory")
    
    subparsers = parser.add_subparsers(dest="command")
    
    # List command
    list_parser = subparsers.add_parser("list", help="List available formats")
    list_parser.add_argument("url", help="Video URL")
    
    # Download command
    dl_parser = subparsers.add_parser("download", help="Download video/audio")
    dl_parser.add_argument("url", help="Video URL")
    dl_parser.add_argument("--mode", choices=["audio", "video", "merge", "progressive"],
                           required=True, help="Download mode")
    dl_parser.add_argument("--best", action="store_true", help="Select best quality")
    dl_parser.add_argument("--a-index", type=int, help="Audio format index")
    dl_parser.add_argument("--v-index", type=int, help="Video format index")
    dl_parser.add_argument("--index", type=int, help="Progressive format index")
    dl_parser.add_argument("--container", help="Output container format")
    
    return parser

# ==================== MAIN ====================

def main():
    """Main entry point"""
    parser = build_parser()
    args = parser.parse_args()
    
    # Load configuration
    config = load_config()
    if args.concurrency:
        config['concurrency'] = args.concurrency
    
    # Handle commands
    if args.command == "list":
        cli_list(args.url)
    elif args.command == "download":
        cli_download(args, config)
    else:
        # Interactive mode
        if RICH_AVAILABLE:
            console.clear()
        show_header()
        
        url = Prompt.ask("[bold cyan]🔗 Enter video URL[/]").strip() if RICH_AVAILABLE else input("Enter video URL: ").strip()
        
        if not url:
            print_error("URL is required")
            return
        
        interactive = InteractiveMode(config)
        interactive.run(url)

if __name__ == "__main__":
    main()
