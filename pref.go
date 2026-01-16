package main

const (
	PREF_TREE_UPDATE_TIME_KEY   = "tree.update_time"
	PREF_TREE_UPDATE_TIME_VALUE = 20000

	PREF_TREE_UPDATE_DELAY_KEY   = "tree.update_delay"
	PREF_TREE_UPDATE_DELAY_VALUE = 1000

	PREF_TASKS_MAX_ENTRIES_KEY   = "tasks.max_entries"
	PREF_TASKS_MAX_ENTRIES_VALUE = 15

	PREF_SERVERS_KEY          = "serverlist"
	PREF_MASTERKEY_TEST_KEY   = "mastertest"
	PREF_MASTERKEY_TEST_VALUE = "Reiner"
)

type Preferences struct {
	TreeUpdateTime  int // msec
	TreeDelayTime   int // msec
	TasksMaxEntries int
	ServerList      string // json String
	MasterKeyTest   string
}

func NewPreferences() *Preferences {
	p := &Preferences{
		TreeUpdateTime:  Gui.App.Preferences().IntWithFallback(PREF_TREE_UPDATE_TIME_KEY, PREF_TREE_UPDATE_TIME_VALUE),
		TreeDelayTime:   Gui.App.Preferences().IntWithFallback(PREF_TREE_UPDATE_DELAY_KEY, PREF_TREE_UPDATE_DELAY_VALUE),
		TasksMaxEntries: Gui.App.Preferences().IntWithFallback(PREF_TASKS_MAX_ENTRIES_KEY, PREF_TASKS_MAX_ENTRIES_VALUE),
		ServerList:      Gui.App.Preferences().StringWithFallback(PREF_SERVERS_KEY, ""),
		MasterKeyTest:   Gui.App.Preferences().StringWithFallback(PREF_MASTERKEY_TEST_KEY, PREF_MASTERKEY_TEST_VALUE),
	}
	return p
}

func (p *Preferences) Store() {
	pref := Gui.App.Preferences()
	pref.SetInt(PREF_TREE_UPDATE_TIME_KEY, p.TreeUpdateTime)
	pref.SetInt(PREF_TREE_UPDATE_DELAY_KEY, p.TreeDelayTime)
	pref.SetInt(PREF_TASKS_MAX_ENTRIES_KEY, p.TasksMaxEntries)
	pref.SetString(PREF_SERVERS_KEY, p.ServerList)
	pref.SetString(PREF_MASTERKEY_TEST_KEY, p.MasterKeyTest)
}
