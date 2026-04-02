package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ============================================================================
// Module: Access Control (Zero Trust Layer 2)
// Policy: RBAC ??Role ??Region ??{read, write, deny}
// Source: .neuronfs/access_policy.yaml
// ============================================================================

// AccessAction represents a permitted action type.
type AccessAction string

const (
	ActionRead  AccessAction = "read"
	ActionWrite AccessAction = "write"
	ActionDeny  AccessAction = "deny"
)

// RegionPolicy defines permissions for a single brain region.
type RegionPolicy struct {
	Actions []string `yaml:"actions"` // ["read", "write"] or ["deny"]
}

// RolePolicy binds a role name to its region permissions.
type RolePolicy struct {
	Regions map[string]RegionPolicy `yaml:"regions"`
}

// AccessPolicy is the top-level policy document.
type AccessPolicy struct {
	Version string                `yaml:"version"`
	Roles   map[string]RolePolicy `yaml:"roles"`
}

var (
	errPolicyNotLoaded = errors.New("access_control: policy not loaded")
	errRoleNotFound    = errors.New("access_control: role not defined in policy")
	errRegionNotFound  = errors.New("access_control: region not defined for role")
	errInvalidAction   = errors.New("access_control: invalid action (use read|write)")
)

// validRegions are the 7 brain regions in NeuronFS.
var validRegions = map[string]bool{
	"brainstem":    true,
	"limbic":       true,
	"hippocampus":  true,
	"sensors":      true,
	"cortex":       true,
	"ego":          true,
	"prefrontal":   true,
}

// globalPolicy holds the loaded policy (module-level singleton).
var globalPolicy *AccessPolicy

// LoadAccessPolicy reads and parses the YAML policy file.
func LoadAccessPolicy(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("access_control: failed to read policy: %w", err)
	}
	var policy AccessPolicy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return fmt.Errorf("access_control: failed to parse policy: %w", err)
	}
	globalPolicy = &policy
	return nil
}

// LoadAccessPolicyFromBytes parses policy from raw YAML bytes (for testing).
func LoadAccessPolicyFromBytes(data []byte) error {
	var policy AccessPolicy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return fmt.Errorf("access_control: failed to parse policy: %w", err)
	}
	globalPolicy = &policy
	return nil
}

// CanAccess checks whether a given role can perform an action on a region.
// Returns (allowed bool, error).
// Deny is explicit ??if "deny" appears in actions, access is refused.
// If a region is not listed for the role, default is DENY.
func CanAccess(role string, region string, action string) (bool, error) {
	if globalPolicy == nil {
		return false, errPolicyNotLoaded
	}

	// Normalize inputs
	role = strings.TrimSpace(strings.ToLower(role))
	region = strings.TrimSpace(strings.ToLower(region))
	action = strings.TrimSpace(strings.ToLower(action))

	if action != string(ActionRead) && action != string(ActionWrite) {
		return false, errInvalidAction
	}

	rolePol, ok := globalPolicy.Roles[role]
	if !ok {
		return false, errRoleNotFound
	}

	regPol, ok := rolePol.Regions[region]
	if !ok {
		// Default deny for unlisted regions (Zero Trust principle)
		return false, nil
	}

	// Check for explicit deny
	for _, a := range regPol.Actions {
		if strings.ToLower(strings.TrimSpace(a)) == string(ActionDeny) {
			return false, nil
		}
	}

	// Check if requested action is permitted
	for _, a := range regPol.Actions {
		if strings.ToLower(strings.TrimSpace(a)) == action {
			return true, nil
		}
	}

	return false, nil
}

// ListRolePermissions returns a human-readable summary of a role's access map.
func ListRolePermissions(role string) (string, error) {
	if globalPolicy == nil {
		return "", errPolicyNotLoaded
	}

	role = strings.TrimSpace(strings.ToLower(role))
	rolePol, ok := globalPolicy.Roles[role]
	if !ok {
		return "", errRoleNotFound
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Role: %s\n", role))
	for region, pol := range rolePol.Regions {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", region, strings.Join(pol.Actions, ", ")))
	}
	return sb.String(), nil
}

