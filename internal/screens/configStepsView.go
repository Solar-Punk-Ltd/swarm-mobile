package screens

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ethereum/go-ethereum/ethclient"
)

type nodeConfig struct {
	path string
	password string
	welcomeMessage string
	swapEnabled bool
	natAdrress string
	rpcEndpoint string
}

func (i *index) showPasswordView() fyne.CanvasObject {
	content := container.NewStack()
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")
	nextButton := widget.NewButton("Next", func() {
		if passwordEntry.Text == "" {
			i.showError(fmt.Errorf("password cannot be blank"))
			return
		}

		i.nodeConfig.password = passwordEntry.Text
		content.Objects = []fyne.CanvasObject{i.showWelcomeMessageView()}
		content.Refresh()
	})
	nextButton.Importance = widget.HighImportance
	content.Objects = []fyne.CanvasObject{container.NewBorder(passwordEntry, nextButton, nil, nil)}
	i.content = content
	i.view = container.NewBorder(container.NewVBox(i.title, widget.NewSeparator(), i.intro), nil, nil, nil, content)
	i.view.Refresh()

	return content
}

func (i *index) showWelcomeMessageView() fyne.CanvasObject {
	i.intro.SetText("Set your welcome message for your swarm node (optional)")
	content := container.NewStack()
	welcomeMessageEntry := widget.NewEntry()
	welcomeMessageEntry.SetPlaceHolder("Welcome Message")

	nextButton := widget.NewButton("Next", func() {
		if welcomeMessageEntry.Text == "" {
			welcomeMessageEntry.Text = defaultWelcomeMsg
			i.logger.Log(fmt.Sprintf("Welcome message is blank, using default: %s", defaultWelcomeMsg))
		} else {
			i.logger.Log(fmt.Sprintf("Welcome message is: %s", welcomeMessageEntry.Text))
		}

		i.nodeConfig.welcomeMessage = welcomeMessageEntry.Text
		content.Objects = []fyne.CanvasObject{i.showSWAPEnableView()}
		content.Refresh()
	})

	backButton := widget.NewButton("Back", func() {
		i.nodeConfig.welcomeMessage = ""
		content.Objects = []fyne.CanvasObject{i.showPasswordView()}
		content.Refresh()
	})
	backButton.Importance = widget.WarningImportance
	nextButton.Importance = widget.HighImportance
	content.Objects = []fyne.CanvasObject{container.NewBorder(welcomeMessageEntry, container.NewVBox(nextButton,backButton), nil, nil)}
	i.content = content
	i.view = container.NewBorder(container.NewVBox(i.title, widget.NewSeparator(), i.intro), nil, nil, nil, content)
	i.view.Refresh()

	return content
}

func (i *index) showSWAPEnableView() fyne.CanvasObject {
	i.intro.SetText("Choose the type of your node (Ultra-Light by default)")
	content := container.NewStack()
	swapEnableRadio := widget.NewRadioGroup(
		[]string{"Light", "Ultra-Light"},
		func(mode string) {
			if mode == "Light" {
				i.nodeConfig.swapEnabled = true
			} else {
				i.nodeConfig.swapEnabled = false
			}
			i.logger.Log(fmt.Sprintf("Node mode selected: %s", mode))
		},
	)

	swapEnableRadio.Selected = "Ultra-Light"

	nextButton := widget.NewButton("Next", func() {
		i.logger.Log(fmt.Sprintf("SWAP enable: %t, running in %s mode", i.nodeConfig.swapEnabled, swapEnableRadio.Selected))
		content.Objects = []fyne.CanvasObject{i.showNATAddressView()}
		content.Refresh()
	})

	backButton := widget.NewButton("Back", func() {
		i.nodeConfig.swapEnabled = false
		content.Objects = []fyne.CanvasObject{i.showWelcomeMessageView()}
		content.Refresh()
	})
	backButton.Importance = widget.WarningImportance
	nextButton.Importance = widget.HighImportance
	content.Objects = []fyne.CanvasObject{container.NewBorder(swapEnableRadio,  container.NewVBox(nextButton,backButton), nil, nil)}
	i.content = content
	i.view = container.NewBorder(container.NewVBox(i.title, widget.NewSeparator(), i.intro), nil, nil, nil, content)
	i.view.Refresh()

	return content
}

func (i *index) showNATAddressView() fyne.CanvasObject {
	i.intro.SetText("Set your NAT Address for your swarm node (optional)")
	content := container.NewStack()
	natAdrrEntry := widget.NewEntry()
	natAdrrEntry.SetPlaceHolder("NAT Address")

	nextButton := widget.NewButton("Next", func() {
		if natAdrrEntry.Text == "" {
			i.logger.Log("NAT address is blank")
		} else {
			i.logger.Log(fmt.Sprintf("Using NAT address: %s", natAdrrEntry.Text))
		}
		i.nodeConfig.natAdrress = natAdrrEntry.Text
		content.Objects = []fyne.CanvasObject{i.showRPCView()}
		content.Refresh()
	})

	backButton := widget.NewButton("Back", func() {
		i.nodeConfig.natAdrress = ""
		content.Objects = []fyne.CanvasObject{i.showSWAPEnableView()}
		content.Refresh()
	})
	backButton.Importance = widget.WarningImportance
	nextButton.Importance = widget.HighImportance
	content.Objects = []fyne.CanvasObject{container.NewBorder(natAdrrEntry, container.NewVBox(nextButton,backButton), nil, nil)}
	i.content = content
	i.view = container.NewBorder(container.NewVBox(i.title, widget.NewSeparator(), i.intro), nil, nil, nil, content)
	i.view.Refresh()

	return content
}

func (i *index) showRPCView() fyne.CanvasObject {
	i.intro.SetText("Swarm mobile needs a RPC endpoint to start (optional)")
	content := container.NewStack()
	rpcEntry := widget.NewEntry()
	rpcEntry.SetPlaceHolder(fmt.Sprintf("RPC Endpoint (default: %s)", defaultRPC))

	nextButton := widget.NewButton("Next", func() {
		if rpcEntry.Text == "" {
			rpcEntry.Text = defaultRPC
			i.logger.Log(fmt.Sprintf("RPC endpoint is blank, using default RPC: %s", defaultRPC))
		}
		// test endpoint is connectable
		eth, err := ethclient.Dial(rpcEntry.Text)
		if err != nil {
			i.logger.Log(fmt.Sprintf("rpc endpoint: %s", err.Error()))
			i.showError(fmt.Errorf("rpc endpoint is invalid or not reachable"))
			return
		}
		// check connections
		_, err = eth.ChainID(context.Background())
		if err != nil {
			i.logger.Log(fmt.Sprintf("rpc endpoint: %s", err.Error()))
			i.showError(fmt.Errorf("rpc endpoint: %s", err.Error()))
			return
		}
		i.nodeConfig.rpcEndpoint = rpcEntry.Text
		content.Objects = []fyne.CanvasObject{i.showStartView()}
		content.Refresh()
	})

	backButton := widget.NewButton("Back", func() {
		i.nodeConfig.rpcEndpoint = ""
		content.Objects = []fyne.CanvasObject{i.showNATAddressView()}
		content.Refresh()
	})
	backButton.Importance = widget.WarningImportance
	nextButton.Importance = widget.HighImportance
	content.Objects = []fyne.CanvasObject{container.NewBorder(rpcEntry, container.NewVBox(nextButton,backButton), nil, nil)}
	i.content = content
	i.view = container.NewBorder(container.NewVBox(i.title, widget.NewSeparator(), i.intro), nil, nil, nil, content)
	i.view.Refresh()

	return content
}

func (i *index) showStartView() fyne.CanvasObject {
	i.intro.SetText("Start your Swarm node")
	content := container.NewStack()

	startButton := widget.NewButton("Start", func() {
		if i.nodeConfig.path == "" ||
			i.nodeConfig.password == "" ||
			i.nodeConfig.rpcEndpoint == "" {
				i.showError(fmt.Errorf("missing required config field(s)"))
			return
		}

		i.start(i.nodeConfig.path,
				i.nodeConfig.password,
				i.nodeConfig.welcomeMessage,
				i.nodeConfig.natAdrress,
				i.nodeConfig.rpcEndpoint,
				i.nodeConfig.swapEnabled)
		content.Refresh()
	})

	backButton := widget.NewButton("Back", func() {
		content.Objects = []fyne.CanvasObject{i.showRPCView()}
		content.Refresh()
	})
	backButton.Importance = widget.WarningImportance
	startButton.Importance = widget.HighImportance

	content.Objects = []fyne.CanvasObject{container.NewBorder(startButton, backButton, nil, nil)}
	i.content = content
	i.view = container.NewBorder(container.NewVBox(i.title, widget.NewSeparator(), i.intro), nil, nil, nil, content)
	i.view.Refresh()

	return content
}
