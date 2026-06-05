package tray

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/google/uuid"
	"github.com/varmiguemunoz/sprintos/internal/api"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
	"github.com/varmiguemunoz/sprintos/internal/pomodoro"
	"gorm.io/gorm"
)

const tasksPerPage = 10
const trayPort = ":8765"

type trayApp struct {
	mu sync.Mutex

	client *Client
	pomo   *pomodoro.Session

	allTasks    []TaskSummary
	currentPage int

	selectedTaskID    uint
	selectedTaskTitle string

	taskSlots  []*systray.MenuItem
	taskIDs    []uint
	taskTitles []string

	timerRunning bool
	timerTaskID  uint
	initialized  bool

	mTimerHeader *systray.MenuItem
	mTimerStatus *systray.MenuItem
	mCurrentTask *systray.MenuItem
	mSelectTask  *systray.MenuItem
	mLoadingTask *systray.MenuItem
	mTaskPrev    *systray.MenuItem
	mTaskPage    *systray.MenuItem
	mTaskNext    *systray.MenuItem
	mStartTimer  *systray.MenuItem
	mStopTimer   *systray.MenuItem

	mPomoHeader *systray.MenuItem
	mPomo15     *systray.MenuItem
	mPomo30     *systray.MenuItem
	mPomo45     *systray.MenuItem
	mPomoStop   *systray.MenuItem

	mQuit *systray.MenuItem
}

func Run(database *gorm.DB) error {
	internalToken := uuid.New().String()

	client := NewClient("http://localhost"+trayPort, internalToken)

	apiServer := api.NewServer(database, internalToken)
	go func() {
		mux := http.NewServeMux()
		apiServer.RegisterRoutes(mux)
		srv := &http.Server{Addr: trayPort, Handler: mux}
		_ = srv.ListenAndServe()
	}()

	pomo := pomodoro.New(
		func() {
			_ = beeep.Notify(
				"SprintOS — Pomodoro",
				"Time's up! Take a break.\nAuto-restarting in 15 seconds — press Stop to cancel.",
				"",
			)
		},
		func() {
			_ = beeep.Notify(
				"SprintOS — Pomodoro",
				"Pomodoro restarted automatically. Stay focused!",
				"",
			)
		},
	)

	session, _ := auth.LoadSession()
	_ = session

	t := &trayApp{
		client:     client,
		pomo:       pomo,
		taskIDs:    make([]uint, tasksPerPage),
		taskTitles: make([]string, tasksPerPage),
		taskSlots:  make([]*systray.MenuItem, tasksPerPage),
	}

	systray.Run(t.onReady, t.onExit)
	return nil
}

func (t *trayApp) onReady() {
	systray.SetIcon(iconBytes())
	systray.SetTooltip("SprintOS")

	t.mTimerHeader = systray.AddMenuItem("⏱  TIMER", "")
	t.mTimerHeader.Disable()

	t.mTimerStatus = systray.AddMenuItem("Status: Stopped", "")
	t.mTimerStatus.Disable()

	t.mCurrentTask = systray.AddMenuItem("No task selected", "")
	t.mCurrentTask.Disable()

	t.mSelectTask = systray.AddMenuItem("Select Task  ▶", "Browse and select a task to track")

	t.mLoadingTask = t.mSelectTask.AddSubMenuItem("Loading tasks…", "")
	t.mLoadingTask.Disable()

	for i := 0; i < tasksPerPage; i++ {
		t.taskSlots[i] = t.mSelectTask.AddSubMenuItem("", "")
		t.taskSlots[i].Hide()
	}

	t.mTaskPrev = t.mSelectTask.AddSubMenuItem("← Previous", "Go to previous page")
	t.mTaskPage = t.mSelectTask.AddSubMenuItem("", "")
	t.mTaskPage.Disable()
	t.mTaskNext = t.mSelectTask.AddSubMenuItem("→ Next", "Go to next page")
	t.mTaskPrev.Hide()
	t.mTaskPage.Hide()
	t.mTaskNext.Hide()

	systray.AddSeparator()

	t.mStartTimer = systray.AddMenuItem("▶  Start Timer", "Start tracking time for the selected task")
	t.mStopTimer = systray.AddMenuItem("■  Stop Timer", "Stop the running timer")
	t.mStopTimer.Disable()

	systray.AddSeparator()

	t.mPomoHeader = systray.AddMenuItem("🍅  POMODORO  —  Focus Mode", "")
	t.mPomoHeader.Disable()

	systray.AddSeparator()

	t.mPomo15 = systray.AddMenuItem("Start 15 minutes", "Focus session of 15 minutes")
	t.mPomo30 = systray.AddMenuItem("Start 30 minutes", "Focus session of 30 minutes")
	t.mPomo45 = systray.AddMenuItem("Start 45 minutes", "Focus session of 45 minutes")

	systray.AddSeparator()

	t.mPomoStop = systray.AddMenuItem("■  Stop Pomodoro", "Stop the current focus session")
	t.mPomoStop.Disable()

	systray.AddSeparator()

	t.mQuit = systray.AddMenuItem("Quit SprintOS", "")

	go t.loadTasks()
	go t.eventLoop()
	go t.updateLoop()
}

func (t *trayApp) onExit() {
	t.pomo.Stop()
}

func (t *trayApp) loadTasks() {
	for i := 0; i < 20; i++ {
		if t.client.IsReady() {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}

	tasks, err := t.client.ListAllTasks()
	if err != nil {
		t.mLoadingTask.SetTitle("Could not load tasks — is the app running?")
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.mLoadingTask.Hide()
	t.allTasks = tasks

	if len(tasks) == 0 {
		t.mLoadingTask.SetTitle("No tasks found")
		t.mLoadingTask.Show()
		return
	}

	t.renderPage(0)
}

func (t *trayApp) renderPage(page int) {
	t.currentPage = page
	total := len(t.allTasks)
	totalPages := (total + tasksPerPage - 1) / tasksPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	for i := range t.taskSlots {
		t.taskSlots[i].Hide()
		t.taskSlots[i].SetTitle("")
		t.taskIDs[i] = 0
		t.taskTitles[i] = ""
	}

	start := page * tasksPerPage
	end := start + tasksPerPage
	if end > total {
		end = total
	}

	for slotIdx, taskIdx := 0, start; taskIdx < end; slotIdx, taskIdx = slotIdx+1, taskIdx+1 {
		title := clampStr(t.allTasks[taskIdx].Title, 38)
		t.taskSlots[slotIdx].SetTitle(title)
		t.taskIDs[slotIdx] = t.allTasks[taskIdx].ID
		t.taskTitles[slotIdx] = title
		t.taskSlots[slotIdx].Show()
	}

	if totalPages <= 1 {
		t.mTaskPrev.Hide()
		t.mTaskPage.Hide()
		t.mTaskNext.Hide()
		return
	}

	t.mTaskPage.SetTitle(fmt.Sprintf("Page %d of %d", page+1, totalPages))
	t.mTaskPrev.Show()
	t.mTaskPage.Show()
	t.mTaskNext.Show()

	if page <= 0 {
		t.mTaskPrev.Disable()
	} else {
		t.mTaskPrev.Enable()
	}

	if page >= totalPages-1 {
		t.mTaskNext.Disable()
	} else {
		t.mTaskNext.Enable()
	}
}

func (t *trayApp) selectTask(idx int) {
	t.mu.Lock()
	taskID := t.taskIDs[idx]
	title := t.taskTitles[idx]
	t.selectedTaskID = taskID
	t.selectedTaskTitle = title
	t.mu.Unlock()

	label := clampStr(title, 40)
	if label == "" {
		label = fmt.Sprintf("Task #%d", taskID)
	}
	t.mCurrentTask.SetTitle("Task: " + label)
}

func (t *trayApp) handleStartTimer() {
	t.mu.Lock()
	taskID := t.selectedTaskID
	title := t.selectedTaskTitle
	running := t.timerRunning
	t.mu.Unlock()

	if taskID == 0 {
		_ = beeep.Notify("SprintOS", "Please select a task first.", "")
		return
	}

	if running {
		_ = beeep.Notify("SprintOS", "A timer is already running. Stop it first.", "")
		return
	}

	if err := t.client.StartTimer(taskID); err != nil {
		_ = beeep.Notify("SprintOS", "Could not start timer: "+err.Error(), "")
		return
	}

	t.mu.Lock()
	t.timerRunning = true
	t.timerTaskID = taskID
	t.mu.Unlock()

	t.mStartTimer.Disable()
	t.mStopTimer.Enable()
	_ = beeep.Notify("SprintOS", "Timer started — "+clampStr(title, 40), "")
}

func (t *trayApp) handleStopTimer() {
	if err := t.client.StopTimer(); err != nil {
		_ = beeep.Notify("SprintOS", "Could not stop timer: "+err.Error(), "")
		return
	}
	t.mu.Lock()
	t.timerRunning = false
	t.timerTaskID = 0
	t.mu.Unlock()
	t.mStartTimer.Enable()
	t.mStopTimer.Disable()
	t.mTimerStatus.SetTitle("Status: Stopped")
	systray.SetTitle("")
}

func (t *trayApp) handlePomoStart(minutes int) {
	t.pomo.Stop()
	t.pomo.Start(minutes)
	t.mPomoStop.Enable()
	t.mPomo15.Disable()
	t.mPomo30.Disable()
	t.mPomo45.Disable()
}

func (t *trayApp) handlePomoStop() {
	t.pomo.Stop()
	t.mPomoStop.Disable()
	t.mPomo15.Enable()
	t.mPomo30.Enable()
	t.mPomo45.Enable()
	systray.SetTitle("")
}

func (t *trayApp) handleTaskPrev() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.currentPage > 0 {
		t.renderPage(t.currentPage - 1)
	}
}

func (t *trayApp) handleTaskNext() {
	t.mu.Lock()
	defer t.mu.Unlock()
	total := len(t.allTasks)
	totalPages := (total + tasksPerPage - 1) / tasksPerPage
	if t.currentPage < totalPages-1 {
		t.renderPage(t.currentPage + 1)
	}
}

func (t *trayApp) eventLoop() {
	for i, slot := range t.taskSlots {
		i, slot := i, slot
		go func() {
			for range slot.ClickedCh {
				t.selectTask(i)
			}
		}()
	}

	for {
		select {
		case <-t.mStartTimer.ClickedCh:
			t.handleStartTimer()
		case <-t.mStopTimer.ClickedCh:
			t.handleStopTimer()
		case <-t.mPomo15.ClickedCh:
			t.handlePomoStart(15)
		case <-t.mPomo30.ClickedCh:
			t.handlePomoStart(30)
		case <-t.mPomo45.ClickedCh:
			t.handlePomoStart(45)
		case <-t.mPomoStop.ClickedCh:
			t.handlePomoStop()
		case <-t.mTaskPrev.ClickedCh:
			t.handleTaskPrev()
		case <-t.mTaskNext.ClickedCh:
			t.handleTaskNext()
		case <-t.mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (t *trayApp) updateLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		t.updateTimerStatus()
		t.updatePomoStatus()
	}
}

func (t *trayApp) updateTimerStatus() {
	status, err := t.client.GetActiveTimer()

	t.mu.Lock()
	wasRunning := t.timerRunning
	wasInitialized := t.initialized
	t.mu.Unlock()

	if err != nil || status == nil || !status.Running {
		t.mu.Lock()
		t.timerRunning = false
		t.timerTaskID = 0
		t.initialized = true
		t.mu.Unlock()

		if wasRunning {
			t.mStartTimer.Enable()
			t.mStopTimer.Disable()
			t.mTimerStatus.SetTitle("Status: Stopped")
			systray.SetTitle("")
		} else if !wasInitialized {
			t.mTimerStatus.SetTitle("Status: Stopped")
		}
		return
	}

	t.mu.Lock()
	prevTaskID := t.timerTaskID
	t.timerRunning = true
	t.timerTaskID = status.TaskID
	t.initialized = true
	t.mu.Unlock()

	if !wasRunning || prevTaskID != status.TaskID {
		t.mStartTimer.Disable()
		t.mStopTimer.Enable()
	}

	elapsed := time.Since(status.StartedAt)
	h := int(elapsed.Hours())
	m := int(elapsed.Minutes()) % 60
	s := int(elapsed.Seconds()) % 60

	var label string
	if h > 0 {
		label = fmt.Sprintf("%dh %02dm", h, m)
	} else {
		label = fmt.Sprintf("%02d:%02d", m, s)
	}

	t.mTimerStatus.SetTitle("Active: " + label + " — " + clampStr(status.TaskTitle, 24))

	if t.pomo.GetState() == pomodoro.Idle {
		systray.SetTitle("⏱ " + label)
	}
}

func (t *trayApp) updatePomoStatus() {
	state := t.pomo.GetState()

	switch state {
	case pomodoro.Running:
		systray.SetTitle("🍅 " + pomodoro.FormatDuration(t.pomo.Remaining()))
	case pomodoro.Grace:
		systray.SetTitle("⚠ " + pomodoro.FormatDuration(t.pomo.GraceRemaining()))
	}
}
