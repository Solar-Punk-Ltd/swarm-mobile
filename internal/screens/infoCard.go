package screens

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (i *index) showInfoCard(showUpload bool) *widget.Card {
	addressContent := i.addressContent()
	batchRadio := i.batchRadio()
	stampsContent := i.stampsContent(batchRadio, showUpload)
	buyBatchButton := i.buyBatchButton(batchRadio, showUpload)
	balanceContent := i.balanceContent()
	balanceContent.Hidden = !showUpload

	infoCard := widget.NewCard("Info",
				fmt.Sprintf("Connected with %d peers", i.bl.TopologyDriver.Snapshot().Connected),
				container.NewVBox(addressContent, balanceContent, stampsContent, buyBatchButton))

	// auto reload
	go func() {
		for {
			time.Sleep(time.Second * 5)
			if i.bl != nil {
				infoCard.SetSubTitle(fmt.Sprintf("Connected with %d peers", i.bl.TopologyDriver.Snapshot().Connected))
			}
		}
	}()

	return infoCard
}

func (i* index) addressContent() *fyne.Container {
	addrCopyButton := widget.NewButtonWithIcon("   Copy   ", theme.ContentCopyIcon(), func() {
		i.Window.Clipboard().SetContent(i.bl.OverlayEthAddress.String())
	})
	addrHeader := container.NewHBox(widget.NewLabel("Overlay address:"))
	addr := container.NewHBox(
		widget.NewLabel(i.bl.OverlayEthAddress.String()),
		addrCopyButton,
	)
	addressContent := container.NewVBox(addrHeader, addr)

	return addressContent
}

func (i *index) balanceContent() *fyne.Container {
	balanceHeader := container.NewHBox(widget.NewLabel("Chequebook balance:"))
	chequebookBalance, err := i.bl.ChequebookBalance()
	if err != nil {
		i.logger.Log(fmt.Sprintf("Cannot get chequebook balance: %s", err.Error()))
		return container.NewHBox(widget.NewLabel("Cannot get chequebook balance"))
	}

	balance := container.NewHBox(widget.NewLabel(chequebookBalance.String()))
	balanceContent := container.NewVBox(balanceHeader, balance)

	return balanceContent
}

func (i *index) stampsContent(batchRadio *widget.RadioGroup, showUpload bool) *fyne.Container {
	stampsHeader := container.NewHBox(widget.NewLabel("Postage stamps:"))
	stamps := i.bl.GetAllBatches()

	if len(stamps) != 0 {
		selectedStamp := i.app.Preferences().String(selectedStampPrefKey)
		for _, v := range stamps {
			batchRadio.Append(shortenHashOrAddress(hex.EncodeToString(v.ID())))
		}

		batchRadio.SetSelected(selectedStamp)
	}
	stampsContent := container.NewVBox(stampsHeader, batchRadio)
	stampsContent.Hidden = !showUpload

	return stampsContent
}

func (i *index) batchRadio() *widget.RadioGroup {
	radioGroup := widget.NewRadioGroup([]string{}, func(s string) {
		if s == "" {
			i.app.Preferences().SetString(selectedStampPrefKey, "")
			i.app.Preferences().SetString(batchPrefKey, "")
			return
		}
		batches := i.bl.GetAllBatches()
		for _, v := range batches {
			stamp := hex.EncodeToString(v.ID())
			if s[0:6] == stamp[0:6] {
				i.app.Preferences().SetString(selectedStampPrefKey, s)
				i.app.Preferences().SetString(batchPrefKey, stamp)
			}
		}
	})

	return radioGroup
}

func (i *index) buyBatchButton(batchRadio *widget.RadioGroup, showUpload bool) *widget.Button {
	button := widget.NewButton("Buy a postage batch", func() {
		child := i.app.NewWindow("Buying a postage batch")
		// buy stamp
		depthStr := defaultDepth
		amountStr := defaultAmount
		content := container.NewStack()
		buyBatchContent := i.buyBatchDialog(&depthStr,&amountStr)
		size := child.Canvas().Content().MinSize()
		if size.Width < 200{
			size.Width = 200
		}
		if size.Height < 100 {
			size.Height = 100
		}
		child.Resize(size)

		buyButton := widget.NewButton("Buy", func() {
			amount, ok := big.NewInt(0).SetString(amountStr, 10)
			if !ok {
				i.showError(fmt.Errorf("invalid amountStr"))
				return
			}
			depth, err := strconv.ParseUint(depthStr, 10, 8)
			if err != nil {
				i.showError(fmt.Errorf("invalid depthStr %s", err.Error()))
				return
			}

			i.showProgressWithMessage(fmt.Sprintf("Buying a postage batch (depth %s, amount %s)", depthStr, amountStr))
			_, id, err := i.bl.BuyStamp(amount, depth, "", false)
			// just stand by
			<-time.After(time.Second * 30)
			if err != nil {
				i.hideProgress()
				i.showError(err)
				return
			}
			i.hideProgress()
			batchRadio.Append(shortenHashOrAddress(hex.EncodeToString(id)))
		})

		content.Objects = []fyne.CanvasObject{container.NewBorder(buyBatchContent, container.NewVBox(buyButton), nil, nil)}
		child.SetContent(content)
		child.Show()
	})
	button.Hidden = !showUpload

	return button
}

func (i *index) buyBatchDialog(depthStr, amountStr *string) fyne.CanvasObject {
	depthBind := binding.BindString(depthStr)
	amountBind := binding.BindString(amountStr)

	amountEntry := widget.NewEntryWithData(amountBind)
	amountEntry.OnChanged = func(s string) {
		amountBind.Set(s)
	}

	depthEntry := widget.NewEntryWithData(depthBind)
	depthEntry.OnChanged = func(s string) {
		depthBind.Set(s)
	}

	optionsForm := widget.NewForm()
	optionsForm.Append(
		"Depth",
		container.NewStack(depthEntry),
	)

	optionsForm.Append(
		"Amount",
		container.NewStack(amountEntry),
	)

	return container.NewStack(optionsForm)
}
