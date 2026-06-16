#!/usr/bin/python
# Custom Ansible Filter to implement the subnet generation logic

from ansible.errors import AnsibleFilterError
# Import ipaddress module for robust IP handling in Python
import ipaddress 

def generate_subnet_details_filter(team_number_input):
    """
    Generates a dictionary of subnet details based on team number rules.
    Returns None if the input is null, empty, or invalid, allowing skips.
    """
    # Check for null, empty string, or non-numeric input immediately
    if team_number_input is None or (isinstance(team_number_input, str) and not team_number_input.strip().isdigit()):
        return None # Indicate that this entry should be skipped

    try:
        team_number = int(team_number_input)
        team_num_str = str(team_number)
        
        if not 1 <= len(team_num_str) <= 5:
            raise ValueError(f"Team number {team_number} has an invalid length.")

        second_octet = 0
        third_octet = 0

        if len(team_num_str) <= 2:
            # For 1 or 2 digit teams, place the whole number in the third octet
            third_octet = team_number
        elif len(team_num_str) == 3:
            # For 3 digit teams: first digit is second octet, last two are third octet
            second_octet = int(team_num_str[0])
            third_octet = int(team_num_str[1:])
        elif len(team_num_str) == 4:
            second_octet = int(team_num_str[:2])
            third_octet = int(team_num_str[2:])
        elif len(team_num_str) == 5:
            second_octet = int(team_num_str[:3])
            third_octet = int(team_num_str[3:])

        network_cidr = f"10.{second_octet}.{third_octet}.0/24"
        network_obj = ipaddress.IPv4Network(network_cidr)
        network_address = str(network_obj.network_address)
        
        broadcast_ip = str(network_obj.broadcast_address)
        first_assignable_ip = str(network_obj.network_address + 20)
        last_assignable_ip = str(network_obj.network_address + 199)

        return {
            'network': network_address,
            'subnet_cidr': network_cidr,
            'base_ip_prefix': f"10.{second_octet}.{third_octet}",
            'cidr_prefix': 24,
            'broadcast_ip': broadcast_ip,
            'first_assignable_ip': first_assignable_ip,
            'last_assignable_ip': last_assignable_ip,
            'is_valid': True # Add a flag for explicit checking in Jinja
        }

    except (ValueError, TypeError, ipaddress.AddressValueError) as e:
        # If processing fails for other reasons, return None to skip it safely
        return None


class FilterModule(object):
    def filters(self):
        return {
            'generate_subnet_details_filter': generate_subnet_details_filter
        }
