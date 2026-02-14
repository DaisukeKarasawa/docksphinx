package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	dgrpc "docksphinx/internal/grpc"
	"docksphinx/internal/snapshotorder"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type panel int

const (
	panelLeft panel = iota
	panelCenter
	panelRight
)

type sortMode int

const (
	sortCPU sortMode = iota
	sortMemory
	sortUptime
	sortName
)

var targetOrder = []string{"containers", "images", "networks", "volumes", "groups"}

type tuiModel struct {
	targetIdx int
	filter    string
	sortMode  sortMode
	paused    bool

	snapshot *pb.Snapshot
	events   []*pb.Event

	left   *tview.List
	center *tview.Table
	right  *tview.TextView
	bottom *tview.TextView
	pages  *tview.Pages

	focusPanel panel
}

func runTUI(parent context.Context, address string) error {
	ctx, cancel := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app := tview.NewApplication()
	model := newTUIModel()
	root := model.layout(app)

	errCh := make(chan error, 1)
	go func() {
		errCh <- model.streamLoop(ctx, app, address)
	}()

	app.SetRoot(root, true)
	app.SetInputCapture(model.captureInput(app, cancel))

	runErr := app.Run()
	cancel()

	select {
	case streamErr := <-errCh:
		if streamErr != nil && streamErr != context.Canceled && streamErr != io.EOF {
			return streamErr
		}
	default:
	}
	return runErr
}

func (m *tuiModel) streamLoop(ctx context.Context, app *tview.Application, address string) error {
	backoff := 500 * time.Millisecond
	for {
		if ctx.Err() != nil {
			return nil
		}

		client, err := dgrpc.NewClient(ctx, address)
		if err != nil {
			m.queueStatus(app, fmt.Sprintf("stream connect failed: %v (retrying)", err))
			if err := waitOrDone(ctx, backoff); err != nil {
				return nil
			}
			backoff = nextBackoff(backoff)
			continue
		}

		stream, err := client.Stream(ctx, true)
		if err != nil {
			_ = client.Close()
			m.queueStatus(app, fmt.Sprintf("stream subscribe failed: %v (retrying)", err))
			if err := waitOrDone(ctx, backoff); err != nil {
				return nil
			}
			backoff = nextBackoff(backoff)
			continue
		}

		m.queueStatus(app, "stream connected")
		backoff = 500 * time.Millisecond

		recvErr := m.consumeStream(ctx, app, stream)
		_ = client.Close()
		if recvErr == nil || errors.Is(recvErr, context.Canceled) || errors.Is(recvErr, io.EOF) {
			if ctx.Err() != nil {
				return nil
			}
		} else {
			m.queueStatus(app, fmt.Sprintf("stream disconnected: %v (reconnecting)", recvErr))
		}

		if err := waitOrDone(ctx, backoff); err != nil {
			return nil
		}
		backoff = nextBackoff(backoff)
	}
}

func newTUIModel() *tuiModel {
	left := tview.NewList().ShowSecondaryText(false)
	center := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)
	right := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	bottom := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	pages := tview.NewPages()

	m := &tuiModel{
		targetIdx: 0,
		filter:    "",
		sortMode:  sortCPU,
		paused:    false,
		left:      left,
		center:    center,
		right:     right,
		bottom:    bottom,
		pages:     pages,
	}

	for _, target := range targetOrder {
		targetCopy := target
		left.AddItem(titleTarget(targetCopy), "", 0, func() {
			m.setTarget(targetCopy)
		})
	}
	left.SetCurrentItem(0)
	m.refreshAll()
	return m
}

func (m *tuiModel) layout(app *tview.Application) tview.Primitive {
	m.left.SetBorder(true).SetTitle("Targets")
	m.center.SetBorder(true).SetTitle("Overview")
	m.right.SetBorder(true).SetTitle("Details / Recent Events")
	m.bottom.SetBorder(true).SetTitle("Status")

	m.center.SetSelectedFunc(func(row, _ int) {
		if row == 0 {
			return
		}
		m.renderRight()
	})

	top := tview.NewFlex().
		AddItem(m.left, 24, 1, true).
		AddItem(m.center, 0, 3, false).
		AddItem(m.right, 42, 2, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(top, 0, 1, true).
		AddItem(m.bottom, 2, 0, false)

	m.pages.AddPage("main", root, true, true)
	return m.pages
}

func (m *tuiModel) captureInput(app *tview.Application, cancel context.CancelFunc) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			m.cycleFocus(app, true)
			return nil
		case tcell.KeyBacktab:
			m.cycleFocus(app, false)
			return nil
		case tcell.KeyRight:
			m.cycleFocus(app, true)
			return nil
		case tcell.KeyLeft:
			m.cycleFocus(app, false)
			return nil
		case tcell.KeyUp:
			m.moveSelection(-1)
			return nil
		case tcell.KeyDown:
			m.moveSelection(1)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				cancel()
				app.Stop()
				return nil
			case 'j':
				m.moveSelection(1)
				return nil
			case 'k':
				m.moveSelection(-1)
				return nil
			case '/':
				m.openSearchModal(app)
				return nil
			case 's':
				m.toggleSort()
				return nil
			case 'p':
				m.paused = !m.paused
				m.renderStatus("")
				return nil
			}
		}
		return event
	}
}

func (m *tuiModel) cycleFocus(app *tview.Application, forward bool) {
	if forward {
		m.focusPanel = (m.focusPanel + 1) % 3
	} else {
		if m.focusPanel == 0 {
			m.focusPanel = 2
		} else {
			m.focusPanel--
		}
	}
	switch m.focusPanel {
	case panelLeft:
		app.SetFocus(m.left)
	case panelCenter:
		app.SetFocus(m.center)
	case panelRight:
		app.SetFocus(m.right)
	}
}

func (m *tuiModel) moveSelection(delta int) {
	switch m.focusPanel {
	case panelLeft:
		next := m.left.GetCurrentItem() + delta
		if next < 0 {
			next = 0
		}
		if next >= len(targetOrder) {
			next = len(targetOrder) - 1
		}
		m.left.SetCurrentItem(next)
		m.setTarget(targetOrder[next])
	case panelCenter:
		row, col := m.center.GetSelection()
		row += delta
		if row < 1 {
			row = 1
		}
		if row >= m.center.GetRowCount() {
			row = m.center.GetRowCount() - 1
		}
		m.center.Select(row, col)
		m.renderRight()
	}
}

func (m *tuiModel) openSearchModal(app *tview.Application) {
	input := tview.NewInputField().
		SetLabel("Filter / search: ").
		SetText(m.filter)
	input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			m.filter = input.GetText()
			m.pages.RemovePage("search")
			m.refreshCenter()
		case tcell.KeyEscape:
			m.pages.RemovePage("search")
		}
		app.SetFocus(m.center)
	})

	form := tview.NewForm().
		AddFormItem(input).
		AddButton("Apply", func() {
			m.filter = input.GetText()
			m.pages.RemovePage("search")
			m.refreshCenter()
			app.SetFocus(m.center)
		}).
		AddButton("Cancel", func() {
			m.pages.RemovePage("search")
			app.SetFocus(m.center)
		})
	form.SetBorder(true).SetTitle("Search").SetTitleAlign(tview.AlignLeft)
	m.pages.AddPage("search", centered(60, 10, form), true, true)
	app.SetFocus(input)
}

func (m *tuiModel) consumeStream(
	ctx context.Context,
	app *tview.Application,
	stream pb.DocksphinxService_StreamClient,
) error {
	if stream == nil {
		return errors.New("stream client is nil")
	}
	if app == nil {
		return errors.New("tui application is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		update, err := stream.Recv()
		if err != nil {
			return err
		}
		if update == nil {
			continue
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		app.QueueUpdateDraw(func() {
			if m.paused {
				m.renderStatus("paused (incoming stream ignored)")
				return
			}
			switch payload := update.GetPayload().(type) {
			case *pb.StreamUpdate_Snapshot:
				m.snapshot = payload.Snapshot
				if payload.Snapshot != nil {
					m.events = compactEvents(payload.Snapshot.GetRecentEvents(), 200)
				} else {
					m.events = nil
				}
				m.refreshAll()
			case *pb.StreamUpdate_Event:
				m.events = compactEvents(append([]*pb.Event{payload.Event}, m.events...), 200)
				m.renderRight()
				m.renderStatus("")
			}
		})
	}
}

func (m *tuiModel) queueStatus(app *tview.Application, message string) {
	if app == nil {
		return
	}
	app.QueueUpdateDraw(func() {
		m.renderStatus(message)
	})
}

func (m *tuiModel) setTarget(target string) {
	for i, t := range targetOrder {
		if t == target {
			m.targetIdx = i
			break
		}
	}
	m.refreshCenter()
}

func (m *tuiModel) refreshAll() {
	m.refreshCenter()
	m.renderRight()
	m.renderStatus("")
}

func (m *tuiModel) refreshCenter() {
	target := targetOrder[m.targetIdx]
	m.center.Clear()

	switch target {
	case "containers":
		m.renderContainers()
	case "images":
		m.renderImages()
	case "networks":
		m.renderNetworks()
	case "volumes":
		m.renderVolumes()
	case "groups":
		m.renderGroups()
	}

	if m.center.GetRowCount() > 1 {
		m.center.Select(1, 0)
	}
	m.renderRight()
	m.renderStatus("")
}

func (m *tuiModel) renderContainers() {
	headers := []string{"NAME", "STATE", "CPU%", "MEM%", "UPTIME(s)", "RX", "TX", "IMAGE", "GROUP"}
	for i, h := range headers {
		m.center.SetCell(0, i, tview.NewTableCell(fmt.Sprintf("[::b]%s", h)))
	}
	if m.snapshot == nil {
		return
	}

	type row struct {
		c         *pb.ContainerInfo
		cpu       float64
		mem       float64
		rx        int64
		tx        int64
		hasMetric bool
		group     string
	}
	rows := make([]row, 0, len(m.snapshot.GetContainers()))
	for _, c := range m.snapshot.GetContainers() {
		if c == nil {
			continue
		}
		metric := m.snapshot.GetMetrics()[c.GetContainerId()]
		r := row{c: c}
		if metric != nil {
			r.cpu = metric.GetCpuPercent()
			r.mem = metric.GetMemoryPercent()
			r.rx = metric.GetNetworkRx()
			r.tx = metric.GetNetworkTx()
			r.hasMetric = true
		}
		r.group = fmt.Sprintf("%s/%s", c.GetComposeProject(), c.GetComposeService())
		if m.matchesFilter(strings.Join([]string{
			c.GetContainerName(),
			c.GetImageName(),
			c.GetState(),
			r.group,
		}, " ")) {
			rows = append(rows, r)
		}
	}

	switch m.sortMode {
	case sortCPU:
		sort.Slice(rows, func(i, j int) bool {
			return lessContainerForMode(m.sortMode, rows[i].c, rows[j].c, rows[i].cpu, rows[j].cpu, rows[i].mem, rows[j].mem)
		})
	case sortMemory:
		sort.Slice(rows, func(i, j int) bool {
			return lessContainerForMode(m.sortMode, rows[i].c, rows[j].c, rows[i].cpu, rows[j].cpu, rows[i].mem, rows[j].mem)
		})
	case sortUptime:
		sort.Slice(rows, func(i, j int) bool {
			return lessContainerForMode(m.sortMode, rows[i].c, rows[j].c, rows[i].cpu, rows[j].cpu, rows[i].mem, rows[j].mem)
		})
	default:
		sort.Slice(rows, func(i, j int) bool {
			return lessContainerForMode(m.sortMode, rows[i].c, rows[j].c, rows[i].cpu, rows[j].cpu, rows[i].mem, rows[j].mem)
		})
	}

	for i, r := range rows {
		prefix := ""
		if r.hasMetric && (r.cpu >= 80 || r.mem >= 85) {
			prefix = "🔥"
		}
		name := trimContainerName(r.c.GetContainerName())
		if ev := m.lastEventType(r.c.GetContainerId()); ev == "died" || ev == "restarted" {
			name = "⚠ " + name
		}
		values := []string{
			prefix + name,
			r.c.GetState(),
			formatFloat1OrNA(r.cpu, r.hasMetric),
			formatFloat1OrNA(r.mem, r.hasMetric),
			formatUptimeOrNA(r.c),
			formatInt64OrNA(r.rx, r.hasMetric),
			formatInt64OrNA(r.tx, r.hasMetric),
			r.c.GetImageName(),
			r.group,
		}
		for col, v := range values {
			m.center.SetCell(i+1, col, tview.NewTableCell(v))
		}
	}
}

func (m *tuiModel) renderImages() {
	headers := []string{"REPOSITORY", "TAG", "SIZE", "CREATED"}
	for i, h := range headers {
		m.center.SetCell(0, i, tview.NewTableCell(fmt.Sprintf("[::b]%s", h)))
	}
	if m.snapshot == nil {
		return
	}
	images := append([]*pb.ImageInfo(nil), m.snapshot.GetImages()...)
	sort.Slice(images, func(i, j int) bool {
		return snapshotorder.LessImageInfo(images[i], images[j])
	})

	row := 1
	for _, img := range images {
		if img == nil {
			continue
		}
		if !m.matchesFilter(img.GetRepository() + " " + img.GetTag()) {
			continue
		}
		values := []string{
			img.GetRepository(),
			img.GetTag(),
			fmt.Sprintf("%d", img.GetSize()),
			formatDateTimeOrNA(img.GetCreatedUnix()),
		}
		for col, v := range values {
			m.center.SetCell(row, col, tview.NewTableCell(v))
		}
		row++
	}
}

func (m *tuiModel) renderNetworks() {
	headers := []string{"NAME", "DRIVER", "SCOPE", "INTERNAL", "CONTAINERS"}
	for i, h := range headers {
		m.center.SetCell(0, i, tview.NewTableCell(fmt.Sprintf("[::b]%s", h)))
	}
	if m.snapshot == nil {
		return
	}
	networks := append([]*pb.NetworkInfo(nil), m.snapshot.GetNetworks()...)
	sort.Slice(networks, func(i, j int) bool {
		return snapshotorder.LessNetworkInfo(networks[i], networks[j])
	})
	row := 1
	for _, n := range networks {
		if n == nil {
			continue
		}
		if !m.matchesFilter(n.GetName() + " " + n.GetDriver()) {
			continue
		}
		values := []string{
			n.GetName(),
			n.GetDriver(),
			n.GetScope(),
			fmt.Sprintf("%t", n.GetInternal()),
			fmt.Sprintf("%d", n.GetContainerCount()),
		}
		for col, v := range values {
			m.center.SetCell(row, col, tview.NewTableCell(v))
		}
		row++
	}
}

func (m *tuiModel) renderVolumes() {
	headers := []string{"NAME", "DRIVER", "REFS", "USAGE", "MOUNTPOINT"}
	for i, h := range headers {
		m.center.SetCell(0, i, tview.NewTableCell(fmt.Sprintf("[::b]%s", h)))
	}
	if m.snapshot == nil {
		return
	}
	volumes := append([]*pb.VolumeInfo(nil), m.snapshot.GetVolumes()...)
	sort.Slice(volumes, func(i, j int) bool {
		return snapshotorder.LessVolumeInfo(volumes[i], volumes[j])
	})
	row := 1
	for _, v := range volumes {
		if v == nil {
			continue
		}
		if !m.matchesFilter(v.GetName() + " " + v.GetDriver()) {
			continue
		}
		values := []string{
			v.GetName(),
			v.GetDriver(),
			fmt.Sprintf("%d", v.GetRefCount()),
			v.GetUsageNote(),
			v.GetMountpoint(),
		}
		for col, value := range values {
			m.center.SetCell(row, col, tview.NewTableCell(value))
		}
		row++
	}
}

func (m *tuiModel) renderGroups() {
	headers := []string{"PROJECT", "SERVICE", "CONTAINERS", "NETWORKS"}
	for i, h := range headers {
		m.center.SetCell(0, i, tview.NewTableCell(fmt.Sprintf("[::b]%s", h)))
	}
	if m.snapshot == nil {
		return
	}
	groups := append([]*pb.ComposeGroup(nil), m.snapshot.GetGroups()...)
	sort.Slice(groups, func(i, j int) bool {
		return snapshotorder.LessComposeGroup(groups[i], groups[j])
	})
	row := 1
	for _, g := range groups {
		if g == nil {
			continue
		}
		containerNames := append([]string(nil), g.GetContainerNames()...)
		networkNames := append([]string(nil), g.GetNetworkNames()...)
		sort.Strings(containerNames)
		sort.Strings(networkNames)
		c := strings.Join(containerNames, ",")
		n := strings.Join(networkNames, ",")
		if !m.matchesFilter(g.GetProject() + " " + g.GetService() + " " + c + " " + n) {
			continue
		}
		values := []string{
			g.GetProject(),
			g.GetService(),
			c,
			n,
		}
		for col, value := range values {
			m.center.SetCell(row, col, tview.NewTableCell(value))
		}
		row++
	}
}

func (m *tuiModel) renderRight() {
	target := targetOrder[m.targetIdx]
	row, _ := m.center.GetSelection()
	if row <= 0 {
		m.right.SetText("[gray]No selection")
		return
	}

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("[yellow]Target:[white] %s\n", target))
	builder.WriteString(fmt.Sprintf("[yellow]Filter:[white] %s\n\n", m.filter))

	switch target {
	case "containers":
		if m.snapshot != nil {
			containers := m.snapshot.GetContainers()
			idx := row - 1
			filtered := m.filteredContainerRowsForDetail()
			if idx >= 0 && idx < len(filtered) {
				c := filtered[idx]
				metric := m.snapshot.GetMetrics()[c.GetContainerId()]
				builder.WriteString(fmt.Sprintf("[green]Container:[white] %s\n", trimContainerName(c.GetContainerName())))
				builder.WriteString(fmt.Sprintf("ID: %s\n", c.GetContainerId()))
				builder.WriteString(fmt.Sprintf("State: %s\n", c.GetState()))
				builder.WriteString(fmt.Sprintf("Image: %s\n", c.GetImageName()))
				builder.WriteString(fmt.Sprintf("Uptime: %s\n", formatUptimeOrNA(c)))
				builder.WriteString(fmt.Sprintf("Compose: %s/%s\n", c.GetComposeProject(), c.GetComposeService()))
				builder.WriteString(fmt.Sprintf("RestartCount: %d\n", c.GetRestartCount()))
				builder.WriteString(fmt.Sprintf("VolumeMounts: %d\n", c.GetVolumeMountCount()))
				if metric != nil {
					builder.WriteString(fmt.Sprintf("CPU: %.2f%%\n", metric.GetCpuPercent()))
					builder.WriteString(fmt.Sprintf("Memory: %.2f%% (%d/%d)\n", metric.GetMemoryPercent(), metric.GetMemoryUsage(), metric.GetMemoryLimit()))
					builder.WriteString(fmt.Sprintf("Network Rx/Tx: %d / %d\n", metric.GetNetworkRx(), metric.GetNetworkTx()))
				} else {
					builder.WriteString("CPU: N/A\n")
					builder.WriteString("Memory: N/A\n")
					builder.WriteString("Network Rx/Tx: N/A\n")
				}
			} else if len(containers) == 0 {
				builder.WriteString("No containers\n")
			}
		}
	}

	builder.WriteString("\n[green]Recent Events[white]\n")
	shown := 0
	for _, ev := range m.events {
		if ev == nil {
			continue
		}
		if shown >= 10 {
			break
		}
		builder.WriteString(fmt.Sprintf(
			"[%s] %-13s %-16s %s\n",
			time.Unix(ev.GetTimestampUnix(), 0).Format("15:04:05"),
			ev.GetType(),
			trimContainerName(ev.GetContainerName()),
			ev.GetMessage(),
		))
		shown++
	}
	m.right.SetText(builder.String())
}

func (m *tuiModel) renderStatus(message string) {
	state := "LIVE"
	if m.paused {
		state = "PAUSED"
	}
	sortName := map[sortMode]string{
		sortCPU:    "CPU",
		sortMemory: "MEM",
		sortUptime: "UPTIME",
		sortName:   "NAME",
	}[m.sortMode]

	if message == "" {
		message = "Tab/←→:panel  j/k:move  /:search  s:sort  p:pause  q:quit"
	}
	m.bottom.SetText(fmt.Sprintf(
		"[yellow]State:[white] %s  [yellow]Target:[white] %s  [yellow]Sort:[white] %s  [yellow]Filter:[white] %s  [yellow]Info:[white] %s",
		state,
		targetOrder[m.targetIdx],
		sortName,
		m.filter,
		message,
	))
}

func (m *tuiModel) toggleSort() {
	m.sortMode = (m.sortMode + 1) % 4
	m.refreshCenter()
}

func (m *tuiModel) matchesFilter(value string) bool {
	if strings.TrimSpace(m.filter) == "" {
		return true
	}
	return strings.Contains(strings.ToLower(value), strings.ToLower(m.filter))
}

func (m *tuiModel) lastEventType(containerID string) string {
	for _, ev := range m.events {
		if ev == nil {
			continue
		}
		if ev.GetContainerId() == containerID {
			return ev.GetType()
		}
	}
	return ""
}

func (m *tuiModel) filteredContainerRowsForDetail() []*pb.ContainerInfo {
	if m.snapshot == nil {
		return nil
	}
	type row struct {
		c   *pb.ContainerInfo
		cpu float64
		mem float64
	}
	rows := make([]row, 0, len(m.snapshot.GetContainers()))
	for _, c := range m.snapshot.GetContainers() {
		if c == nil {
			continue
		}
		metric := m.snapshot.GetMetrics()[c.GetContainerId()]
		cpu := 0.0
		mem := 0.0
		if metric != nil {
			cpu = metric.GetCpuPercent()
			mem = metric.GetMemoryPercent()
		}
		group := fmt.Sprintf("%s/%s", c.GetComposeProject(), c.GetComposeService())
		if m.matchesFilter(strings.Join([]string{c.GetContainerName(), c.GetImageName(), c.GetState(), group}, " ")) {
			rows = append(rows, row{c: c, cpu: cpu, mem: mem})
		}
	}
	switch m.sortMode {
	case sortCPU:
		sort.Slice(rows, func(i, j int) bool {
			return lessContainerForMode(m.sortMode, rows[i].c, rows[j].c, rows[i].cpu, rows[j].cpu, rows[i].mem, rows[j].mem)
		})
	case sortMemory:
		sort.Slice(rows, func(i, j int) bool {
			return lessContainerForMode(m.sortMode, rows[i].c, rows[j].c, rows[i].cpu, rows[j].cpu, rows[i].mem, rows[j].mem)
		})
	case sortUptime:
		sort.Slice(rows, func(i, j int) bool {
			return lessContainerForMode(m.sortMode, rows[i].c, rows[j].c, rows[i].cpu, rows[j].cpu, rows[i].mem, rows[j].mem)
		})
	default:
		sort.Slice(rows, func(i, j int) bool {
			return lessContainerForMode(m.sortMode, rows[i].c, rows[j].c, rows[i].cpu, rows[j].cpu, rows[i].mem, rows[j].mem)
		})
	}
	out := make([]*pb.ContainerInfo, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.c)
	}
	return out
}

func centered(width, height int, p tview.Primitive) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false),
			width, 1, true,
		).
		AddItem(nil, 0, 1, false)
}

func titleTarget(v string) string {
	if v == "" {
		return ""
	}
	return strings.ToUpper(v[:1]) + v[1:]
}

func formatFloat1OrNA(value float64, ok bool) string {
	if !ok {
		return "N/A"
	}
	return fmt.Sprintf("%.1f", value)
}

func formatInt64OrNA(value int64, ok bool) string {
	if !ok {
		return "N/A"
	}
	return fmt.Sprintf("%d", value)
}

func lessContainerForMode(mode sortMode, a, b *pb.ContainerInfo, cpuA, cpuB, memA, memB float64) bool {
	switch mode {
	case sortCPU:
		if cpuA == cpuB {
			return lessContainerNameID(a, b)
		}
		return cpuA > cpuB
	case sortMemory:
		if memA == memB {
			return lessContainerNameID(a, b)
		}
		return memA > memB
	case sortUptime:
		if a.GetUptimeSeconds() == b.GetUptimeSeconds() {
			return lessContainerNameID(a, b)
		}
		return a.GetUptimeSeconds() > b.GetUptimeSeconds()
	default:
		return lessContainerNameID(a, b)
	}
}

func lessContainerNameID(a, b *pb.ContainerInfo) bool {
	if a.GetContainerName() == b.GetContainerName() {
		return a.GetContainerId() < b.GetContainerId()
	}
	return a.GetContainerName() < b.GetContainerName()
}

func compactEvents(events []*pb.Event, max int) []*pb.Event {
	if len(events) == 0 || max == 0 {
		return nil
	}
	if max < 0 {
		max = len(events)
	}
	out := make([]*pb.Event, 0, len(events))
	for _, ev := range events {
		if ev == nil {
			continue
		}
		out = append(out, ev)
		if len(out) >= max {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func formatDateTimeOrNA(unix int64) string {
	if unix <= 0 {
		return "N/A"
	}
	return time.Unix(unix, 0).Format("2006-01-02 15:04")
}
