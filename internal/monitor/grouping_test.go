package monitor

import "testing"

func TestBuildComposeGroupsUsesComposeLabels(t *testing.T) {
	states := map[string]*ContainerState{
		"id1": {
			ContainerID:    "id1",
			ContainerName:  "web-1",
			ComposeProject: "shop",
			ComposeService: "web",
			NetworkNames:   []string{"shop_default"},
		},
		"id2": {
			ContainerID:    "id2",
			ContainerName:  "web-2",
			ComposeProject: "shop",
			ComposeService: "web",
			NetworkNames:   []string{"shop_default"},
		},
	}

	groups := buildComposeGroups(states)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	g := groups[0]
	if g.Project != "shop" || g.Service != "web" {
		t.Fatalf("unexpected group key: %s/%s", g.Project, g.Service)
	}
	if len(g.ContainerIDs) != 2 {
		t.Fatalf("expected 2 container ids, got %d", len(g.ContainerIDs))
	}
	if len(g.NetworkNames) != 1 || g.NetworkNames[0] != "shop_default" {
		t.Fatalf("unexpected network names: %#v", g.NetworkNames)
	}
}

func TestBuildComposeGroupsFallsBackToNonSystemNetwork(t *testing.T) {
	states := map[string]*ContainerState{
		"id1": {
			ContainerID:   "id1",
			ContainerName: "api-1",
			NetworkNames:  []string{"bridge", "custom_net"},
		},
	}

	groups := buildComposeGroups(states)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	g := groups[0]
	if g.Project != "network:custom_net" {
		t.Fatalf("expected network fallback project, got %s", g.Project)
	}
	if g.Service != "(heuristic)" {
		t.Fatalf("expected heuristic service, got %s", g.Service)
	}
}

func TestBuildComposeGroupsKeepsUngroupedWhenOnlySystemNetworks(t *testing.T) {
	states := map[string]*ContainerState{
		"id1": {
			ContainerID:   "id1",
			ContainerName: "job-1",
			NetworkNames:  []string{"bridge", "host", "none"},
		},
	}

	groups := buildComposeGroups(states)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	g := groups[0]
	if g.Project != "(ungrouped)" {
		t.Fatalf("expected ungrouped project, got %s", g.Project)
	}
	if g.Service != "(service)" {
		t.Fatalf("expected default service, got %s", g.Service)
	}
}
