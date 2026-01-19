package ratiolayout

import "fyne.io/fyne/v2"

type RatioLayout struct {
	LeftRatio float32 // z.B. 0.33
}

func (l *RatioLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 2 {
		return
	}

	leftW := size.Width * l.LeftRatio

	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(leftW, size.Height))

	objects[1].Move(fyne.NewPos(leftW, 0))
	objects[1].Resize(fyne.NewSize(size.Width-leftW, size.Height))
}

func (l *RatioLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 2 {
		return fyne.NewSize(0, 0)
	}

	left := objects[0].MinSize()
	right := objects[1].MinSize()

	return fyne.NewSize(
		left.Width+right.Width,
		max(left.Height, right.Height),
	)
}
