package main

import (
	"bytemystery-com/vboxssh/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TaskStatusType int

const (
	TaskStatus_running TaskStatusType = iota
	TaskStatus_finished
	TaskStatus_aborted
)

type TaskData struct {
	uuid   string
	name   string
	status string
	state  TaskStatusType
}

type TasksInfos struct {
	content  fyne.CanvasObject
	list     *widget.List
	tasks    []*TaskData
	nameSize fyne.Size
	iconSize fyne.Size
}

func NewTasksInfos() *TasksInfos {
	t := TasksInfos{}
	t.nameSize = util.GetDefaultTextSize("XXXXXXXXXXXXXXXXXXXXXXX")
	t.iconSize = fyne.NewSize(16, 16)
	t.list = widget.NewList(t.listGetNumberOfItems, t.listCreateItem, t.listUpdateItem)
	t.content = container.NewBorder(nil, nil, nil, nil, t.list)
	t.tasks = make([]*TaskData, 0, Gui.Settings.TasksMaxEntries)

	return &t
}

func (t *TasksInfos) AddTask(uuid, name, status string) {
	task := TaskData{
		uuid:   uuid,
		name:   name,
		status: status,
		state:  TaskStatus_running,
	}
	t.tasks = append(t.tasks, &task)
	t.checkTasks(false)
	t.list.Refresh()
}

func (t *TasksInfos) findTask(uuid string) *TaskData {
	for _, item := range t.tasks {
		if item.uuid == uuid {
			return item
		}
	}
	return nil
}

func (t *TasksInfos) setTaskStatus(uuid, status string, append bool, state TaskStatusType) {
	fyne.Do(func() {
		task := t.findTask(uuid)
		if task == nil {
			return
		}
		if append {
			task.status += status
		} else {
			task.status = status
		}
		task.state = state
		t.list.Refresh()
	})
}

func (t *TasksInfos) UpdateTaskStatus(uuid, status string, append bool) {
	t.setTaskStatus(uuid, status, append, TaskStatus_running)
}

func (t *TasksInfos) FinishTask(uuid, status string, append bool) {
	t.setTaskStatus(uuid, status, append, TaskStatus_finished)
}

func (t *TasksInfos) AbortTask(uuid, status string, append bool) {
	t.setTaskStatus(uuid, status, append, TaskStatus_aborted)
}

func (t *TasksInfos) checkTasks(refresh bool) {
	if len(t.tasks) <= Gui.Settings.TasksMaxEntries {
		return
	}
	count := len(t.tasks) - Gui.Settings.TasksMaxEntries
	newList := make([]*TaskData, 0, Gui.Settings.TasksMaxEntries)
	for _, item := range t.tasks {
		if count > 0 && (item.state == TaskStatus_finished || item.state == TaskStatus_aborted) {
			count--
		} else {
			newList = append(newList, item)
		}
	}
	t.tasks = newList
	if refresh {
		t.list.Refresh()
	}
}

func (t *TasksInfos) listGetNumberOfItems() int {
	return len(t.tasks)
}

func (t *TasksInfos) listCreateItem() fyne.CanvasObject {
	icon := canvas.NewImageFromResource(theme.CancelIcon())
	icon.SetMinSize(t.iconSize)
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	name := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	name.Refresh()

	status := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	status.Refresh()

	return container.NewBorder(nil, nil, container.NewHBox(icon,
		container.NewGridWrap(t.nameSize, name)), nil, status)
}

func (t *TasksInfos) listUpdateItem(id widget.ListItemID, o fyne.CanvasObject) {
	c, ok := o.(*fyne.Container)
	if !ok {
		return
	}
	status, ok := c.Objects[0].(*canvas.Text)
	if !ok {
		return
	}

	c, ok = c.Objects[1].(*fyne.Container)
	if !ok {
		return
	}

	icon, ok := c.Objects[0].(*canvas.Image)
	if !ok {
		return
	}

	c, ok = c.Objects[1].(*fyne.Container)
	if !ok {
		return
	}

	name, ok := c.Objects[0].(*canvas.Text)
	if !ok {
		return
	}

	task := t.tasks[id]
	name.Text = util.TruncateText(task.name, t.nameSize.Width, name, util.Begin)

	switch task.state {
	case TaskStatus_running:
		icon.Resource = theme.MediaPlayIcon()
		name.Color = theme.Color(theme.ColorNameWarning)
	case TaskStatus_finished:
		icon.Resource = theme.ConfirmIcon()
		name.Color = theme.Color(theme.ColorNameSuccess)
	case TaskStatus_aborted:
		icon.Resource = theme.ContentClearIcon()
		name.Color = theme.Color(theme.ColorNameError)
	}
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(t.iconSize)

	status.Text = util.TruncateText(task.status, t.list.Size().Width-t.nameSize.Width-t.iconSize.Width, status, util.Begin)
	status.Color = theme.Color(theme.ColorNamePrimary)

	status.Refresh()
	name.Refresh()
	icon.Refresh()
}
