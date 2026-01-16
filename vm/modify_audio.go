package vm

func (m *VMachine) SetAudioEnabled(client *VmSshClient, audioEnabled bool, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-enabled", audioEnabled, callBack)
}

func (m *VMachine) SetAudioController(client *VmSshClient, audioController AudioControllerType, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-controller", audioController, callBack)
}

func (m *VMachine) SetAudioCodec(client *VmSshClient, audioCodec AudioCodecType, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-codec", audioCodec, callBack)
}

func (m *VMachine) SetAudioInEnabled(client *VmSshClient, audioInEnabled bool, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-in", audioInEnabled, callBack)
}

func (m *VMachine) SetAudioOutEnabled(client *VmSshClient, audioOutEnabled bool, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-out", audioOutEnabled, callBack)
}

func (m *VMachine) SetAudioDriver(client *VmSshClient, audioDriver AudioDriverType, callBack func(uuid string)) error {
	return m.setProperty(client, "audio-driver", audioDriver, callBack)
}
