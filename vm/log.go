package vm

import (
	"fmt"
)

func (m *VMachine) addLogEntry(entry []string, lock bool) {
	if lock {
		m.lock.Lock()
		defer m.lock.Unlock()
	}
	m.logBuffer = append(m.logBuffer, entry)
	if len(m.logBuffer) > MAX_LOG_ENTRIES {
		m.logBuffer = m.logBuffer[1:]
	}
}

func (m *VMachine) PrintLogEntries(lock bool) {
	if lock {
		m.lock.RLock()
		defer m.lock.RUnlock()
	}
	for _, entry := range m.logBuffer {
		fmt.Println("---------------------------")
		for _, line := range entry {
			fmt.Println(line)
		}
	}
}

func (m *VMachine) GetLogBuffer() [][]string {
	return m.logBuffer
}
