package grpc

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	"docksphinx/internal/event"
	"docksphinx/internal/monitor"
)

// EventToProto converts internal event to proto Event
func EventToProto(ev *event.Event) *pb.Event {
	if ev == nil {
		return nil
	}
	data := make(map[string]string)
	for k, v := range ev.Data {
		data[k] = fmtString(v)
	}
	return &pb.Event{
		Id:            ev.ID,
		Type:          string(ev.Type),
		TimestampUnix: ev.Timestamp.Unix(),
		ContainerId:   ev.ContainerID,
		ContainerName: ev.ContainerName,
		ImageName:     ev.ImageName,
		Message:       ev.Message,
		Data:          data,
	}
}

func fmtString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return fmt.Sprintf("%v", x)
	}
}

// StateToSnapshot builds proto Snapshot from StateManager
func StateToSnapshot(sm *monitor.StateManager) *pb.Snapshot {
	states := sm.GetAllStates()
	containers := make([]*pb.ContainerInfo, 0, len(states))
	metrics := make(map[string]*pb.ContainerMetrics)
	for _, st := range states {
		containers = append(containers, &pb.ContainerInfo{
			ContainerId:      st.ContainerID,
			ContainerName:    st.ContainerName,
			ImageName:        st.ImageName,
			State:            st.State,
			Status:           st.Status,
			LastSeenUnix:     st.LastSeen.Unix(),
			StartedAtUnix:    st.StartedAt.Unix(),
			UptimeSeconds:    st.UptimeSeconds,
			ComposeProject:   st.ComposeProject,
			ComposeService:   st.ComposeService,
			RestartCount:     clampIntToInt32(st.RestartCount),
			VolumeMountCount: clampIntToInt32(st.VolumeMountCount),
		})
		metrics[st.ContainerID] = &pb.ContainerMetrics{
			ContainerId:   st.ContainerID,
			CpuPercent:    st.CPUPercent,
			MemoryUsage:   st.MemoryUsage,
			MemoryLimit:   st.MemoryLimit,
			MemoryPercent: st.MemoryPercent,
			NetworkRx:     st.NetworkRx,
			NetworkTx:     st.NetworkTx,
		}
	}
	sort.Slice(containers, func(i, j int) bool {
		if containers[i].GetContainerName() == containers[j].GetContainerName() {
			return containers[i].GetContainerId() < containers[j].GetContainerId()
		}
		return containers[i].GetContainerName() < containers[j].GetContainerName()
	})

	resources := sm.GetResources()
	images := make([]*pb.ImageInfo, 0, len(resources.Images))
	for _, img := range resources.Images {
		images = append(images, &pb.ImageInfo{
			ImageId:     img.ID,
			Repository:  img.Repository,
			Tag:         img.Tag,
			Size:        img.Size,
			CreatedUnix: img.Created,
		})
	}
	sort.Slice(images, func(i, j int) bool {
		if images[i].GetRepository() != images[j].GetRepository() {
			return images[i].GetRepository() < images[j].GetRepository()
		}
		if images[i].GetTag() != images[j].GetTag() {
			return images[i].GetTag() < images[j].GetTag()
		}
		return images[i].GetImageId() < images[j].GetImageId()
	})

	networks := make([]*pb.NetworkInfo, 0, len(resources.Networks))
	for _, net := range resources.Networks {
		networks = append(networks, &pb.NetworkInfo{
			NetworkId:      net.ID,
			Name:           net.Name,
			Driver:         net.Driver,
			Scope:          net.Scope,
			Internal:       net.Internal,
			ContainerCount: clampIntToInt32(net.ContainerCount),
		})
	}
	sort.Slice(networks, func(i, j int) bool {
		if networks[i].GetName() != networks[j].GetName() {
			return networks[i].GetName() < networks[j].GetName()
		}
		if networks[i].GetDriver() != networks[j].GetDriver() {
			return networks[i].GetDriver() < networks[j].GetDriver()
		}
		if networks[i].GetScope() != networks[j].GetScope() {
			return networks[i].GetScope() < networks[j].GetScope()
		}
		return networks[i].GetNetworkId() < networks[j].GetNetworkId()
	})

	volumes := make([]*pb.VolumeInfo, 0, len(resources.Volumes))
	for _, vol := range resources.Volumes {
		volumes = append(volumes, &pb.VolumeInfo{
			Name:       vol.Name,
			Driver:     vol.Driver,
			Mountpoint: vol.Mountpoint,
			RefCount:   clampInt64ToInt32(vol.RefCount),
			UsageNote:  vol.UsageNote,
		})
	}
	sort.Slice(volumes, func(i, j int) bool {
		if volumes[i].GetName() != volumes[j].GetName() {
			return volumes[i].GetName() < volumes[j].GetName()
		}
		if volumes[i].GetDriver() != volumes[j].GetDriver() {
			return volumes[i].GetDriver() < volumes[j].GetDriver()
		}
		if volumes[i].GetMountpoint() != volumes[j].GetMountpoint() {
			return volumes[i].GetMountpoint() < volumes[j].GetMountpoint()
		}
		if volumes[i].GetUsageNote() != volumes[j].GetUsageNote() {
			return volumes[i].GetUsageNote() < volumes[j].GetUsageNote()
		}
		return volumes[i].GetRefCount() < volumes[j].GetRefCount()
	})

	groups := make([]*pb.ComposeGroup, 0, len(resources.Groups))
	for _, g := range resources.Groups {
		containerIDs := append([]string(nil), g.ContainerIDs...)
		containerNames := append([]string(nil), g.ContainerNames...)
		networkNames := append([]string(nil), g.NetworkNames...)
		sort.Strings(containerIDs)
		sort.Strings(containerNames)
		sort.Strings(networkNames)
		groups = append(groups, &pb.ComposeGroup{
			Project:        g.Project,
			Service:        g.Service,
			ContainerIds:   containerIDs,
			ContainerNames: containerNames,
			NetworkNames:   networkNames,
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].GetProject() != groups[j].GetProject() {
			return groups[i].GetProject() < groups[j].GetProject()
		}
		if groups[i].GetService() != groups[j].GetService() {
			return groups[i].GetService() < groups[j].GetService()
		}
		li := strings.Join(groups[i].GetContainerIds(), ",")
		lj := strings.Join(groups[j].GetContainerIds(), ",")
		if li != lj {
			return li < lj
		}
		return strings.Join(groups[i].GetNetworkNames(), ",") < strings.Join(groups[j].GetNetworkNames(), ",")
	})

	return &pb.Snapshot{
		Containers: containers,
		Metrics:    metrics,
		Images:     images,
		Networks:   networks,
		Volumes:    volumes,
		Groups:     groups,
		AtUnix:     time.Now().Unix(),
	}
}

func EventsToProto(events []*event.Event) []*pb.Event {
	out := make([]*pb.Event, 0, len(events))
	for _, ev := range events {
		converted := EventToProto(ev)
		if converted != nil {
			out = append(out, converted)
		}
	}
	return out
}

func clampIntToInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

func clampInt64ToInt32(v int64) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}
