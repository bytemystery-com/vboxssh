// Copyright (c) 2026 Reiner Pröls
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// SPDX-License-Identifier: MIT
//
// Author: Reiner Pröls

package filebrowser

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"bytemystery-com/vboxssh/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bytemystery-com/colorlabel"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SftpHelper struct{}

func (s *SftpHelper) DeleteFile(client *ssh.Client, file string) error {
	if client != nil {
		sc, err := sftp.NewClient(client)
		if err != nil {
			return err
		}

		defer sc.Close()
		return sc.Remove(file)
	} else {
		return os.Remove(file)
	}
}

type BackFileInfo struct{}

func (b BackFileInfo) Name() string      { return ".." }
func (b BackFileInfo) Size() int64       { return 0 }
func (b BackFileInfo) Mode() fs.FileMode { return fs.ModeDir }

func (b BackFileInfo) ModTime() time.Time { return time.Time{} }
func (b BackFileInfo) IsDir() bool        { return true }
func (b BackFileInfo) Sys() any           { return nil }

type DummyFileInfo struct {
	name    string
	isDir   bool
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

func (d DummyFileInfo) Name() string      { return d.name }
func (d DummyFileInfo) Size() int64       { return d.size }
func (d DummyFileInfo) Mode() fs.FileMode { return d.mode }

func (d DummyFileInfo) ModTime() time.Time { return d.modTime }
func (d DummyFileInfo) IsDir() bool        { return d.isDir }
func (d DummyFileInfo) Sys() any           { return nil }

type SftpFileBrowserModeType int

const (
	SftpFileBrowserMode_openfile SftpFileBrowserModeType = iota
	SftpFileBrowserMode_savefile
	SftpFileBrowserMode_selectdir
)

type SftpFileBrowser struct {
	sshClient    *ssh.Client
	sftpClient   *sftp.Client
	ActualDir    string
	list         *widget.List
	server       *colorlabel.ColorLabel
	label        *widget.Label
	name         *widget.Entry
	fileList     []os.FileInfo
	selectedItem os.FileInfo
	filter       *regexp.Regexp
	title        string
	mode         SftpFileBrowserModeType
}

func (s *SftpFileBrowser) RealPath(file string) (string, error) {
	if s.sftpClient != nil {
		return s.sftpClient.RealPath(file)
	} else {
		abs, err := filepath.Abs(file)
		if err != nil {
			return "", err
		}

		real, err := filepath.EvalSymlinks(abs)
		if err != nil {
			return "", err
		}
		return real, nil
	}
}

func (s *SftpFileBrowser) ReadDir(file string) ([]os.FileInfo, error) {
	if s.sftpClient != nil {
		return s.sftpClient.ReadDir(file)
	} else {
		dirs, err := os.ReadDir(file)
		if err != nil {
			return nil, err
		}
		fi := make([]os.FileInfo, 0, len(dirs))
		for _, item := range dirs {
			i, err := item.Info()
			if err == nil {
				fi = append(fi, i)
			}
		}
		return fi, nil
	}
}

func (s *SftpFileBrowser) Stat(file string) (os.FileInfo, error) {
	if s.sftpClient != nil {
		return s.sftpClient.Stat(file)
	} else {
		return os.Stat(file)
	}
}

func (s *SftpFileBrowser) Close() {
	if s.sftpClient != nil {
		s.sftpClient.Close()
	}
}

func (s *SftpFileBrowser) Join(elem ...string) string {
	if s.sftpClient != nil {
		return path.Join(elem...)
	} else {
		return filepath.Join(elem...)
	}
}

func (s *SftpFileBrowser) Dir(file string) string {
	if s.sftpClient != nil {
		return path.Dir(file)
	} else {
		return filepath.Dir(file)
	}
}

func (s *SftpFileBrowser) Base(file string) string {
	if s.sftpClient != nil {
		return path.Base(file)
	} else {
		return filepath.Base(file)
	}
}

func (s *SftpFileBrowser) Remove(file string) error {
	if s.sftpClient != nil {
		return s.sftpClient.Remove(file)
	} else {
		return os.Remove(file)
	}
}

func NewSftpBrowser(sshClient *ssh.Client, startDir string, f *regexp.Regexp, title string, mode SftpFileBrowserModeType) *SftpFileBrowser {
	ftp := SftpFileBrowser{}
	ftp.ActualDir = startDir
	ftp.sshClient = sshClient
	ftp.filter = f
	ftp.title = title
	ftp.mode = mode
	if ftp.ActualDir == "" {
		ftp.ActualDir = "."
	}
	if ftp.sshClient != nil {
		sc, err := sftp.NewClient(ftp.sshClient)
		if err != nil {
			return nil
		}
		ftp.sftpClient = sc
	}

	p, err := ftp.RealPath(ftp.ActualDir)
	if err != nil {
		err := ftp.ResetActualDir()
		if err != nil {
			ftp.Close()
			return nil
		}
	} else {
		ftp.ActualDir = p
	}
	i, err := ftp.Stat(ftp.ActualDir)
	if err != nil {
		err := ftp.ResetActualDir()
		if err != nil {
			ftp.Close()
			return nil
		}
	} else {
		if !i.IsDir() {
			err := ftp.ResetActualDir()
			if err != nil {
				ftp.Close()
				return nil
			}
		}
	}

	return &ftp
}

func (s *SftpFileBrowser) ResetActualDir() error {
	s.ActualDir = "."
	p, err := s.RealPath(s.ActualDir)
	if err != nil {
		s.Close()
		return nil
	}
	s.ActualDir = p
	return nil
}

func (s *SftpFileBrowser) Show(w fyne.Window, windowScale float32, fOk func(string, os.FileInfo, string)) {
	server := util.GetServerAddressAsString(s.sshClient)
	s.server = colorlabel.NewColorLabel(fmt.Sprintf(lang.X("filebrowser.server", "Server: %s"), server), theme.ColorNamePrimary, nil, 1.0)
	s.server.SetTextStyle(&fyne.TextStyle{
		Bold: true,
	})
	var dia *dialog.ConfirmDialog
	s.label = widget.NewLabel("")
	s.list = widget.NewList(s.listNumberOfItems, s.listCreateItem, s.listUpdateItem)
	s.list.OnSelected = s.listOnSelected
	s.list.OnUnselected = s.listOnUnSelected
	s.name = widget.NewEntry()
	s.name.PlaceHolder = lang.X("filebrowser.filename", "Filename")
	s.name.OnSubmitted = func(string) {
		dia.Confirm()
	}

	var c *fyne.Container
	if s.mode == SftpFileBrowserMode_openfile || s.mode == SftpFileBrowserMode_selectdir {
		c = container.NewBorder(container.NewVBox(s.server, s.label), util.NewVFiller(1.0), nil, nil, s.list)
	} else {
		c = container.NewBorder(container.NewVBox(s.server, s.label), container.NewVBox(util.NewVFiller(1.0), s.name, util.NewVFiller(1.0)), nil, nil, s.list)
	}
	dia = dialog.NewCustomConfirm(s.title,
		lang.X("filebrowser.ok", "Ok"),
		lang.X("filebrowser.cancel", "Cancel"),
		c, func(ok bool) {
			defer s.Close()
			if ok {
				var p string
				var err error
				switch s.mode {
				case SftpFileBrowserMode_openfile:
					p, err = s.getAcualSelectedItem()
					if err != nil {
						return
					}
				case SftpFileBrowserMode_savefile:
					p, err = s.getSaveFilePath()
					if err != nil {
						return
					}
				case SftpFileBrowserMode_selectdir:
					p, err = s.RealPath(s.ActualDir)
					if err != nil {
						return
					}
				}
				fi, err := s.Stat(p)
				if err != nil {
					fi = nil
				}
				fOk(p, fi, s.Dir(p))
			}
		}, w)

	si := w.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
	dia.Show()
	s.browse()
}

func (s *SftpFileBrowser) listOnSelected(id widget.ListItemID) {
	s.selectedItem = s.fileList[id]
	if s.selectedItem.IsDir() {
		n, err := s.getAcualSelectedItem()
		if err != nil {
			return
		}
		s.ActualDir = n
		s.browse()
	} else {
		s.name.SetText(s.selectedItem.Name())
	}
}

func (s *SftpFileBrowser) listOnUnSelected(id widget.ListItemID) {
	s.selectedItem = nil
	s.name.SetText("")
}

func (s *SftpFileBrowser) listNumberOfItems() int {
	return len(s.fileList)
}

func (s *SftpFileBrowser) listCreateItem() fyne.CanvasObject {
	text := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	icon := canvas.NewImageFromResource(theme.FileIcon())
	icon.SetMinSize(fyne.NewSize(16, 16)) // gewünschte Größe
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()

	return container.NewHBox(icon, util.NewFiller(16, 0), text)
}

func (s *SftpFileBrowser) listUpdateItem(id widget.ListItemID, o fyne.CanvasObject) {
	c, ok := o.(*fyne.Container)
	if !ok {
		return
	}

	text, ok := c.Objects[2].(*canvas.Text)
	if !ok {
		return
	}

	icon, ok := c.Objects[0].(*canvas.Image)
	if !ok {
		return
	}

	f := s.fileList[id]
	text.Text = f.Name()
	if f.IsDir() {
		icon.Resource = theme.FolderIcon()
		text.Color = theme.Color(theme.ColorNameForeground)
		text.TextStyle = fyne.TextStyle{
			Bold: true,
		}
	} else {
		icon.Resource = theme.FileIcon()
		text.Color = theme.Color(theme.ColorNamePrimary)
		text.TextStyle = fyne.TextStyle{
			Bold: false,
		}
	}
	text.Refresh()
	icon.Refresh()
}

func (s *SftpFileBrowser) getAcualSelectedItem() (string, error) {
	if s.selectedItem == nil {
		return "", errors.New("nothing selected")
	}
	cwd, err := s.RealPath(s.ActualDir)
	if err != nil {
		return "", err
	}
	p, err := s.RealPath(s.Join(cwd, s.selectedItem.Name()))
	if err != nil {
		return "", err
	}
	return p, nil
}

func (s *SftpFileBrowser) getSaveFilePath() (string, error) {
	cwd, err := s.RealPath(s.ActualDir)
	if err != nil {
		return "", err
	}
	return s.Join(cwd, s.name.Text), nil
}

func (s *SftpFileBrowser) followSymlink(parent string, f os.FileInfo) (os.FileInfo, error) {
	if (f.Mode() & os.ModeSymlink) != 0 {
		t, err := s.RealPath(s.Join(parent, f.Name()))
		if err != nil {
			return nil, err
		}

		target, err := s.Stat(t)
		if err != nil {
			return nil, err
		}

		if (target.Mode() & os.ModeSymlink) != 0 {
			return s.followSymlink(s.Dir(t), target)
		}

		if target.IsDir() {
			return &DummyFileInfo{
				name:    f.Name(),
				isDir:   true,
				mode:    target.Mode(),
				modTime: f.ModTime(),
			}, nil
		}
		return f, nil
	} else {
		return f, nil
	}
}

func (s *SftpFileBrowser) browse() error {
	files, err := s.ReadDir(s.ActualDir)
	if err != nil {
		return err
	}
	dirList := make([]os.FileInfo, 0, len(files))
	dirList = append(dirList, BackFileInfo{})

	fileList := make([]os.FileInfo, 0, len(files))
	for _, file := range files {
		file, err := s.followSymlink(s.ActualDir, file)
		if err == nil {
			if file.IsDir() {
				dirList = append(dirList, file)
			} else if s.mode != SftpFileBrowserMode_selectdir {
				if s.filter == nil || s.filter.MatchString(file.Name()) {
					fileList = append(fileList, file)
				}
			}
		}
	}

	slices.SortFunc(dirList, func(a, b os.FileInfo) int {
		A := a.Name()
		B := b.Name()
		an := strings.ToLower(A)
		bn := strings.ToLower(B)

		if an == bn {
			if A < B {
				return -1
			}
			if A > B {
				return 1
			}
			return 0
		}
		if an < bn {
			return -1
		}
		return 1
	})

	slices.SortFunc(fileList, func(a, b os.FileInfo) int {
		A := a.Name()
		B := b.Name()
		an := strings.ToLower(A)
		bn := strings.ToLower(B)

		if an == bn {
			if A < B {
				return -1
			}
			if A > B {
				return 1
			}
			return 0
		}
		if an < bn {
			return -1
		}
		return 1
	})
	s.fileList = dirList
	s.fileList = append(s.fileList, fileList...)
	s.list.Refresh()
	s.list.ScrollToTop()
	s.list.UnselectAll()

	l, err := s.RealPath(s.ActualDir)
	if err != nil {
		return err
	}
	s.label.SetText(fmt.Sprintf(lang.X("filebrowser.label", "%s  --  %d directories, %d files"),
		l, len(dirList)-1, len(fileList)))

	return nil
}
