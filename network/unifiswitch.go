// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Cisco Switch 3500-series switch for team VLANs.

package network

import (
	"encoding/json"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"
)

const (
	// This are not used yet, but may in the future if we tap into the API vs ansible
	UnifiSwitchConfigBackoffDurationSec = 5
	UnifiSwitchConfigPauseDurationSec   = 2
	UnifiSwitchTeamGatewayAddress       = 4
	UnifiSwitchAPIPort                  = 9999
)

// Don't need these for now, ansible has these defined in it already.
// const (
// 	red1Vlan  = 10
// 	red2Vlan  = 20
// 	red3Vlan  = 30
// 	blue1Vlan = 40
// 	blue2Vlan = 50
// 	blue3Vlan = 60
// )

type UnifiSwitch struct {
	address               string
	port                  int
	password              string
	mutex                 sync.Mutex
	configBackoffDuration time.Duration
	configPauseDuration   time.Duration
	Status                string
}

func NewUnifiSwitch(address, password string) *UnifiSwitch {
	return &UnifiSwitch{
		address:               address,
		port:                  UnifiSwitchAPIPort,
		password:              password,
		configBackoffDuration: UnifiSwitchConfigBackoffDurationSec * time.Second,
		configPauseDuration:   UnifiSwitchConfigPauseDurationSec * time.Second,
		Status:                "UNKNOWN",
	}
}

// Sets up wired networks for the given set of teams.
// Helper function to run commands safely using os/exec and capture output on failure.
func (sw *UnifiSwitch) runExecCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	fmt.Printf("Executing command: %s\n", cmd.Args) // Log the command

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("command failed to start: %w", err)
	}

	// Read output in separate goroutines to prevent deadlocks
	stdoutBytes, errStdout := io.ReadAll(stdout)
	stderrBytes, errStderr := io.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		// Log the error detail and the captured output
		log.Printf("Command failed: %s\nError: %v\n", name, err)

		// Check if reading output pipes had errors, log them as well
		if errStdout != nil {
			log.Printf("Error reading stdout: %v\n", errStdout)
		} else if len(stdoutBytes) > 0 {
			log.Printf("Captured STDOUT:\n%s\n", string(stdoutBytes))
		}
		if errStderr != nil {
			log.Printf("Error reading stderr: %v\n", errStderr)
		} else if len(stderrBytes) > 0 {
			log.Printf("Captured STDERR:\n%s\n", string(stderrBytes))
		}

		// Return a comprehensive error message including standard output/error streams
		return fmt.Errorf("command execution failed: %w. STDOUT: %s. STDERR: %s", err, string(stdoutBytes), string(stderrBytes))
	}

	// Optionally log successful output if needed (commented out by default)
	// if len(stdoutBytes) > 0 {
	//     log.Printf("Command STDOUT:\n%s\n", string(stdoutBytes))
	// }
	// if len(stderrBytes) > 0 {
	//     log.Printf("Command STDERR:\n%s\n", string(stderrBytes))
	// }

	return nil
}

func (sw *UnifiSwitch) UnifiConfigureTeamEthernet(teams [6]*model.Team) error {
	// Make sure multiple configurations aren't being set at the same time.
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	sw.Status = "CONFIGURING"

	// --- Step 1: Remove old team VLANs (cleaner command execution) ---
	err := sw.runExecCommand("ansible-playbook", "create_vlans.yaml")
	if err != nil {
		sw.Status = "ERROR"
		return err
	}
	time.Sleep(sw.configPauseDuration)

	// --- Step 2: Create the new team VLANs (safe JSON and exec) ---

	// Prepare data structure for JSON marshaling
	type AnsibleVars struct {
		// TeamNumbers will always contain 6 entries. An empty string indicates a null or empty position.
		TeamNumbers [6]string `json:"team_numbers"`
	}

	var vars AnsibleVars

	// Populate the fixed-size array, using an empty string for nil entries to maintain position.
	hasTeams := false
	for i, team := range teams {
		if team != nil {
			vars.TeamNumbers[i] = fmt.Sprintf("%d", team.Id)
			hasTeams = true
		} else {
			vars.TeamNumbers[i] = "" // Use empty string to represent an empty position.
		}
	}

	if !hasTeams {
		log.Println("No valid team IDs provided; skipping 'config_dhcp.yaml' playbook as all entries are empty.")
		sw.Status = "ACTIVE"
		return nil
	}

	// Safely marshal the Go struct into a JSON byte slice
	jsonData, err := json.Marshal(vars)
	if err != nil {
		sw.Status = "ERROR"
		return fmt.Errorf("failed to marshal JSON for ansible vars: %w", err)
	}

	// Pass the raw JSON string as a single argument for the -e flag
	extraVarsArg := string(jsonData)

	// Execute the command safely, passing each argument as a separate element in the list
	err = sw.runExecCommand("ansible-playbook", "-e", extraVarsArg, "config_dhcp.yaml")
	if err != nil {
		sw.Status = "ERROR"
		return err
	}

	// Give some time for the configuration to take before another one can be attempted.
	time.Sleep(sw.configBackoffDuration)

	sw.Status = "ACTIVE"
	return nil
}
