package screens

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type uploadedItem struct {
	Name      string
	Reference string
	Size      int64
	Timestamp int64
	Mimetype  string
}

func (i *index) showUploadCard(showUpload bool) *widget.Card {
	upForm := i.uploadForm()
	listButton := i.listUploadsButton(fyne.NewSize(200, 100))

	uploadCard := widget.NewCard("Upload", "upload content into swarm", container.NewVBox(upForm, listButton))
	uploadCard.Hidden = !showUpload

	return uploadCard
}

func (i *index) uploadForm() *widget.Form {
	filepath := ""
	mimetype := ""
	var pathBind = binding.BindString(&filepath)
	path := widget.NewEntry()
	path.Bind(pathBind)
	path.Disable()
	var file io.Reader
	openFileButton := widget.NewButton("File Open", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				i.showError(err)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()
			data, err := io.ReadAll(reader)
			if err != nil {
				i.showError(err)
				return
			}

			mimetype = reader.URI().MimeType()
			pathBind.Set(reader.URI().Name())
			file = bytes.NewReader(data)
			data = nil
		}, i.Window)
		fd.Show()
	})

	upForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Add file", Widget: path, HintText: "Filepath"},
			{Text: "Choose File", Widget: openFileButton},
		},
	}
	upForm.OnSubmit = func() {
		go func() {
			defer func() {
				pathBind.Set("")
				file = nil
			}()
			if file == nil {
				i.showError(fmt.Errorf("please select a file"))
				return
			}
			batchID := i.app.Preferences().String(batchPrefKey)
			if batchID == "" {
				i.showError(fmt.Errorf("please select a batch of stamp"))
				return
			}
			i.showProgressWithMessage(fmt.Sprintf("Uploading %s", path.Text))

			ref, err := i.bl.AddFileBzz(context.Background(), batchID, path.Text, mimetype, file)
			// just stand by
			<-time.After(time.Second * 120)
			if err != nil {
				i.hideProgress()
				i.showError(err)
				return
			}
			filename := path.Text
			uploadedSrt := i.app.Preferences().String(uploadsPrefKey)
			uploads := []uploadedItem{}
			if uploadedSrt != "" {
				err := json.Unmarshal([]byte(uploadedSrt), &uploads)
				if err != nil {
					i.showError(err)
				}
			}
			uploads = append(uploads, uploadedItem{
				Name:      filename,
				Reference: ref.String(),
			})
			data, err := json.Marshal(uploads)
			if err != nil {
				i.hideProgress()
				i.showError(err)
				return
			}
			i.app.Preferences().SetString(uploadsPrefKey, string(data))
			d := dialog.NewCustomConfirm("Upload successful", "Ok", "Cancel", i.refDialog(ref.String()), func(b bool) {}, i.Window)
			i.hideProgress()
			d.Show()
		}()
	}

	return upForm
}

func (i *index) listUploadsButton(minSize fyne.Size) *widget.Button {
	button := widget.NewButton("All Uploads", func() {
		uploadedContent := container.NewVBox()
		uploadedContentWrapper := container.NewScroll(uploadedContent)
		uploadedSrt := i.app.Preferences().String(uploadsPrefKey)
		uploads := []uploadedItem{}
		if uploadedSrt != "" {
			err := json.Unmarshal([]byte(uploadedSrt), &uploads)
			if err != nil {
				i.showError(err)
			}
			for _, v := range uploads {
				ref := v.Reference
				name := v.Name
				label := widget.NewLabel(name)
				label.Wrapping = fyne.TextWrapWord
				item := container.NewBorder(label, nil, nil, widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
					i.Window.Clipboard().SetContent(ref)
				}))
				uploadedContent.Add(item)
			}
		}

		if len(uploads) == 0 {
			uploadedContent.Add(widget.NewLabel("Empty upload list"))
		}

		child := i.app.NewWindow("Uploaded content")
		size := child.Canvas().Content().MinSize()
		if size.Width < minSize.Width{
			size.Width = minSize.Width
		}
		if size.Height < minSize.Height {
			size.Height = minSize.Height
		}
		child.Resize(size)
		child.SetContent(uploadedContentWrapper)
		child.Show()
	})

	return button
}
