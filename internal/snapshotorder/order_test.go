package snapshotorder

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	pb "docksphinx/api/docksphinx/v1"
)

func TestLessContainerInfo(t *testing.T) {
	items := []*pb.ContainerInfo{
		{ContainerName: "same", ContainerId: "id-b"},
		{ContainerName: "same", ContainerId: "id-a"},
		{ContainerName: "alpha", ContainerId: "id-z"},
	}
	sort.Slice(items, func(i, j int) bool { return LessContainerInfo(items[i], items[j]) })
	got := []string{
		items[0].GetContainerName() + "/" + items[0].GetContainerId(),
		items[1].GetContainerName() + "/" + items[1].GetContainerId(),
		items[2].GetContainerName() + "/" + items[2].GetContainerId(),
	}
	want := []string{"alpha/id-z", "same/id-a", "same/id-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected order: got=%v want=%v", got, want)
	}
}

func TestLessComposeGroup(t *testing.T) {
	items := []*pb.ComposeGroup{
		{Project: "same", Service: "svc", ContainerIds: []string{"b", "a"}, NetworkNames: []string{"n2", "n1"}},
		{Project: "same", Service: "svc", ContainerIds: []string{"a"}, NetworkNames: []string{"n2", "n1"}},
		{Project: "alpha", Service: "svc", ContainerIds: []string{"z"}, NetworkNames: []string{"n9"}},
	}
	sort.Slice(items, func(i, j int) bool { return LessComposeGroup(items[i], items[j]) })
	got := []string{
		items[0].GetProject() + "/" + items[0].GetService() + ":" + join(items[0].GetContainerIds()) + ":" + join(items[0].GetNetworkNames()),
		items[1].GetProject() + "/" + items[1].GetService() + ":" + join(items[1].GetContainerIds()) + ":" + join(items[1].GetNetworkNames()),
		items[2].GetProject() + "/" + items[2].GetService() + ":" + join(items[2].GetContainerIds()) + ":" + join(items[2].GetNetworkNames()),
	}
	want := []string{
		"alpha/svc:z:n9",
		"same/svc:a:n2,n1",
		"same/svc:b,a:n2,n1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected order: got=%v want=%v", got, want)
	}
}

func TestLessNetworkInfo(t *testing.T) {
	items := []*pb.NetworkInfo{
		{Name: "same", Driver: "bridge", Scope: "local", NetworkId: "n2"},
		{Name: "same", Driver: "bridge", Scope: "local", NetworkId: "n1"},
		{Name: "alpha", Driver: "bridge", Scope: "local", NetworkId: "n0"},
	}
	sort.Slice(items, func(i, j int) bool { return LessNetworkInfo(items[i], items[j]) })
	got := []string{
		items[0].GetName() + "/" + items[0].GetNetworkId(),
		items[1].GetName() + "/" + items[1].GetNetworkId(),
		items[2].GetName() + "/" + items[2].GetNetworkId(),
	}
	want := []string{"alpha/n0", "same/n1", "same/n2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected order: got=%v want=%v", got, want)
	}
}

func TestLessVolumeInfo(t *testing.T) {
	items := []*pb.VolumeInfo{
		{Name: "same", Driver: "local", Mountpoint: "/m", UsageNote: "note", RefCount: 2},
		{Name: "same", Driver: "local", Mountpoint: "/m", UsageNote: "note", RefCount: 1},
		{Name: "alpha", Driver: "local", Mountpoint: "/a", UsageNote: "note", RefCount: 9},
	}
	sort.Slice(items, func(i, j int) bool { return LessVolumeInfo(items[i], items[j]) })
	got := []string{
		fmt.Sprintf("%s/%s/%s/%s/%d", items[0].GetName(), items[0].GetDriver(), items[0].GetMountpoint(), items[0].GetUsageNote(), items[0].GetRefCount()),
		fmt.Sprintf("%s/%s/%s/%s/%d", items[1].GetName(), items[1].GetDriver(), items[1].GetMountpoint(), items[1].GetUsageNote(), items[1].GetRefCount()),
		fmt.Sprintf("%s/%s/%s/%s/%d", items[2].GetName(), items[2].GetDriver(), items[2].GetMountpoint(), items[2].GetUsageNote(), items[2].GetRefCount()),
	}
	want := []string{
		"alpha/local//a/note/9",
		"same/local//m/note/1",
		"same/local//m/note/2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected order: got=%v want=%v", got, want)
	}
}

func TestLessImageInfo(t *testing.T) {
	items := []*pb.ImageInfo{
		{Repository: "same", Tag: "latest", ImageId: "img-b"},
		{Repository: "same", Tag: "latest", ImageId: "img-a"},
		{Repository: "alpha", Tag: "latest", ImageId: "img-z"},
	}
	sort.Slice(items, func(i, j int) bool { return LessImageInfo(items[i], items[j]) })
	got := []string{
		items[0].GetRepository() + ":" + items[0].GetTag() + ":" + items[0].GetImageId(),
		items[1].GetRepository() + ":" + items[1].GetTag() + ":" + items[1].GetImageId(),
		items[2].GetRepository() + ":" + items[2].GetTag() + ":" + items[2].GetImageId(),
	}
	want := []string{
		"alpha:latest:img-z",
		"same:latest:img-a",
		"same:latest:img-b",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected order: got=%v want=%v", got, want)
	}
}

func TestLessContainerInfoNilSafety(t *testing.T) {
	item := &pb.ContainerInfo{ContainerName: "a", ContainerId: "id-a"}
	if !LessContainerInfo(item, nil) {
		t.Fatalf("expected non-nil to sort before nil")
	}
	if LessContainerInfo(nil, item) {
		t.Fatalf("expected nil not to sort before non-nil")
	}
	if LessContainerInfo(nil, nil) {
		t.Fatalf("expected nil/nil comparison to be false")
	}
}

func TestLessComposeGroupNilSafety(t *testing.T) {
	item := &pb.ComposeGroup{Project: "a", Service: "svc"}
	if !LessComposeGroup(item, nil) {
		t.Fatalf("expected non-nil to sort before nil")
	}
	if LessComposeGroup(nil, item) {
		t.Fatalf("expected nil not to sort before non-nil")
	}
	if LessComposeGroup(nil, nil) {
		t.Fatalf("expected nil/nil comparison to be false")
	}
}

func TestLessNetworkInfoNilSafety(t *testing.T) {
	item := &pb.NetworkInfo{Name: "a", NetworkId: "n1"}
	if !LessNetworkInfo(item, nil) {
		t.Fatalf("expected non-nil to sort before nil")
	}
	if LessNetworkInfo(nil, item) {
		t.Fatalf("expected nil not to sort before non-nil")
	}
	if LessNetworkInfo(nil, nil) {
		t.Fatalf("expected nil/nil comparison to be false")
	}
}

func TestLessVolumeInfoNilSafety(t *testing.T) {
	item := &pb.VolumeInfo{Name: "a"}
	if !LessVolumeInfo(item, nil) {
		t.Fatalf("expected non-nil to sort before nil")
	}
	if LessVolumeInfo(nil, item) {
		t.Fatalf("expected nil not to sort before non-nil")
	}
	if LessVolumeInfo(nil, nil) {
		t.Fatalf("expected nil/nil comparison to be false")
	}
}

func TestLessImageInfoNilSafety(t *testing.T) {
	item := &pb.ImageInfo{Repository: "a", Tag: "latest", ImageId: "img-a"}
	if !LessImageInfo(item, nil) {
		t.Fatalf("expected non-nil to sort before nil")
	}
	if LessImageInfo(nil, item) {
		t.Fatalf("expected nil not to sort before non-nil")
	}
	if LessImageInfo(nil, nil) {
		t.Fatalf("expected nil/nil comparison to be false")
	}
}

func TestLessComposeGroupCanonicalizesSlicesAndKeepsInputsUnchanged(t *testing.T) {
	gA := &pb.ComposeGroup{
		Project:        "same",
		Service:        "svc",
		ContainerIds:   []string{"b", "a"},
		NetworkNames:   []string{"n2", "n1"},
		ContainerNames: []string{"web-2", "web-1"},
	}
	gB := &pb.ComposeGroup{
		Project:        "same",
		Service:        "svc",
		ContainerIds:   []string{"a", "b"},
		NetworkNames:   []string{"n1", "n2"},
		ContainerNames: []string{"api-2", "api-1"},
	}

	beforeAIDs := append([]string(nil), gA.GetContainerIds()...)
	beforeANets := append([]string(nil), gA.GetNetworkNames()...)
	beforeANames := append([]string(nil), gA.GetContainerNames()...)
	beforeBIDs := append([]string(nil), gB.GetContainerIds()...)
	beforeBNets := append([]string(nil), gB.GetNetworkNames()...)
	beforeBNames := append([]string(nil), gB.GetContainerNames()...)

	items := []*pb.ComposeGroup{gA, gB}
	sort.Slice(items, func(i, j int) bool { return LessComposeGroup(items[i], items[j]) })

	if items[0] != gB || items[1] != gA {
		t.Fatalf("expected container_names tie-break to place gB before gA")
	}

	if !reflect.DeepEqual(beforeAIDs, gA.GetContainerIds()) || !reflect.DeepEqual(beforeANets, gA.GetNetworkNames()) || !reflect.DeepEqual(beforeANames, gA.GetContainerNames()) {
		t.Fatalf("expected gA slices unchanged, ids=%v nets=%v names=%v", gA.GetContainerIds(), gA.GetNetworkNames(), gA.GetContainerNames())
	}
	if !reflect.DeepEqual(beforeBIDs, gB.GetContainerIds()) || !reflect.DeepEqual(beforeBNets, gB.GetNetworkNames()) || !reflect.DeepEqual(beforeBNames, gB.GetContainerNames()) {
		t.Fatalf("expected gB slices unchanged, ids=%v nets=%v names=%v", gB.GetContainerIds(), gB.GetNetworkNames(), gB.GetContainerNames())
	}
}

func join(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ","
		}
		out += p
	}
	return out
}
