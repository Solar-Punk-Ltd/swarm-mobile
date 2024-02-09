package screens

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"runtime/debug"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	beelite "github.com/Solar-Punk-Ltd/bee-lite"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethersphere/bee/pkg/api"
	"github.com/ethersphere/bee/pkg/swarm"
)

const (
	TestnetChainID    = 5 //testnet
	MainnetChainID    = 100
	MainnetNetworkID  = uint64(1)
	defaultRPC        = "https://gnosis.publicnode.com"
	defaultWelcomeMsg = "Welcome from Swarm Mobile by Solar Punk"
	debugLogLevel	  = "4"
	defaultDepth      = "17"
	defaultAmount     = "100000000"
)

var (
	MainnetBootnodes = []string{
		"/dnsaddr/mainnet.ethswarm.org",
	}

	TestnetBootnodes = []string{
		"/dnsaddr/testnet.ethswarm.org",
	}
)

type logger struct{}

func (*logger) Write(p []byte) (int, error) {
	log.Println(string(p))
	return len(p), nil
}

func (*logger) Log(s string) {
	log.Println(s)
}

type index struct {
	fyne.Window
	app          fyne.App
	view         *fyne.Container
	content      *fyne.Container
	title        *widget.Label
	intro        *widget.Label
	progress     dialog.Dialog
	bl           *beelite.Beelite
	logger       *logger
	beeVersion   string
	nodeConfig	 *nodeConfig
}

type uploadedItem struct {
	Name      string
	Reference string
	Size      int64
	Timestamp int64
	Mimetype  string
}

func Make(a fyne.App, w fyne.Window) fyne.CanvasObject {
	i := &index{
		Window: w,
		app:    a,
		logger: &logger{},
		nodeConfig: &nodeConfig{},
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		i.logger.Log("No build info found")
		i.beeVersion = "unknown"
	} else {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/ethersphere/bee" {
				i.beeVersion = dep.Version
			}
		}
	}

	i.nodeConfig.path = a.Storage().RootURI().Path()
	i.title = widget.NewLabel("Swarm")
	i.title.TextStyle.Bold = true
	i.intro = widget.NewLabel("Initialise your swarm node with a strong password")
	i.intro.Wrapping = fyne.TextWrapWord
	content := container.NewStack()

	// check if any of the configuration is saved
	savedPassword := i.app.Preferences().String("password")
	savedRPCEndpoint := i.app.Preferences().String("rpcEndpoint")
	savedNATAddr := i.app.Preferences().String("natAddress")
	savedWelcomeMsg := i.app.Preferences().String("welcomeMessage")
	savedSwapEnable := i.app.Preferences().Bool("swapEnabled")
	if savedPassword != "" &&
	   savedRPCEndpoint != "" &&
	   savedNATAddr != "" &&
	   savedWelcomeMsg != "" {
		go i.start(i.nodeConfig.path, savedPassword, savedWelcomeMsg, savedNATAddr, savedRPCEndpoint, savedSwapEnable)
		content.Objects = []fyne.CanvasObject{
			container.NewBorder(
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel("Please wait..."),
			),
		}
		i.content = content
		i.view = container.NewBorder(container.NewVBox(i.title, widget.NewSeparator(), i.intro), nil, nil, nil, content)
		return i.view
	}

	i.showPasswordView()

	i.view = container.NewBorder(container.NewVBox(i.title, widget.NewSeparator(), i.intro), nil, nil, nil, i.content)
	return i.view
}

func (i *index) start(path, password, welcomeMessage, natAddress, rpcEndpoint string, swapEnabled bool) {
	if password == "" {
		i.showError(fmt.Errorf("password cannot be blank"))
		return
	}
	i.showProgressWithMessage("Starting Bee")

	err := i.initSwarm(path, path, welcomeMessage, password, natAddress, rpcEndpoint, swapEnabled)
	i.hideProgress()
	if err != nil {
		addr, addrErr := beelite.OverlayAddr(path, password)
		if addrErr != nil {
			i.showError(addrErr)
			return
		}
		i.showErrorWithAddr(addr, err)
		return
	}

	if swapEnabled {
		if i.bl.BeeNodeMode != api.LightMode {
			i.showError(fmt.Errorf("swap is enabled but the current node mode is: %s", i.bl.BeeNodeMode))
			return
		}
	} else if i.bl.BeeNodeMode != api.UltraLightMode {
		i.showError(fmt.Errorf("swap disabled but the current node mode is: %s", i.bl.BeeNodeMode))
		return
	}

	i.app.Preferences().SetString("welcomeMessage", welcomeMessage)
	i.app.Preferences().SetBool("swapEnabled", swapEnabled)
	i.app.Preferences().SetString("natAddress", natAddress)
	i.app.Preferences().SetString("rpcEndpoint", rpcEndpoint)
	err = i.loadView()
	if err != nil {
		i.showError(err)
		return
	}
	i.intro.SetText("")
}

func (i *index) initSwarm(keystore, dataDir, welcomeMessage, password, natAddress, rpcEndpoint string, swapEnabled bool) error {
	i.logger.Log(welcomeMessage)
	i.logger.Log(fmt.Sprintf("bee version: %s", i.beeVersion))

	lo := &beelite.LiteOptions {
		FullNodeMode:              false,
		BootnodeMode:              false,
		Bootnodes:                 MainnetBootnodes,
		DataDir:                   dataDir,
		WelcomeMessage:            welcomeMessage,
		BlockchainRpcEndpoint:     rpcEndpoint,
		SwapInitialDeposit:        "0",
		PaymentThreshold:          "100000000",
		SwapEnable:                swapEnabled,
		ChequebookEnable:          true,
		UsePostageSnapshot:        false,
		DebugAPIEnable:            true,
		Mainnet:                   true,
		NetworkID:                 MainnetNetworkID,
		NATAddr:                   natAddress,
		CacheCapacity:             32 * 1024 * 1024,
		DBOpenFilesLimit:          50,
		DBWriteBufferSize:         32 * 1024 * 1024,
		DBBlockCacheCapacity:      32 * 1024 * 1024,
		DBDisableSeeksCompaction:  false,
		RetrievalCaching:          true,
	}

	bl, err := beelite.Start(lo, password, debugLogLevel)
	if err != nil {
		return err
	}

	i.app.Preferences().SetString("password", password)
	i.bl = bl
	return err
}

func (i *index) loadView() error {
	addrCopyButton := widget.NewButtonWithIcon("   Copy   ", theme.ContentCopyIcon(), func() {
		i.Window.Clipboard().SetContent(i.bl.OverlayEthAddress.String())
	})
	addrHeader := container.NewHBox(widget.NewLabel("Overlay address:"))
	addr := container.NewHBox(
		widget.NewLabel(i.bl.OverlayEthAddress.String()),
		addrCopyButton,
	)
	addrContent := container.NewVBox(addrHeader, addr)

	stampsHeader := container.NewHBox(widget.NewLabel("Postage stamps:"))
	stamps := i.bl.GetAllBatches()
	batchRadio := widget.NewRadioGroup([]string{}, func(s string) {
		if s == "" {
			i.app.Preferences().SetString("selected_stamp", "")
			i.app.Preferences().SetString("batch", "")
			return
		}
		batches := i.bl.GetAllBatches()
		for _, v := range batches {
			stamp := hex.EncodeToString(v.ID())
			if s[0:6] == stamp[0:6] {
				i.app.Preferences().SetString("selected_stamp", s)
				i.app.Preferences().SetString("batch", stamp)
			}
		}
	})

	buyBatchButton := widget.NewButton("Buy a postage batch", func() {
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

	if len(stamps) != 0 {
		selectedStamp := i.app.Preferences().String("selected_stamp")
		for _, v := range stamps {
			batchRadio.Append(shortenHashOrAddress(hex.EncodeToString(v.ID())))
		}

		batchRadio.SetSelected(selectedStamp)
	}
	stampsContent := container.NewVBox(stampsHeader, batchRadio)

	balanceHeader := container.NewHBox(widget.NewLabel("Chequebook balance:"))
	chequebookBalance, err := i.bl.ChequebookBalance()
	if err != nil {
		i.showError(err)
		return err
	}
	balanceStr := chequebookBalance.String()
	balanceBind := binding.BindString(&balanceStr)
	balance := container.NewHBox(widget.NewLabel(balanceStr))
	balanceContent := container.NewVBox(balanceHeader, balance)

	infoCard := widget.NewCard("Info",
				fmt.Sprintf("Connected with %d peers", i.bl.TopologyDriver.Snapshot().Connected),
				container.NewVBox(addrContent, balanceContent, stampsContent, buyBatchButton))
	go func() {
		// auto reload
		for {
			time.Sleep(time.Second * 5)
			if i.bl != nil {
				infoCard.SetSubTitle(fmt.Sprintf("Connected with %d peers", i.bl.TopologyDriver.Snapshot().Connected))
				chequebookBalance, err = i.bl.ChequebookBalance()
				if err != nil {
					balanceBind.Set(err.Error())
				} else {
					balanceBind.Set(chequebookBalance.String())
				}
			}
		}
	}()

	filepath := ""
	mimetype := ""
	var pathBind = binding.BindString(&filepath)
	path := widget.NewEntry()
	path.Bind(pathBind)
	path.Disable()
	var file io.Reader
	openFile := widget.NewButton("File Open", func() {
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
			{Text: "Choose File", Widget: openFile},
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
			batchID := i.app.Preferences().String("batch")
			if batchID == "" {
				i.showError(fmt.Errorf("please select a batch of stamp"))
				return
			}
			i.progress = dialog.NewProgressInfinite("",fmt.Sprintf("Uploading %s", path.Text), i)
			i.progress.Show()

			ref, err := i.bl.AddFileBzz(context.Background(), batchID, path.Text, mimetype, file)
			// just stand by
			<-time.After(time.Second * 60)
			if err != nil {
				i.progress.Hide()
				i.showError(err)
				return
			}
			filename := path.Text
			uploadedSrt := i.app.Preferences().String("uploads")
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
				i.progress.Hide()
				i.showError(err)
				return
			}
			i.app.Preferences().SetString("uploads", string(data))
			d := dialog.NewCustomConfirm("Upload successful", "Ok", "Cancel", i.refDialog(ref.String()), func(b bool) {}, i.Window)
			i.progress.Hide()
			d.Show()
		}()
	}
	listButton := widget.NewButton("All Uploads", func() {
		uploadedContent := container.NewVBox()
		uploadedContentWrapper := container.NewScroll(uploadedContent)
		uploadedSrt := i.app.Preferences().String("uploads")
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
		if size.Width < 200{
			size.Width =200
		}
		if size.Height < 100 {
			size.Height = 100
		}
		child.Resize(size)
		child.SetContent(uploadedContentWrapper)
		child.Show()
	})
	uploadCard := widget.NewCard("Upload", "upload content into swarm", container.NewVBox(upForm, listButton))

	hash := widget.NewEntry()
	hash.SetPlaceHolder("Swarm Hash")
	dlForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Swarm Hash", Widget: hash, HintText: "Swarm Hash"},
		},
		OnSubmit: func() {
			dlAddr, err := swarm.ParseHexAddress(hash.Text)
			if err != nil {
				i.showError(err)
				return
			}
			go func() {
				i.progress = dialog.NewProgressInfinite("", fmt.Sprintf("Downloading %s", shortenHashOrAddress(hash.Text)), i)
				i.progress.Show()
				ref, fileName, err := i.bl.GetBzz(context.Background(), dlAddr)
				if err != nil {
					i.progress.Hide()
					i.showError(err)
					return
				}
				hash.Text = ""
				data, err := io.ReadAll(ref)
				if err != nil {
					i.progress.Hide()
					i.showError(err)
					return
				}
				i.progress.Hide()
				saveFile := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil {
						i.showError(err)
						return
					}
					if writer == nil {
						return
					}
					_, err = writer.Write(data)
					if err != nil {
						i.showError(err)
						return
					}
					writer.Close()
				}, i.Window)
				saveFile.SetFileName(fileName)
				saveFile.Show()
			}()

		},
	}
	downloadCard := widget.NewCard("Download", "download content from swarm", dlForm)
	i.content.Objects = []fyne.CanvasObject{container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		container.NewScroll(container.NewGridWithColumns(1, infoCard, uploadCard, downloadCard))),
	}
	i.content.Refresh()
	return nil
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

func (i *index) refDialog(ref string) fyne.CanvasObject {
	refButton := widget.NewButtonWithIcon("   Copy    ", theme.ContentCopyIcon(), func() {
		i.Window.Clipboard().SetContent(ref)
	})
	return container.NewStack(container.NewBorder(nil, nil, nil, refButton, widget.NewLabel(shortenHashOrAddress(ref))))
}

func (i *index) showProgressWithMessage(message string) {
	i.progress = dialog.NewProgressInfinite("", message, i) //lint:ignore SA1019 fyne-io/fyne/issues/2782
	i.progress.Show()
}

func (i *index) hideProgress() {
	i.progress.Hide()
}

func (i *index) showError(err error) {
	label := widget.NewLabel(err.Error())
	label.Wrapping = fyne.TextWrapWord
	d := dialog.NewCustom("Error", "       Close       ", label, i.Window)
	parentSize := i.Window.Canvas().Size()
	d.Resize(fyne.NewSize(parentSize.Width*90/100, 0))
	d.Show()
}

func (i *index) showErrorWithAddr(addr common.Address, err error) {
	addrStr := shortenHashOrAddress(addr.String())
	addrCopyButton := widget.NewButtonWithIcon("   Copy    ", theme.ContentCopyIcon(), func() {
		i.Window.Clipboard().SetContent(addr.String())
	})
	header := container.NewHBox(widget.NewLabel(addrStr), addrCopyButton)
	label := widget.NewLabel(err.Error())
	label.Wrapping = fyne.TextWrapWord
	content := container.NewBorder(header, label, nil, nil)
	d := dialog.NewCustom("Error", "       Close       ", content, i.Window)
	parentSize := i.Window.Canvas().Size()
	d.Resize(fyne.NewSize(parentSize.Width*90/100, 0))
	d.Show()
}

func shortenHashOrAddress(item string) string {
	return fmt.Sprintf("%s[...]%s", item[0:6], item[len(item)-6:])
}
