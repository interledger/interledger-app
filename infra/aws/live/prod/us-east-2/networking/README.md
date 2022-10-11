# AWS Networking Dev EU-WEST-1

This project creates the main networking stack for the Prod account in `US-EAST-2` using the CIDR block `10.30.0.0/16`.
A single NAT gateway is configured. Below it the VPC setup


| Type    | AZ         | Subnet         |
|---------|------------|----------------|
| Public  | eu-west-1a | 10.30.128.0/20 |
| Public  | eu-west-1b | 10.30.144.0/20 |
| Public  | eu-west-1c | 10.30.160.0/20 |
| Private | eu-west-1a | 10.30.0.0/19   |
| Private | eu-west-1b | 10.30.32.0/19  |
| Private | eu-west-1c | 10.30.64.0/19  |
| Intra   | eu-west-1a | 10.30.192.0/20 |
| Intra   | eu-west-1b | 10.30.208.0/20 |
| Intra   | eu-west-1c | 10.30.224.0/20 |

