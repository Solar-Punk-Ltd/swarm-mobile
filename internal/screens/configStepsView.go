package screens

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethersphere/bee/pkg/api"
)

type nodeConfig struct {
	path string
	password string
	welcomeMessage string
	swapEnable bool
	natAddress string
	rpcEndpoint string
}

func (i *index) showPasswordView() fyne.CanvasObject {
	i.intro.SetText("Initialise your swarm node with a strong password")
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
	welcomeMessageEntry.SetPlaceHolder(defaultWelcomeMsg)

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
	i.intro.SetText("Choose the type of your node")
	content := container.NewStack()
	swapEnableRadio := widget.NewRadioGroup(
		[]string{api.LightMode.String(), api.UltraLightMode.String()},
		func(mode string) {
			if mode == "Light" {
				i.nodeConfig.swapEnable = true
			} else {
				i.nodeConfig.swapEnable = false
			}
			i.logger.Log(fmt.Sprintf("Node mode selected: %s", mode))
		},
	)
	// default to ultra-light
	swapEnableRadio.SetSelected(api.UltraLightMode.String())
	nextButton := widget.NewButton("Next", func() {
		if swapEnableRadio.Selected == "" {
			i.showError(fmt.Errorf("please select the node mode"))
			return
		}

		i.logger.Log(fmt.Sprintf("SWAP enable: %t, running in %s mode", i.nodeConfig.swapEnable, swapEnableRadio.Selected))
		content.Objects = []fyne.CanvasObject{i.showNATAddressView()}
		content.Refresh()
	})

	backButton := widget.NewButton("Back", func() {
		i.nodeConfig.swapEnable = false
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
	i.intro.SetText("Set the NAT Address of your swarm node (optional)")
	content := container.NewStack()
	natAdrrEntry := widget.NewEntry()
	natAdrrEntry.SetPlaceHolder("123.123.123.123:1634")
	nextButton := widget.NewButton("Next", func() {
		if natAdrrEntry.Text == "" {
			i.logger.Log("NAT address is blank")
		} else {
			i.logger.Log(fmt.Sprintf("Using NAT address: %s", natAdrrEntry.Text))
		}
		i.nodeConfig.natAddress = natAdrrEntry.Text

		// in ultra-light mode no RPC endpoint is necessary
		if i.nodeConfig.swapEnable {
			content.Objects = []fyne.CanvasObject{i.showRPCView()}
		} else {
			content.Objects = []fyne.CanvasObject{i.showStartView()}
		}

		content.Refresh()
	})

	backButton := widget.NewButton("Back", func() {
		i.nodeConfig.natAddress = ""
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
	rpcEntry.SetPlaceHolder(fmt.Sprintf("RPC Endpoint (%s)", defaultRPC))

	nextButton := widget.NewButton("Next", func() {
		if i.nodeConfig.swapEnable {
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
		}
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
	i.intro.TextStyle.Bold = true
	content := container.NewStack()

	startButton := widget.NewButton("Start", func() {
		if i.nodeConfig.path == "" ||
			i.nodeConfig.password == "" ||
			(i.nodeConfig.swapEnable && i.nodeConfig.rpcEndpoint == "" ) ||
			(!i.nodeConfig.swapEnable && i.nodeConfig.rpcEndpoint != "" ) {
				i.showError(fmt.Errorf("missing required config field(s)"))
			return
		}

		i.start(i.nodeConfig.path,
				i.nodeConfig.password,
				i.nodeConfig.welcomeMessage,
				i.nodeConfig.natAddress,
				i.nodeConfig.rpcEndpoint,
				i.nodeConfig.swapEnable)
		content.Refresh()
	})

	backButton := widget.NewButton("Back", func() {
		if i.nodeConfig.swapEnable {
			content.Objects = []fyne.CanvasObject{i.showRPCView()}
		} else {
			content.Objects = []fyne.CanvasObject{i.showNATAddressView()}
		}
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
