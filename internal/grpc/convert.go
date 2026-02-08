package grpc

import (
	"fmt"
	"strconv"
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
			ContainerId:   st.ContainerID,
			ContainerName: st.ContainerName,
			ImageName:     st.ImageName,
			State:         st.State,
			Status:        st.Status,
			LastSeenUnix:  st.LastSeen.Unix(),
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
	return &pb.Snapshot{
		Containers: containers,
		Metrics:    metrics,
		AtUnix:     time.Now().Unix(),
	}
}
