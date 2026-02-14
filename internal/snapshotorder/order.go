package snapshotorder

import (
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
	li := strings.Join(a.GetContainerIds(), ",")
	lj := strings.Join(b.GetContainerIds(), ",")
	if li != lj {
		return li < lj
	}
	return strings.Join(a.GetNetworkNames(), ",") < strings.Join(b.GetNetworkNames(), ",")
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
