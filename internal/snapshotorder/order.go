package snapshotorder

import (
	"sort"
	"strings"

	pb "docksphinx/api/docksphinx/v1"
)

func LessContainerInfo(a, b *pb.ContainerInfo) bool {
	if a == nil || b == nil {
		return a != nil && b == nil
	}
	if a.GetContainerName() == b.GetContainerName() {
		return a.GetContainerId() < b.GetContainerId()
	}
	return a.GetContainerName() < b.GetContainerName()
}

func LessComposeGroup(a, b *pb.ComposeGroup) bool {
	if a == nil || b == nil {
		return a != nil && b == nil
	}
	if a.GetProject() != b.GetProject() {
		return a.GetProject() < b.GetProject()
	}
	if a.GetService() != b.GetService() {
		return a.GetService() < b.GetService()
	}
	li := sortedJoined(a.GetContainerIds())
	lj := sortedJoined(b.GetContainerIds())
	if li != lj {
		return li < lj
	}
	ni := sortedJoined(a.GetNetworkNames())
	nj := sortedJoined(b.GetNetworkNames())
	if ni != nj {
		return ni < nj
	}
	return sortedJoined(a.GetContainerNames()) < sortedJoined(b.GetContainerNames())
}

func LessNetworkInfo(a, b *pb.NetworkInfo) bool {
	if a == nil || b == nil {
		return a != nil && b == nil
	}
	if a.GetName() != b.GetName() {
		return a.GetName() < b.GetName()
	}
	if a.GetDriver() != b.GetDriver() {
		return a.GetDriver() < b.GetDriver()
	}
	if a.GetScope() != b.GetScope() {
		return a.GetScope() < b.GetScope()
	}
	return a.GetNetworkId() < b.GetNetworkId()
}

func LessVolumeInfo(a, b *pb.VolumeInfo) bool {
	if a == nil || b == nil {
		return a != nil && b == nil
	}
	if a.GetName() != b.GetName() {
		return a.GetName() < b.GetName()
	}
	if a.GetDriver() != b.GetDriver() {
		return a.GetDriver() < b.GetDriver()
	}
	if a.GetMountpoint() != b.GetMountpoint() {
		return a.GetMountpoint() < b.GetMountpoint()
	}
	if a.GetUsageNote() != b.GetUsageNote() {
		return a.GetUsageNote() < b.GetUsageNote()
	}
	return a.GetRefCount() < b.GetRefCount()
}

func LessImageInfo(a, b *pb.ImageInfo) bool {
	if a == nil || b == nil {
		return a != nil && b == nil
	}
	if a.GetRepository() != b.GetRepository() {
		return a.GetRepository() < b.GetRepository()
	}
	if a.GetTag() != b.GetTag() {
		return a.GetTag() < b.GetTag()
	}
	return a.GetImageId() < b.GetImageId()
}

func sortedJoined(values []string) string {
	if len(values) == 0 {
		return ""
	}
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return strings.Join(copied, ",")
}
