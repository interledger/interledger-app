# AWS Networking Dev EU-WEST-1

This project creates the main networking stack for the Dev account in `EU-WEST-1` using the CIDR block `10.10.0.0/16`.
A single NAT gateway is configured. Below it the VPC setup


| Type    | AZ         | Subnet         |
|---------|------------|----------------|
| Public  | eu-west-1a | 10.10.128.0/20 |
| Public  | eu-west-1b | 10.10.144.0/20 |
| Public  | eu-west-1c | 10.10.160.0/20 |
| Private | eu-west-1a | 10.10.0.0/19   |
| Private | eu-west-1b | 10.10.32.0/19  |
| Private | eu-west-1c | 10.10.64.0/19  |
| Intra   | eu-west-1a | 10.10.192.0/20 |
| Intra   | eu-west-1b | 10.10.208.0/20 |
| Intra   | eu-west-1c | 10.10.224.0/20 |

