package screens

import (
	"fmt"
	"log"
	"runtime/debug"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	beelite "github.com/Solar-Punk-Ltd/bee-lite"
	"github.com/ethersphere/bee/pkg/api"
)

const (
	TestnetChainID    = 5 //testnet
	MainnetChainID    = 100
	MainnetNetworkID  = uint64(1)
	defaultRPC        = "https://gnosis.publicnode.com"
	defaultWelcomeMsg = "Welcome from Swarm Mobile by Solar Punk"
	debugLogLevel     = "4"
	defaultDepth      = "17"
	defaultAmount     = "100000000"
	passwordPrefKey   = "password"
	welcomeMessagePrefKey = "welcomeMessage"
	swapEnablePrefKey     = "swapEnable"
	natAddressPrefKey     = "natAddress"
	rpcEndpointPrefKey    = "rpcEndpoint"
	selectedStampPrefKey  = "selected_stamp"
	batchPrefKey          = "batch"
	uploadsPrefKey        = "uploads"
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
	savedPassword := i.app.Preferences().String(passwordPrefKey)
	savedRPCEndpoint := i.app.Preferences().String(rpcEndpointPrefKey)
	savedNATAddr := i.app.Preferences().String(natAddressPrefKey)
	savedWelcomeMsg := i.app.Preferences().String(welcomeMessagePrefKey)
	savedSwapEnable := i.app.Preferences().Bool(swapEnablePrefKey)
	if savedPassword != "" &&
		((savedSwapEnable && savedRPCEndpoint != "" ) ||
		 !savedSwapEnable && savedRPCEndpoint == "") &&
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

func (i *index) start(path, password, welcomeMessage, natAddress, rpcEndpoint string, swapEnable bool) {
	if password == "" {
		i.showError(fmt.Errorf("password cannot be blank"))
		return
	}
	i.showProgressWithMessage("Starting Bee")

	err := i.initSwarm(path, path, welcomeMessage, password, natAddress, rpcEndpoint, swapEnable)
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

	if swapEnable {
		if i.bl.BeeNodeMode != api.LightMode {
			i.showError(fmt.Errorf("swap is enabled but the current node mode is: %s", i.bl.BeeNodeMode))
			return
		}
	} else if i.bl.BeeNodeMode != api.UltraLightMode {
		i.showError(fmt.Errorf("swap disabled but the current node mode is: %s", i.bl.BeeNodeMode))
		return
	}

	i.app.Preferences().SetString(welcomeMessagePrefKey, welcomeMessage)
	i.app.Preferences().SetBool(swapEnablePrefKey, swapEnable)
	i.app.Preferences().SetString(natAddressPrefKey, natAddress)
	i.app.Preferences().SetString(rpcEndpointPrefKey, rpcEndpoint)
	err = i.loadMenuView()
	if err != nil {
		i.showError(err)
		return
	}
	i.intro.SetText("")
}

func (i *index) initSwarm(keystore, dataDir, welcomeMessage, password, natAddress, rpcEndpoint string, swapEnable bool) error {
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
		SwapEnable:                swapEnable,
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

	i.app.Preferences().SetString(passwordPrefKey, password)
	i.bl = bl
	return err
}

func (i *index) loadMenuView() error {
	// only show certain views if the node mode is NOT ultra light
	showUpload := i.bl.BeeNodeMode != api.UltraLightMode
	infoCard := i.showInfoCard(showUpload)
	uploadCard := i.showUploadCard(showUpload)
	downloadCard := i.showDownloadCard()
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