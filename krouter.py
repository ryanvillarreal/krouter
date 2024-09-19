#!/usr/bin/env python3

from configparser import ConfigParser, ExtendedInterpolation
from simple_term_menu import TerminalMenu
from datetime import datetime
from utils import arguments
from utils import mkdir
from utils.packagemanager import PackageManager
from utils import richard as r
import logging
import os
import psutil
import re
import shutil
import socket
import subprocess
import sys
import time


# App - directories and filepaths.
app_filepath = __file__
app_dir = os.path.dirname(__file__)
bckup_dir = os.path.join(app_dir, "backups")
hostapd_dir = "/etc/hostapd"

# App - relative filepaths.
config_ini_fp = "configs/config.ini"

# Create required dirs.
directories = [bckup_dir, hostapd_dir]
dirs = [mkdir.mkdir(directory) for directory in directories]
[print(f"Created directory: {d}") for d in dirs if d is not None]
[r.logging.warning(f"Missing directory was created: {d}") for d in dirs if d is not None]

# Argparse - init and parse.
args = arguments.parser.parse_args()


def read_ini(ini_fp, converters="", delimiters='^'):
    ''' Read 'ini' config file and return ConfigParser Object '''
    converters = converters
    delimiters = delimiters
    cp_obj = ConfigParser(
        allow_no_value=True,
        converters=converters,
        delimiters=delimiters, 
        interpolation=ExtendedInterpolation()
        )
    cp_obj.optionxform = str
    if os.path.exists(ini_fp):
        cp_obj.read(ini_fp)
    else:
        r.logging.warning(f"ConfigParser file failed to load: '{ini_fp}'")
        return None
    return cp_obj


def time_stamp():
    '''Time Stamp used for start/stop and archive filename'''
    current_time = datetime.now()
    return current_time.strftime("%b-%d-%y-%H-%M-%S")


def copy_file(src, dst):
    ''' Copy file from src to dst via the shutil module.
    Return filepath.
    arg(s) src:str, dst:str'''
    
    try:
        filepath = shutil.copy(src, dst)
    except IOError as e:
        raise e
    except Exception as e:
        raise e
    else:
        return filepath


def backup_file(filepath, timestamp):
    bckup_subdir = os.path.join(bckup_dir, timestamp)
    bckup_ext = filepath + ".bak"
    mkdir.mkdir(bckup_subdir)
    bckup_fp = os.path.join(bckup_subdir, os.path.basename(bckup_ext))
    if os.path.isfile(filepath):
        fp = copy_file(filepath, bckup_fp)
        return {os.path.relpath(fp)}


def manage_service(action, service_name):
    try:
        command = ['sudo', 'systemctl', action, service_name]
        result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if result.returncode == 0:
            print(f"Success: '{' '.join(command)}'")
        else:
            print(f"Failed to {action} {service_name}. Error: {result.stderr.decode()}")
    except Exception as e:
        print(f"An error occurred: {e}")


def run_cmd(cmd):
    try:
        result = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)             
        if result.returncode == 0:
            return cmd
        else:
            r.logging.debug(f'Error: {result.stderr.decode()}')
            return f'FAILED: {cmd}'
    except Exception as e:
        print(f'An error occurred: {e}')


class BaseMenu:
    opt_packages_install = "[-] Packages - Install required packages"
    opt_packages_view = "[-] Packages - Show/hide installed packages"
    opt_services_start = "[-] Services - Enable & start services"
    opt_services_restart = "[-] Services - Restart services"
    opt_firewall_toggle = "[-] Firewall - Enable/disable firewall rules"
    opt_firewall_mitmproxy_toggle = "[-] Firewall - Enable/disable mitmproxy rules"
    opt_dhcp_leases_refresh = "[-] DHCP Leases - Refresh screen to view new leases"
    opt_exit = "[-] Exit"
    opt_back = "[-] Back"
    
    OPTIONS = [
        opt_packages_install,
        opt_packages_view,
        opt_services_start,
        opt_services_restart,
        opt_firewall_toggle,
        opt_firewall_mitmproxy_toggle,
        opt_dhcp_leases_refresh
    ]

    def __init__(self, cp_obj):
        self.cp_obj = cp_obj
        self.menu_options = self.OPTIONS + [i for i in self.cp_obj.sections()] + [self.opt_exit]
        self.menu = TerminalMenu(self.menu_options)

    def present_menu(self):
        index = self.menu.show()
        selected_menu_item = self.menu_options[index]
        return selected_menu_item


class MainMenu(BaseMenu):
    def __init__(self, cp_obj):
        super().__init__(cp_obj)
        self.menu_options = self.OPTIONS + [self.opt_exit]
        self.menu = TerminalMenu(self.menu_options)


class Session:
    def __init__(self, cp_obj, services):
        self.cp_obj = cp_obj
        self.interfaces = {k:v for k,v in self.cp_obj["interface"].items()}
        self.packages = [k for k in self.cp_obj["packages"]]
        self.services = services
        self.packages_toggle = False
        self.firewall_toggle = True
        self.mitmproxy_toggle = False
        self.fw_lst = []
        self.mitm_lst = []
        self.loglevel_numeric_value = r.logger.getEffectiveLevel()
        self.loglevel_name = r.logging.getLevelName(self.loglevel_numeric_value )
        # Toggle clear screen function off when loglevel is DEBUG..
        if not self.loglevel_name  == 'DEBUG':
            os.system("clear")
        
 
    def check_interface(self, interface):
        ''' Sets icon based on inteface status, I.e. Up, down or not found '''
        addrs = psutil.net_if_addrs()
        stats = psutil.net_if_stats()
        if interface in stats:
            is_up = stats[interface].isup
            if is_up:
                return f"[green]up[/green] {interface}"
            else:
                return f"[orange3]down[/orange3] {interface}"
        else:
            return f"[red]not-found[/red] {interface}"

    def run_firewall_router(self, rules):
        ''' Firewall Router rules - backup, flush and apply new rules for router to function '''
        self.fw_lst = []
        for rule in rules:
            # Remove rule if toggle is false.
            if not self.firewall_toggle:
                rule = rule.replace('-A', '-D')
            try:
                result = subprocess.run(rule, stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE, shell=True)             
                if result.returncode == 0:
                    if not self.firewall_toggle:
                        self.fw_lst.append(f'   {rule}')
                    else:
                        self.fw_lst.append(f'   :fire:{rule}')
                else:
                    r.logging.debug(f'Error: {result.stderr.decode()}')
                    self.fw_lst.append(f'[red]failed:[/red]{rule}')
            except Exception as e:
                print(f'An error occurred: {e}')
        return self.fw_lst

    def run_firewall_mitmproxy(self, rules):
        ''' Firewall Router rules - backup, flush and apply new rules for router to function '''
        self.mitm_lst = []
        for rule in rules:
            # Remove rule if toggle is false.
            if not self.mitmproxy_toggle:
                rule = rule.replace('-A', '-D')
            try:
                result = subprocess.run(rule, stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE, shell=True)             
                if result.returncode == 0:
                    if not self.mitmproxy_toggle:
                        self.mitm_lst.append(f'   {rule}')
                    else:
                        self.mitm_lst.append(f'   :fire:{rule}')
                else:
                    r.logging.debug(f'Error: {result.stderr.decode()}')
                    self.mitm_lst.append(f'[red]failed:[/red]{rule}')
            except Exception as e:
                print(f'An error occurred: {e}')
        return self.mitm_lst

    def check_package_installed(self, package):
        ''' Sets icon based on package installation status. '''
        try:
            result = subprocess.run(['dpkg-query', '-W', '-f=${Status}', package], 
                stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            if 'install ok installed' in result.stdout.decode():
                return f"[green]installed[/green] {package}"
            else:
                return f"[red]not found[/red] {package}"
        except Exception as e:
            return f"[red]failed[/red] {package}"

    def check_service_status(self, service):
        ''' Sets icon based on service status '''
        try:
            result = subprocess.run(['systemctl', 'is-active', f'{service}'], 
                stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            status = result.stdout.strip()
            if status == 'active':
                return f"[green]active[/green] {service}"
            elif status == 'inactive':
                return f"[red]inactive[/red] {service}"
            elif status == 'failed':
                return f"[red]failed[/red] {service}"
        except Exception as e:
            print(f" An error occured: {e}")

    def parse_dhcp_leases(self, file_path):
        ''' Parse DHCPv4 active leases '''
        leases = {}
        with open(file_path, 'r') as f1:
            lease_data = f1.read()

        lease_blocks = lease_data.split('lease ')
        for block in lease_blocks[1:]:
            ipv4_match = re.search(r'(\d+\.\d+.\d+.\d+)', block)
            mac_match = re.search(r'hardware ethernet ([0-9a-fA-F:]+);', block)
            if ipv4_match and mac_match:
                ipv4_address = ipv4_match.group(1)
                mac_address = mac_match.group(1)
                leases[ipv4_address] = mac_address
        return leases

    def parse_dhcpv6_leases(self, file_path):
        ''' Parse DHCPv6 active leases '''
        leases = {}
        with open(file_path, 'r') as f1:
            lease_data = f1.read()

        lease_blocks = lease_data.split('iaaddr')
        for block in lease_blocks[1:]:
            ipv6_match = re.search(r'([0-9a-fA-F:]+)', block)
            state_match = re.search(r'binding state active', block)
            if ipv6_match and state_match:
                ipv6_address = ipv6_match.group(1)
                leases[ipv6_address] = "Active"
        return leases

    def print_header(self) -> None:
        ''' Print Session Banner '''

        # Toggle clear screen function off when loglevel is DEBUG.
        if not self.loglevel_name == 'DEBUG':
            os.system("clear")

        # Whitespaces
        wp = 3*' '
        
        # Network Interfaces:
        int_lst = []
        for interface in self.interfaces.values():
            result = self.check_interface(interface)
            int_lst.append(f'{wp}{result}')
        list_text = r.Text("\n".join(int_lst))
        service_panel = r.Panel(f'Interfaces:\n{list_text}')
        r.console.print(service_panel)

        # Assigned IP Addresses:
        ip_lst = []
        addresses = psutil.net_if_addrs()
        for interface, addr_list in addresses.items():
            if interface in self.interfaces.values():
                for addr in addr_list:
                    if addr.family == socket.AF_INET:
                        ip_lst.append(f'{wp}{interface}: {addr.address}')
                    elif addr.family == socket.AF_INET6 and not addr.address.startswith('fe80::'):
                        ip_lst.append(f'{wp}{interface}: {addr.address}')
        list_text = r.Text("\n".join(ip_lst))
        service_panel = r.Panel(f'IP Addresses:\n{list_text}')
        r.console.print(service_panel)

        # Firewall:
        list_text = r.Text("\n".join(self.fw_lst))
        service_panel = r.Panel(f'Firewall Rules:\n{list_text}')
        r.console.print(service_panel)
        
        # Mitmproxy:
        list_text = r.Text("\n".join(self.mitm_lst))
        service_panel = r.Panel(f'Mitmproxy Rules:\n{list_text}')
        r.console.print(service_panel)

        # Packages:
        pkg_lst = []
        for package in self.packages:
            result = self.check_package_installed(package)
            pkg_lst.append(f' {wp}{result}')
        list_text = r.Text("\n".join(pkg_lst))
        service_panel = r.Panel(f'Packages:\n{list_text}')
        if self.packages_toggle:
            r.console.print(service_panel) 
        
        # Services:
        service_lst = []
        for service in self.services:
            result = self.check_service_status(service)
            if result is not None:
                service_lst.append(f'{wp}{result}')
        list_text = r.Text("\n".join(service_lst))
        service_panel = r.Panel(f'Services:\n{list_text}')
        r.console.print(service_panel)

        # DHCP Leases:
        dhcpd_leases_fp = '/var/lib/dhcp/dhcpd.leases'
        ipv4_leases = self.parse_dhcp_leases(dhcpd_leases_fp)
        leases_list = []
        for ipv4, mac in ipv4_leases.items():
            ipv4_mac = f' {wp}{ipv4} / {mac}'
            leases_list.append(ipv4_mac)
        
        dhcpdv6_leases_fp  = '/var/lib/dhcp/dhcpd6.leases'
        ipv6_leases = self.parse_dhcpv6_leases(dhcpdv6_leases_fp)
        for ipv6, status in ipv6_leases.items():
            leases_list.append(f' {wp}{ipv6}')
        list_text = r.Text("\n".join(leases_list))
        service_panel = r.Panel(f'DHCP Leases:\n{list_text}')
        r.console.print(service_panel)

def main():
    # Config Parser Objects
    config_obj = read_ini(config_ini_fp, delimiters='=')

    # Services List
    service_lst = ['networking.service', 'isc-dhcp-server', 'radvd', 'dnsmasq', 'hostapd']

    # Session Manager
    session = Session(config_obj, service_lst)

    # Session - Firewall rules.
    fw_router_rules = [v for v in config_obj['firewall-router'].values()]
    session.run_firewall_router(fw_router_rules)

    # ConfigParser - Packages List
    packages = [k for k in config_obj["packages"]]

    # Config Files - Time stamp for backups.
    ts = time_stamp()
    
    # Config Files - ignored sections
    sections_ignored = ['interface', 'firewall-router', 'firewall-mitmproxy', 'packages']

    # Config Files - Create Config Files list.
    config_files = []
    for section in config_obj.sections():
        if section not in sections_ignored:
            config_files.append(section)

    # Config Files - Backup.
    backups = False
    try:
        backups = []
        if config_files != None:
            for config_file in config_files:
                fp = config_obj[config_file]["fp"]
                results = backup_file(fp, ts)
                backups.append(results)
    except Exception as e:
        raise e
    else:
        backups = True

    # Write Config Files.
    wrote_config_files = False
    try:
       results = []
       for config_file in config_files:
           fp = config_obj[config_file]["fp"]
           option = config_obj[config_file]["option"]
           with open(fp, "w") as f1:
               f1.write(option)
           results.append(fp)
    except Exception as e:
        raise e
    else:
        wrote_config_files = True

    while True:
        ### Main-menu ###
        session.print_header()
        if backups:
            r.console.print(f'Backups: {bckup_dir}')
        main_menu = MainMenu(config_obj)
        selected_main_menu_item = main_menu.present_menu()
        r.logging.debug(f'Main Menu Option Selected: {selected_main_menu_item}')
        
        # PackageManager - Install
        if selected_main_menu_item == main_menu.opt_packages_install:
            pm = PackageManager()
            pm.update_cache()
            # Install required packages.
            try:
                [pm.install_package(package) for package in packages]
            except Exception as e:
                print(f"{e}")
            # Newly installed packages.
            if pm.new_packages:
                [print(f"Installed: {package}") for package in pm.new_packages]
            input("Press Enter to continue")
        # PackageManager - View
        if selected_main_menu_item == main_menu.opt_packages_view:
            session.packages_toggle = not session.packages_toggle

        # Services - Start
        elif selected_main_menu_item == main_menu.opt_services_start:
            for service in service_lst:
                if service == 'hostapd':
                    manage_service('unmask', service)

            for service in service_lst:
                manage_service('enable', service)

            for service in service_lst:
                manage_service('start', service)
            # reload sysctl
            command = 'sudo sysctl -p /etc/sysctl.conf'
            run_cmd(command)
            input("Press Enter to continue")
        # Services - Restart
        elif selected_main_menu_item == main_menu.opt_services_restart:
            for service in service_lst:
                manage_service('stop', service)
            for service in service_lst:
                manage_service('start', service)
            # reload sysctl
            command = 'sudo sysctl -p /etc/sysctl.conf'
            run_cmd(command)
            input("Press Enter to continue")
        elif selected_main_menu_item == main_menu.opt_firewall_toggle:
            session.firewall_toggle = not session.firewall_toggle
            # Firewall rules.
            fw_router_rules = [v for v in config_obj['firewall-router'].values()]
            session.run_firewall_router(fw_router_rules)
        elif selected_main_menu_item == main_menu.opt_firewall_mitmproxy_toggle:
            session.mitmproxy_toggle = not session.mitmproxy_toggle
            # Firewall Mitmproxy rules.
            fw_mitmproxy_rules = [v for v in config_obj['firewall-mitmproxy'].values()]
            session.run_firewall_mitmproxy(fw_mitmproxy_rules)
        # Exit
        elif selected_main_menu_item == main_menu.opt_exit:
            session.firewall_toggle = False
            session.mitmproxy_toggle = False
            # Firewall rules - Revert before exting.
            fw_router_rules = [v for v in config_obj['firewall-router'].values()]
            session.run_firewall_router(fw_router_rules)
            # Firewall Mitmproxy rules - Revert before exting.
            fw_mitmproxy_rules = [v for v in config_obj['firewall-mitmproxy'].values()]
            session.run_firewall_mitmproxy(fw_mitmproxy_rules)
            sys.exit()
            sys.exit()

if __name__ == '__main__':
    main()