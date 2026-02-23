package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	computev2 "danny.vn/greennode/services/compute/v2"
	portalv1 "danny.vn/greennode/services/portal/v1"

	"github.com/dannyota/hotpot/pkg/base/config"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	configService := config.NewService(config.ServiceOptions{
		Source: config.NewFileSource("config.yaml"),
	})
	if err := configService.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	iamAuth := &auth.IAMUserAuth{
		RootEmail: configService.GreenNodeRootEmail(),
		Username:  configService.GreenNodeUsername(),
		Password:  configService.GreenNodePassword(),
		TOTP:      &auth.SecretTOTP{Secret: configService.GreenNodeTOTPSecret()},
	}

	regions := configService.GreenNodeRegions()
	if len(regions) == 0 {
		fmt.Println("No regions configured")
		os.Exit(0)
	}
	region := regions[0]

	// --- Test 1: Login ---
	fmt.Printf("=== Test 1: Login (username=%s) ===\n", iamAuth.Username)
	start := time.Now()
	token, expiresAt, err := iamAuth.Authenticate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK (%v) token=%s... expires=%v\n\n", time.Since(start), token[:min(20, len(token))], time.Unix(0, expiresAt))

	// --- Test 2: ListProjects ---
	fmt.Printf("=== Test 2: ListProjects (region=%s) ===\n", region)
	start = time.Now()
	sdkNoProject, err := greennode.NewClient(ctx, greennode.Config{
		Region:  region,
		IAMAuth: iamAuth,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: create client: %v\n", err)
		os.Exit(1)
	}

	projects, err := sdkNoProject.PortalV1.ListProjects(ctx, portalv1.NewListProjectsRequest())
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK (%v) %d projects:\n", time.Since(start), len(projects.Items))
	if len(projects.Items) == 0 {
		fmt.Println("  (none — cannot continue without a project)")
		os.Exit(0)
	}
	projectID := projects.Items[0].ProjectID
	for _, p := range projects.Items {
		fmt.Printf("  - %s\n", p.ProjectID)
	}
	fmt.Printf("Using project: %s\n\n", projectID)

	// Create SDK client with project ID for remaining tests
	sdk, err := greennode.NewClient(ctx, greennode.Config{
		Region:    region,
		ProjectID: projectID,
		IAMAuth:   iamAuth,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: create client with project: %v\n", err)
		os.Exit(1)
	}

	// --- Test 3: ListServers ---
	fmt.Printf("=== Test 3: ListServers ===\n")
	start = time.Now()
	servers, err := sdk.Compute.ListServers(ctx, computev2.NewListServersRequest(1, 50))
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
	} else {
		fmt.Printf("OK (%v) %d servers:\n", time.Since(start), len(servers.Items))
		for _, s := range servers.Items {
			fmt.Printf("  - %s (name=%s, status=%s)\n", s.Uuid, s.Name, s.Status)
		}
	}
	fmt.Println()

	// --- Test 4: ListSSHKeys ---
	fmt.Printf("=== Test 4: ListSSHKeys ===\n")
	start = time.Now()
	sshKeys, err := sdk.Compute.ListSSHKeys(ctx, computev2.NewListSSHKeysRequest(1, 50))
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
	} else {
		fmt.Printf("OK (%v) %d SSH keys:\n", time.Since(start), len(sshKeys.Items))
		for _, k := range sshKeys.Items {
			fmt.Printf("  - %s (name=%s, status=%s)\n", k.ID, k.Name, k.Status)
		}
	}
	fmt.Println()

	// --- Test 5: ListServerGroups ---
	fmt.Printf("=== Test 5: ListServerGroups ===\n")
	start = time.Now()
	serverGroups, err := sdk.Compute.ListServerGroups(ctx, computev2.NewListServerGroupsRequest(1, 50))
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
	} else {
		fmt.Printf("OK (%v) %d server groups:\n", time.Since(start), len(serverGroups.Items))
		for _, g := range serverGroups.Items {
			fmt.Printf("  - %s (name=%s, policy=%s)\n", g.UUID, g.Name, g.PolicyName)
		}
	}
	fmt.Println()

	// --- Test 6: ListRegions ---
	fmt.Printf("=== Test 6: ListRegions ===\n")
	start = time.Now()
	regionList, err := sdk.Portal.ListRegions(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
	} else {
		fmt.Printf("OK (%v) %d regions:\n", time.Since(start), len(regionList.Items))
		for _, r := range regionList.Items {
			fmt.Printf("  - %s (name=%s, desc=%s)\n", r.ID, r.Name, r.Description)
		}
	}
	fmt.Println()

	// --- Test 7: ListAllQuotaUsed ---
	fmt.Printf("=== Test 7: ListAllQuotaUsed ===\n")
	start = time.Now()
	quotas, err := sdk.Portal.ListAllQuotaUsed(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
	} else {
		fmt.Printf("OK (%v) %d quotas:\n", time.Since(start), len(quotas.Items))
		for _, q := range quotas.Items {
			fmt.Printf("  - %s: %d/%d (%s)\n", q.Name, q.Used, q.Limit, q.Type)
		}
	}
	fmt.Println()

	fmt.Println("=== All tests complete ===")
}
