#!/usr/bin/env python3

import sys
import argparse
from argparse import RawTextHelpFormatter

# Custom usage / help menu.
class HelpFormatter(argparse.HelpFormatter):
    def add_usage(self, usage, actions, groups, prefix=None):
        if prefix is None:
            prefix = ''
        return super(HelpFormatter, self).add_usage(
            usage, actions, groups, prefix)


# Custom help menu.
custom_usage = f"""
  
Krouter
{'-'*100}\n
Usage Examples: 
  python krouter.py --debug
  
"""

# Define parser
parser = argparse.ArgumentParser(formatter_class=HelpFormatter, description='', usage=custom_usage, add_help=False)


# Main Options.
main_group = parser.add_argument_group('Main Options')
main_group.add_argument('--debug', dest='loglevel', action='store_true', help='Set logging level [DEBUG]')


# Print 'help'.
if len(sys.argv) == 2:
    if sys.argv[1] == '-h' or sys.argv[1] == '--help':
        parser.print_help(sys.stderr)
        sys.exit(1)

# Print 'help' if no options are defined.
# if len(sys.argv) == 1 \
# or sys.argv[1] == '-h' \
# or sys.argv[1] == '--help':
#   parser.print_help(sys.stderr)
#   sys.exit(1)

