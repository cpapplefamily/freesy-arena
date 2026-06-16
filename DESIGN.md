# freesy-arena — Fork Design Notes

**Status:** Living document (fork notes)
**Last updated:** 2026-05-25
**Upstream:** `cpapplefamily/freesy-arena` (itself a fork of [Team254/cheesy-arena](https://github.com/Team254/cheesy-arena), the official Cheesy Arena FMS)
**Active branch:** `Freezy-Arena-2025`
**README:** retains upstream Cheesy Arena content; this document captures what *this fork* changes.

This document records the design intent of the freesy-arena fork — what it adds on top of upstream Cheesy Arena and why. Update it whenever a new fork-specific feature lands or an upstream feature is reverted/disabled.

## Why fork?

Off-season and small-venue MN FRC operations needed capabilities upstream Cheesy Arena doesn't provide out of the box — specifically:

- Small-bracket double-elimination playoffs (smaller than the official 8-team)
- Switching from FMS-managed Cisco switches to UniFi-managed switches
- Local DHCP for team networks (vs. upstream's static-IP-only assumption)
- Operational ergonomics (FTA monitor, quick-keys)

Forking with upstream's permission lets these changes ship without round-trips through Team254 review.

## Fork-specific features (on top of upstream Cheesy Arena)

### Tournament

- **3-8 team double eliminations** — upstream supports 8-team only; this fork adds 3, 4, 5, 6, 7-team double-elimination bracket generators in `playoff/`.

### Operator UX

- **FTA Monitor (Alliance)** — color-coded field-status interface for the Field Technical Advisor; auto-switches to the Game Score display during a match.
- **Quick keys** — keyboard shortcuts for common team / match control actions.

### Networking — UniFi switch path

- **`network/unifiswitch.go`** — UniFi switch driver as an alternative to upstream's Cisco-only `sccswitch.go`. Provisioned via SSH/REST against the UniFi controller.
- **`unifi_cycle_switch_ports.yaml`** — Ansible playbook that bounces a team's switch port to force a DHCP-lease refresh (workaround for sticky-MAC issues when teams swap robots between matches).
- **`create_vlans.yaml`** — Ansible playbook that provisions per-team VLANs on the UniFi controller (analogue to the Cisco VLAN-create step upstream).
- **Cisco switch architecture fully disabled** — `sccswitch.go` code paths are not exercised in fork builds; the build is UniFi-first.

### Networking — local DHCP

- **`config_dhcp.yaml`** — Ansible playbook to configure a local DHCP server on the FMS host so team driver-station laptops get addresses without the upstream's static-IP convention.
- **Filter plugins** under `filter_plugins/` handle null/empty team-number values without erroring (frequent in early-event scheduling when not all teams have arrived).

### Operational

- **`configure_cheesy_sudo.yaml`** — installs a sudoers drop-in that lets the `cheesy` service user run the network/DHCP playbooks without password prompts (so they can be invoked from the running Cheesy Arena process when needed).

## Branch model

- `main` — tracking upstream Cheesy Arena.
- `Works` — historical "known-good" branch used during early fork stabilization. Most fork-specific code originated here.
- `Freezy-Arena-2025` — **active fork branch**, contains the 2025 Reefscape upstream + all fork-specific features ported on top. This is what runs in production at MN off-season events.
- `remotes/upstream/*` — periodically pulled to track upstream feature releases (e.g. `2025-Reefscape-Week-0`, `2-min-head-referee-timer`, `API-for-Auto-Scoring`).

## Design choices

- **Fork rather than upstream PR.** UniFi support and local DHCP are MN-specific operational needs; upstream serves the larger FRC community where Cisco + static IPs is the convention. Keeping these in a fork avoids putting upstream maintainers on the hook for hardware they don't have.
- **Ansible for network provisioning, not Go.** Switch and DHCP configuration is delegated to Ansible playbooks invoked from the Go arena process. Easier to debug at-event (run the playbook by hand) and easier for the next operator to extend without recompiling.
- **Keep upstream README authoritative for setup/operation.** Operators read the upstream README to install and use Cheesy Arena, and refer here only to understand what's different in this fork.

## Known limitations / future work

- **No CI for fork-specific code paths.** Upstream test suite runs; UniFi driver and Ansible playbooks have no automated coverage. Adding `network/unifiswitch_test.go` parity with `sccswitch_test.go` is the obvious next step.
- **Documentation gap** — the fork-specific features above are not surfaced in the upstream-derived README. End users find them by reading the bracket-size dropdown or the FTA monitor UI directly. A `docs/fork-features.md` walkthrough with screenshots would help operators new to the fork.
- **Upstream merge cadence** — currently periodic (`Merge pull request #1/#2 from joshzcold/master`-style flow). Decide whether to track every upstream release or only ones that affect Reefscape rules / scoring.
- **3-team and 5-team / 7-team brackets** — verify the bracket-generation correctness with a structured test rather than the spot-check used during initial dev.
- **UniFi controller credential handling** — currently in Ansible vars. Migrate to a secrets store (ansible-vault or HashiCorp Vault) before the fork is used at events beyond MN off-season.
- **Cisco switch code** — disabled but still present. Decide whether to remove it entirely (smaller surface, cleaner reads) or keep it for future option-value (Cisco hardware comes back).
