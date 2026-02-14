package eventorder

import (
	pb "docksphinx/api/docksphinx/v1"
	"docksphinx/internal/event"
)

type orderedEvent struct {
	TimestampUnix int64
	ID            string
	ContainerName string
	Type          string
	Message       string
	ContainerID   string
	ImageName     string
}

func LessPB(a, b *pb.Event) bool {
	return lessOrdered(
		orderedEvent{
			TimestampUnix: a.GetTimestampUnix(),
			ID:            a.GetId(),
			ContainerName: a.GetContainerName(),
			Type:          a.GetType(),
			Message:       a.GetMessage(),
			ContainerID:   a.GetContainerId(),
			ImageName:     a.GetImageName(),
		},
		orderedEvent{
			TimestampUnix: b.GetTimestampUnix(),
			ID:            b.GetId(),
			ContainerName: b.GetContainerName(),
			Type:          b.GetType(),
			Message:       b.GetMessage(),
			ContainerID:   b.GetContainerId(),
			ImageName:     b.GetImageName(),
		},
	)
}

func LessInternal(a, b *event.Event) bool {
	return lessOrdered(
		orderedEvent{
			TimestampUnix: a.Timestamp.Unix(),
			ID:            a.ID,
			ContainerName: a.ContainerName,
			Type:          string(a.Type),
			Message:       a.Message,
			ContainerID:   a.ContainerID,
			ImageName:     a.ImageName,
		},
		orderedEvent{
			TimestampUnix: b.Timestamp.Unix(),
			ID:            b.ID,
			ContainerName: b.ContainerName,
			Type:          string(b.Type),
			Message:       b.Message,
			ContainerID:   b.ContainerID,
			ImageName:     b.ImageName,
		},
	)
}

func lessOrdered(a, b orderedEvent) bool {
	if a.TimestampUnix != b.TimestampUnix {
		return a.TimestampUnix > b.TimestampUnix
	}
	if a.ID != b.ID {
		return a.ID < b.ID
	}
	if a.ContainerName != b.ContainerName {
		return a.ContainerName < b.ContainerName
	}
	if a.Type != b.Type {
		return a.Type < b.Type
	}
	if a.Message != b.Message {
		return a.Message < b.Message
	}
	if a.ContainerID != b.ContainerID {
		return a.ContainerID < b.ContainerID
	}
	if a.ImageName != b.ImageName {
		return a.ImageName < b.ImageName
	}
	return false
}
