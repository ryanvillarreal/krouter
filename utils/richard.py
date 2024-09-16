#!/usr/bin/env python3

# from rich.table import Table
from rich import box
from rich.console import Console
from rich.logging import RichHandler
from rich.panel import Panel
from rich.syntax import Syntax
from rich.table import Table
from rich.text import Text
from rich.theme import Theme
from utils import arguments
import logging


# Argparse - init and parse.
args = arguments.parser.parse_args()
loglevel = args.loglevel
if loglevel:
	loglevel = "DEBUG"
else:
	loglevel = "WARNING"

# Rich console and theme init.
themefile = './utils/theme.ini'
mytheme = Theme().read(themefile)
console = Console(theme=mytheme)

# logger - Rich
logging.basicConfig(
	# filename='',
	level=loglevel,
    # level="DEBUG",
	format='%(message)s',
	datefmt='[%X]',
	handlers=[RichHandler(console=console, rich_tracebacks=True, omit_repeated_times=False)]
	)
logger = logging.getLogger('rich')