# AWS Networking Shared EU-WEST-1

This project creates the main networking stack for the Shared account in `EU-WEST-1` using the CIDR block `10.100.0.0/16`.
A single NAT gateway is configured. Below it the VPC setup


| Type    | AZ         | Subnet          |
|---------|------------|-----------------|
| Public  | eu-west-1a | 10.100.128.0/20 |
| Public  | eu-west-1b | 10.100.144.0/20 |
| Public  | eu-west-1c | 10.100.160.0/20 |
| Private | eu-west-1a | 10.100.0.0/19   |
| Private | eu-west-1b | 10.100.32.0/19  |
| Private | eu-west-1c | 10.100.64.0/19  |
| Intra   | eu-west-1a | 10.100.192.0/20 |
| Intra   | eu-west-1b | 10.100.208.0/20 |
| Intra   | eu-west-1c | 10.100.224.0/20 |

