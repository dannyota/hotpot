package main

import "github.com/dannyota/hotpot/pkg/ingest"

//go:generate go run github.com/dannyota/hotpot/tools/ingestgen

var _ = ingest.ProviderSet("gcp", "greennode", "sentinelone")
var _ = ingest.DisableServiceSet("greennode", "dns", "glb", "loadbalancer")
