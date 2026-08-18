## Context

I started this project to learn two tools that caught my attention: Pulumi and SaltStack. I've been using Terraform and Ansible for a while and wanted to explore other options. Pulumi lets you define infrastructure in a real programming language (Go, in my case), while Salt is a configuration-management tool that I kept hearing about but hadn't used yet.

I run a production OpenNebula cloud on OVH, and I wanted to see how these two tools work in a real setup rather than a throwaway lab.

On OpenNebula I'm deploying an internal Keycloak instance for authentication and authorization, meant for potential use across our internal applications, with a PostgreSQL database on a separate VM to store the Keycloak data.

Since these workloads sit behind an OPNsense firewall with a public IP (that I want to keep private), I'm also deploying an Nginx reverse proxy on Azure (Edge Proxy) to take the incoming traffic and route it to Keycloak. The proxy reaches back to the internal Keycloak through a WireGuard VPN tunnel (managed by Salt). It also handles SSL termination, so users get a secure connection while Keycloak itself never gets exposed to the public internet.

## Infrastructure Architecture

```
            Internet
               │ https (443)
        ┌──────▼───────┐
        │  edge-az-01  │  Azure — nginx TLS reverse proxy (Salt role: edge)
        └──────┬───────┘
               │ WireGuard tunnel (Salt-managed)
 ══════════════╪══════════════  OPNsense: WG + Salt 4505/4506, source-restricted
               │
   OpenNebula / KVM (OVH)
   ┌───────────▼──────────┐     ┌────────────────────┐
   │  app-01 · Keycloak   │────▶│ db-01 · PostgreSQL │
   └──────────────────────┘     └────────────────────┘
   ┌──────────────────────────────────────────────────┐
   │ ctl-01 · Salt master + reactor + Pulumi program  │  (the one hand-made VM)
   └──────────────────────────────────────────────────┘
```

### Provisioning: Pulumi

The Pulumi program instantiates two VMs on OpenNebula from a template (`Ubuntu-20.04-Base`), one for Keycloak, one for PostgreSQL, and one VM on Azure, where it also creates the resource group and virtual network.

### Bootstrap (important step)

When a VM first boots, a bootstrap script installs the Salt minion, tags the VM with its role, and points it at the Salt master so the two can talk over the secure bus. This is the handoff: Pulumi's last job is injecting the bootstrap, and Salt takes over from there, Pulumi never touches the VM again.

### Configuration: SaltStack

The Salt master accepts the minion connections and manages the Keycloak and PostgreSQL VMs from there. Salt also manages the WireGuard tunnel between the Azure Nginx proxy and the internal network, keeping the traffic between the two environments encrypted.

### Control Plane

This VM is created manually, as a one-time setup, and hosts the Salt master, the Pulumi program, and the Salt reactor. It's manual on purpose: the tool can't provision the machine it runs on. The reactor automates actions based on events happening across the infrastructure, for example, re-applying a config when someone changes a watched file.

### Engineering Decisions

OpenNebula has no Pulumi provider, so I bridge the community Terraform provider: Pulumi reads its schema and generates a typed Go SDK from it. No Terraform binary and no HCL end up in the workflow, I just import the generated SDK in my Go program. Usually I run ``ònetemplate instantiate`` manually with the OpenNebula Frontend's XML-RPC API, now The Pulumi provider will do the same by connecting to OpenNebula 's XML-RPC API directly (https://<frontend>:2633/RPC2)

I put the reverse proxy on Azure to keep Keycloak and its database isolated from the public internet. Only the stateless proxy is exposed; the identity and data layers stay on the private cloud.

Secrets are handled in two places. Cloud credentials (the OpenNebula API user/password, etc.) live in Pulumi's stack config, encrypted at rest so the config file is safe to commit. In-guest secrets (the database password, Keycloak admin, WireGuard keys) live in Salt's GPG-encrypted pillar, the master decrypts them and hands each minion only its own, so nothing sensitive is ever stored in git in plaintext.

Split master address for a hybrid setup. The Salt minions don't all reach the master the same way. db-01 and app-01 sit on the same internal vnet as the control plane, so they reach the master directly on its private address (10.0.0.19). The Azure edge is outside that network, so it reaches the master through the OPNsense public IP, where ports 4505/4506 are forwarded and source-restricted to the edge's IP. I kept both values as separate Pulumi config keys (masterAddrInternal / masterAddrPublic) and the bootstrap picks the right one per VM at provision time.

## SetUp details

#### Step1: Ctl-01 (Salt master + Pulumi program + Salt reactor)

Instantiated a VM from the `Ubuntu-20.04-Base` template (since Pulumi can't provision the machine it runs on), with 2 vCPU, 4GB RAM, 50GB disk. Pulumi and Salt are installed on it, and the Pulumi program is cloned from this repo.

Pulumi Backend choice: I use the local backend, which stores the state file in ~/.pulumi. 

Note: I access the Ctl-01 VM  which is behind the OPNsense firewall through SSH with PKI (OPNsense translates the public key to an internal Bastion VM through highport NAT, so from that node I ssh to the Ctl-01 on the local vnet). 
ssh -A <high-port>  adel@<public-ip>

#### Step 2: OpenNebula prep

Created a dedicated OpenNebula user for Pulumi instead of using oneadmin (never), with USE rights only on the `Ubuntu-24.04-Base` template, its image, and the vnet. Automation gets its own least-privilege identity, so anything Pulumi does is traceable to that user and can't touch the rest of the cloud.

Exported the template for reference so the repo is self-contained:
`onetemplate show -x <id> > templates/ubuntu-24.04-base.yaml`

The OpenNebula credentials for this user are stored as Pulumi secret config (`pulumi config set --secret`).

#### Step 3: Pulumi project scaffold and OpenNebula SDK generation

In this would sset up Pulumi config locally and hood the OpenNebula Terraform provider to generate a Go SDK. The Pulumi config includes the OpenNebula API endpoint, the dedicated user credentials, and the Azure credentials. Setting the Master addresses (Internal/Public).




