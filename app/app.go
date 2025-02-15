package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	serverapi "github.com/cosmos/cosmos-sdk/server/api"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7"
	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/keeper"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ibclocalhost "github.com/cosmos/ibc-go/v7/modules/light-clients/09-localhost"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/spf13/cast"

	"github.com/CoreumFoundation/coreum/v4/app/openapi"
	appupgrade "github.com/CoreumFoundation/coreum/v4/app/upgrade"
	appupgradev4 "github.com/CoreumFoundation/coreum/v4/app/upgrade/v4"
	"github.com/CoreumFoundation/coreum/v4/docs"
	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	assetft "github.com/CoreumFoundation/coreum/v4/x/asset/ft"
	assetftkeeper "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	assetnft "github.com/CoreumFoundation/coreum/v4/x/asset/nft"
	assetnftkeeper "github.com/CoreumFoundation/coreum/v4/x/asset/nft/keeper"
	assetnfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/v4/x/auth/ante"
	"github.com/CoreumFoundation/coreum/v4/x/customparams"
	customparamskeeper "github.com/CoreumFoundation/coreum/v4/x/customparams/keeper"
	customparamstypes "github.com/CoreumFoundation/coreum/v4/x/customparams/types"
	"github.com/CoreumFoundation/coreum/v4/x/delay"
	delaykeeper "github.com/CoreumFoundation/coreum/v4/x/delay/keeper"
	delaytypes "github.com/CoreumFoundation/coreum/v4/x/delay/types"
	"github.com/CoreumFoundation/coreum/v4/x/deterministicgas"
	deterministicgastypes "github.com/CoreumFoundation/coreum/v4/x/deterministicgas/types"
	"github.com/CoreumFoundation/coreum/v4/x/dex"
	dexkeeper "github.com/CoreumFoundation/coreum/v4/x/dex/keeper"
	dextypes "github.com/CoreumFoundation/coreum/v4/x/dex/types"
	"github.com/CoreumFoundation/coreum/v4/x/feemodel"
	feemodelkeeper "github.com/CoreumFoundation/coreum/v4/x/feemodel/keeper"
	feemodeltypes "github.com/CoreumFoundation/coreum/v4/x/feemodel/types"
	cnftkeeper "github.com/CoreumFoundation/coreum/v4/x/nft/keeper"
	cnftmodule "github.com/CoreumFoundation/coreum/v4/x/nft/module"
	wasmcustomhandler "github.com/CoreumFoundation/coreum/v4/x/wasm/handler"
	cwasmtypes "github.com/CoreumFoundation/coreum/v4/x/wasm/types"
	"github.com/CoreumFoundation/coreum/v4/x/wbank"
	wbankkeeper "github.com/CoreumFoundation/coreum/v4/x/wbank/keeper"
	"github.com/CoreumFoundation/coreum/v4/x/wibctransfer"
	wibctransferkeeper "github.com/CoreumFoundation/coreum/v4/x/wibctransfer/keeper"
	"github.com/CoreumFoundation/coreum/v4/x/wnft"
	wnftkeeper "github.com/CoreumFoundation/coreum/v4/x/wnft/keeper"
	"github.com/CoreumFoundation/coreum/v4/x/wstaking"
)

const (
	// Name is the blockchain name.
	Name = "core"

	// DefaultChainID is the default chain id of the network.
	DefaultChainID = constant.ChainIDMain
)

// ChosenNetwork is a hacky solution to pass network config
// from cmd package to app.
var ChosenNetwork config.NetworkConfig

var (
	// DefaultNodeHome default home directories for the application daemon.
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		wbank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			[]govclient.ProposalHandler{
				// TODO(v5): Remove once IBC upgrades to the new param management mechanism.
				// Check ibc-go/modules/core/02-client/types/params.go
				paramsclient.ProposalHandler,
				ibcclientclient.UpdateClientProposalHandler,
				ibcclientclient.UpgradeProposalHandler,
			},
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctm.AppModuleBasic{},
		ibchooks.AppModuleBasic{},
		packetforward.AppModuleBasic{},
		ica.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		wibctransfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		consensus.AppModuleBasic{},
		wasm.AppModuleBasic{},
		feemodel.AppModuleBasic{},
		wnft.AppModuleBasic{},
		cnftmodule.AppModuleBasic{},
		assetft.AppModuleBasic{},
		assetnft.AppModuleBasic{},
		customparams.AppModuleBasic{},
		delay.AppModuleBasic{},
		dex.AppModuleBasic{},
	)

	// module account permissions.
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		icatypes.ModuleName:            nil,
		wasmtypes.ModuleName:           {authtypes.Burner},
		assetfttypes.ModuleName:        {authtypes.Minter, authtypes.Burner},
		assetnfttypes.ModuleName:       {authtypes.Burner},
		// the line is required by the nft module to have the module account stored in the account keeper
		nft.ModuleName: {},
	}
)

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	AuthzKeeper      authzkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    *stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	// IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCKeeper              *ibckeeper.Keeper
	IBCHooksKeeper         ibchookskeeper.Keeper
	PacketForwardKeeper    *packetforwardkeeper.Keeper
	ICAHostKeeper          icahostkeeper.Keeper
	ICAControllerKeeper    icacontrollerkeeper.Keeper
	TransferKeeper         wibctransferkeeper.TransferKeeperWrapper
	EvidenceKeeper         evidencekeeper.Keeper
	FeeGrantKeeper         feegrantkeeper.Keeper
	ConsensusParamsKeeper  consensusparamkeeper.Keeper
	WasmKeeper             wasmkeeper.Keeper
	WasmPermissionedKeeper *wasmkeeper.PermissionedKeeper
	GroupKeeper            groupkeeper.Keeper

	AssetFTKeeper      assetftkeeper.Keeper
	AssetNFTKeeper     assetnftkeeper.Keeper
	FeeModelKeeper     feemodelkeeper.Keeper
	BankKeeper         wbankkeeper.BaseKeeperWrapper
	NFTKeeper          wnftkeeper.Wrapper
	CustomParamsKeeper customparamskeeper.Keeper
	DelayKeeper        delaykeeper.Keeper
	DEXKeeper          dexkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedWASMKeeper          capabilitykeeper.ScopedKeeper

	// ModuleManager is the module manager
	ModuleManager *module.Manager

	// sm is the simulation manager
	sm *module.SimulationManager

	configurator module.Configurator

	// IBC Hooks.
	Ics20WasmHooks   *ibchooks.WasmHooks
	HooksICS4Wrapper ibchooks.ICS4Middleware
}

// New returns a reference to an initialized blockchain app.
//
//nolint:funlen // Disable linting for code generated by Cosmos SDK
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	deterministicGasConfig := deterministicgas.DefaultConfig()
	encodingConfig := config.NewEncodingConfig(ModuleBasics)
	appCodec := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	txConfig := encodingConfig.TxConfig
	interfaceRegistry := encodingConfig.InterfaceRegistry
	// Since 0.47 all ibc clients must be registered explicitly and are not registered automatically.
	// we need to register localhost client since we use cosmos relayer in our integration tests and it
	// relies on localhost client to be registered.
	ibclocalhost.RegisterInterfaces(interfaceRegistry)

	bApp := baseapp.NewBaseApp(Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, authz.ModuleName, banktypes.StoreKey,
		stakingtypes.StoreKey, crisistypes.StoreKey, minttypes.StoreKey,
		distrtypes.StoreKey, slashingtypes.StoreKey, govtypes.StoreKey,
		paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, capabilitytypes.StoreKey, consensusparamtypes.StoreKey,
		wasmtypes.StoreKey, feemodeltypes.StoreKey, assetfttypes.StoreKey,
		assetnfttypes.StoreKey, nftkeeper.StoreKey, ibcexported.StoreKey,
		ibctransfertypes.StoreKey, ibchookstypes.StoreKey, packetforwardtypes.StoreKey,
		icahosttypes.StoreKey, icacontrollertypes.StoreKey, delaytypes.StoreKey,
		customparamstypes.StoreKey, group.StoreKey, dextypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey, feemodeltypes.TransientStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		txConfig:          txConfig,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		keys[consensusparamtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	bApp.SetParamStore(&app.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.ScopedICAHostKeeper = app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.ScopedICAControllerKeeper = app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.ScopedWASMKeeper = app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	app.CapabilityKeeper.Seal()

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		ChosenNetwork.Provider.GetAddressPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	delayRouter := delaytypes.NewRouter()
	app.DelayKeeper = delaykeeper.NewKeeper(appCodec, keys[delaytypes.StoreKey], delayRouter, app.interfaceRegistry)

	originalBankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.ModuleAccountAddrs(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.WasmPermissionedKeeper = wasmkeeper.NewGovPermissionKeeper(&app.WasmKeeper)
	app.AssetFTKeeper = assetftkeeper.NewKeeper(
		appCodec,
		keys[assetfttypes.StoreKey],
		// for the assetft we use the clear bank keeper without the assets integration to prevent cycling calls.
		originalBankKeeper,
		app.DelayKeeper,
		// pointer is used here because there is cycle in keeper dependencies:
		// AssetFTKeeper -> WasmKeeper -> BankKeeper -> AssetFTKeeper
		&app.WasmKeeper,
		app.WasmPermissionedKeeper,
		&app.AccountKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	err := delayRouter.RegisterHandler(
		&assetfttypes.DelayedTokenUpgradeV1{},
		assetfttypes.NewTokenUpgradeV1Handler(app.AssetFTKeeper),
	)
	if err != nil {
		panic(err)
	}

	app.BankKeeper = wbankkeeper.NewKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		// pointer is used here because there is cycle in keeper dependencies:
		// AssetFTKeeper -> WasmKeeper -> BankKeeper -> AssetFTKeeper
		&app.WasmKeeper,
		app.ModuleAccountAddrs(),
		app.AssetFTKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[slashingtypes.StoreKey],
		app.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		keys[crisistypes.StoreKey],
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))

	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	app.FeeModelKeeper = feemodelkeeper.NewKeeper(
		keys[feemodeltypes.StoreKey],
		tkeys[feemodeltypes.TransientStoreKey],
		appCodec,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.CustomParamsKeeper = customparamskeeper.NewKeeper(
		keys[customparamstypes.StoreKey],
		appCodec,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.IBCKeeper = ibckeeper.NewKeeper(appCodec, keys[ibcexported.StoreKey], app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper, app.UpgradeKeeper, app.ScopedIBCKeeper)

	nftKeeper := nftkeeper.NewKeeper(keys[nftkeeper.StoreKey], appCodec, app.AccountKeeper, app.BankKeeper)
	app.AssetNFTKeeper = assetnftkeeper.NewKeeper(
		appCodec,
		keys[assetnfttypes.StoreKey],
		nftKeeper,
		// for the assetnft we use the clear bank keeper without the assets integration
		// because it interacts only with native token.
		originalBankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.NFTKeeper = wnftkeeper.NewWrappedNFTKeeper(nftKeeper, app.AssetNFTKeeper)

	// IBC Hooks.
	// The contract WASM keeper needs to be set later since it depends on WASM hooks.
	wasmHooks := ibchooks.NewWasmHooks(&app.IBCHooksKeeper, nil, ChosenNetwork.Provider.GetAddressPrefix())
	app.Ics20WasmHooks = &wasmHooks
	app.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		app.IBCKeeper.ChannelKeeper,
		app.Ics20WasmHooks,
	)

	// Packet Forward Middleware.
	app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		keys[packetforwardtypes.StoreKey],
		nil, // will be zero-value here, reference is set later on with SetTransferKeeper.
		app.IBCKeeper.ChannelKeeper,
		app.DistrKeeper,
		app.BankKeeper,
		app.HooksICS4Wrapper, // Wrap IBC hooks with PFM.
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create Transfer Keepers
	app.TransferKeeper = wibctransferkeeper.NewTransferKeeperWrapper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.PacketForwardKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.ScopedTransferKeeper,
	)

	app.PacketForwardKeeper.SetTransferKeeper(app.TransferKeeper)

	app.IBCHooksKeeper = ibchookskeeper.NewKeeper(
		app.keys[ibchookstypes.StoreKey],
	)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		app.keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.HooksICS4Wrapper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.ScopedICAHostKeeper,
		bApp.MsgServiceRouter(),
	)

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec,
		app.keys[icacontrollertypes.StoreKey],
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.HooksICS4Wrapper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.ScopedICAControllerKeeper,
		bApp.MsgServiceRouter(),
	)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		// TODO(v5): Remove once IBC upgrades to the new param management mechanism.
		// Check ibc-go/modules/core/02-client/types/params.go
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper))

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	govConfig := govtypes.DefaultConfig()
	/*
		Example of setting gov params:
		govConfig.MaxMetadataLen = 10000
	*/
	govKeeper := govkeeper.NewKeeper(
		appCodec,
		keys[govtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	groupConfig := group.DefaultConfig()
	/*
		Example of setting group params:
		groupConfig.MaxExecutionPeriod = 2 * time.Hour * 24 // 2 days
		groupConfig.MaxMetadataLen = 1000
	*/
	app.GroupKeeper = groupkeeper.NewKeeper(
		keys[group.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		groupConfig,
	)

	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	wasmDir := filepath.Join(homePath, "wasm-data")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(errors.Wrapf(err, "error while reading wasm config"))
	}

	wasmOpts := []wasmkeeper.Option{
		wasmkeeper.WithAcceptedAccountTypesOnContractInstantiation(
			&authtypes.BaseAccount{},
			&vestingtypes.ContinuousVestingAccount{},
			&vestingtypes.DelayedVestingAccount{},
			&vestingtypes.PeriodicVestingAccount{},
			&vestingtypes.BaseVestingAccount{},
			&vestingtypes.PermanentLockedAccount{},
		),
		wasmkeeper.WithAccountPruner(cwasmtypes.AccountPruner{}),
		wasmkeeper.WithCoinTransferrer(cwasmtypes.NewBankCoinTransferrer(app.BankKeeper)),
		wasmkeeper.WithMessageHandler(wasmcustomhandler.NewMessengerWrapper(wasmkeeper.NewDefaultMessageHandler(
			app.MsgServiceRouter(),
			app.IBCKeeper.ChannelKeeper,
			app.IBCKeeper.ChannelKeeper,
			app.ScopedWASMKeeper,
			app.BankKeeper,
			appCodec,
			&app.TransferKeeper,
			wasmcustomhandler.NewCoreumMsgHandler(),
		))),
		wasmkeeper.WithQueryPlugins(wasmcustomhandler.NewCoreumQueryHandler(
			assetftkeeper.NewQueryService(app.AssetFTKeeper, app.BankKeeper),
			assetnftkeeper.NewQueryService(app.AssetNFTKeeper),
			app.NFTKeeper,
		)),
	}

	if cast.ToBool(appOpts.Get("telemetry.enabled")) {
		wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
	}

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	// See https://github.com/CosmWasm/cosmwasm/blob/main/docs/CAPABILITIES-BUILT-IN.md
	availableCapabilities := "iterator,staking,stargate,cosmwasm_1_1,cosmwasm_1_2,cosmwasm_1_3,cosmwasm_1_4"

	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		keys[wasmtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.ScopedWASMKeeper,
		&app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		availableCapabilities,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	// Set WASM keeper in WASM hooks.
	app.Ics20WasmHooks.ContractKeeper = &app.WasmKeeper

	// TODO(v5): drop once we drop gov v1beta1 compatibility.
	// Set legacy router for backwards compatibility with gov v1beta1
	app.GovKeeper.SetLegacyRouter(govRouter)

	// IBC transfer stack contains (from top to bottom):
	// - wibctransfer
	// - packetforward
	// - ibchooks
	// - ibctransfer
	var ibcTransferStack ibcporttypes.IBCModule
	ibcTransferStack = transfer.NewIBCModule(app.TransferKeeper.Keeper)
	ibcTransferStack = ibchooks.NewIBCMiddleware(ibcTransferStack, &app.HooksICS4Wrapper)
	ibcTransferStack = packetforward.NewIBCMiddleware(
		ibcTransferStack,
		app.PacketForwardKeeper,
		0,
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
	)
	ibcTransferStack = wibctransfer.NewPurposeMiddleware(ibcTransferStack)

	// Create ICAHost Stack
	icaHostStack := icahost.NewIBCModule(app.ICAHostKeeper)

	// Create Interchain Accounts Controller Stack
	icaControllerStack := icacontroller.NewIBCMiddleware(nil, app.ICAControllerKeeper)

	ibcWasmStack := wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, ibcTransferStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(wasmtypes.ModuleName, ibcWasmStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	app.DEXKeeper = dexkeeper.NewKeeper(
		appCodec,
		keys[dextypes.StoreKey],
	)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	assetFTModule := assetft.NewAppModule(
		appCodec,
		app.AssetFTKeeper,
		app.AccountKeeper,
		app.BankKeeper.BaseKeeper,
		app.ParamsKeeper,
	)
	assetNFTModule := assetnft.NewAppModule(
		appCodec,
		app.AssetNFTKeeper,
		app.NFTKeeper.Keeper,
		app.WasmKeeper,
		app.ParamsKeeper,
	)
	feeModule := feemodel.NewAppModule(app.FeeModelKeeper, app.ParamsKeeper)

	wnftModule := wnft.NewAppModule(appCodec, app.NFTKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry)

	customParamsModule := customparams.NewAppModule(app.CustomParamsKeeper, app.ParamsKeeper)
	wstakingModule := wstaking.NewAppModule(
		appCodec,
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(stakingtypes.ModuleName),
		app.CustomParamsKeeper,
	)

	delayModule := delay.NewAppModule(app.DelayKeeper)
	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx, txConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		wbank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(
			appCodec,
			app.SlashingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(slashingtypes.ModuleName),
		),
		distr.NewAppModule(
			appCodec, app.DistrKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(distrtypes.ModuleName),
		),
		wstakingModule,
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		wibctransfer.NewAppModule(app.TransferKeeper),
		packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),
		wasm.NewAppModule(
			appCodec,
			&app.WasmKeeper,
			app.StakingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.MsgServiceRouter(),
			app.GetSubspace(wasmtypes.ModuleName),
		),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		feeModule,
		assetFTModule,
		assetNFTModule,
		wnftModule,
		customParamsModule,
		delayModule,
		dex.NewAppModule(appCodec, app.DEXKeeper),
		// always be last to make sure that it checks for all invariants and not only part of them
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.ModuleManager.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		customparamstypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		consensusparamtypes.ModuleName,
		ibchookstypes.ModuleName,
		packetforwardtypes.ModuleName,
		icatypes.ModuleName,
		wasmtypes.ModuleName,
		feemodeltypes.ModuleName,
		assetfttypes.ModuleName,
		assetnfttypes.ModuleName,
		nft.ModuleName,
		delaytypes.ModuleName,
		dextypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		customparamstypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		vestingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		consensusparamtypes.ModuleName,
		ibchookstypes.ModuleName,
		packetforwardtypes.ModuleName,
		icatypes.ModuleName,
		wasmtypes.ModuleName,
		feemodeltypes.ModuleName,
		assetfttypes.ModuleName,
		assetnfttypes.ModuleName,
		nft.ModuleName,
		delaytypes.ModuleName,
		dextypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	genesisModuleOrder := []string{
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		customparamstypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibcexported.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		packetforwardtypes.ModuleName,
		icatypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		consensusparamtypes.ModuleName,
		ibchookstypes.ModuleName,
		wasmtypes.ModuleName,
		feemodeltypes.ModuleName,
		nft.ModuleName,
		assetfttypes.ModuleName,
		assetnfttypes.ModuleName,
		delaytypes.ModuleName,
		dextypes.ModuleName,
	}

	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(genesisModuleOrder...)

	app.ModuleManager.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(
		app.appCodec,
		deterministicgastypes.NewDeterministicMsgServer(
			app.MsgServiceRouter(),
			deterministicGasConfig,
			app.AssetFTKeeper,
		),
		app.GRPCQueryRouter(),
	)
	app.ModuleManager.RegisterServices(app.configurator)

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	// TODO (v5): remove cnftModule.RegisterServices alongside the module when we drop deprecated handlers of the module.
	cnftmodule.
		NewAppModule(app.AppCodec(), cnftkeeper.NewKeeper(app.NFTKeeper)).
		RegisterServices(app.configurator)

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	// testdata_pulsar.RegisterQueryServer(app.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(
			app.appCodec,
			app.AccountKeeper,
			authsims.RandomGenesisAccounts,
			app.GetSubspace(authtypes.ModuleName),
		),
	}

	// exclude the nft simulation since it is incompatible with the asset nft
	simModules := excludeModules(app.ModuleManager.Modules, []string{nft.ModuleName})
	app.sm = module.NewSimulationManagerFromAppModules(simModules, overrideModules)
	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			HandlerOptions: authante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				SignModeHandler: txConfig.SignModeHandler(),
				FeegrantKeeper:  app.FeeGrantKeeper,
				SigGasConsumer:  authante.DefaultSigVerificationGasConsumer,
			},
			DeterministicGasConfig: deterministicGasConfig,
			IBCKeeper:              app.IBCKeeper,
			GovKeeper:              &app.GovKeeper,
			FeeModelKeeper:         app.FeeModelKeeper,
			WasmTXCounterStoreKey:  keys[wasmtypes.StoreKey],
			WasmConfig:             wasmConfig,
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	// In v0.46, the SDK introduces _postHandlers_. PostHandlers are like
	// antehandlers, but are run _after_ the `runMsgs` execution. They are also
	// defined as a chain, and have the same signature as antehandlers.
	//
	// In baseapp, postHandlers are run in the same store branch as `runMsgs`,
	// meaning that both `runMsgs` and `postHandler` state will be committed if
	// both are successful, and both will be reverted if any of the two fails.
	//
	// The SDK exposes a default postHandlers chain, which comprises of only
	// one decorator: the Transaction Tips decorator. However, some chains do
	// not need it by default, so feel free to comment the next line if you do
	// not need tips.
	// To read more about tips:
	// https://docs.cosmos.network/main/core/tips.html
	//
	// Please note that changing any of the anteHandler or postHandler chain is
	// likely to be a state-machine breaking change, which needs a coordinated
	// upgrade.
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)

	// must be before Loading version
	// requires the snapshot store to be created and registered as a BaseAppOption
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			panic(errors.Wrapf(err, "failed to register wasm snapshot extension"))
		}
	}

	/**** Upgrades ****/
	upgrades := []appupgrade.Upgrade{
		appupgradev4.New(app.ModuleManager, app.configurator),
	}

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(errors.Errorf("failed to read upgrade info from disk %s", err))
	}

	isSkipHeight := app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height)

	// register the upgrades
	for _, upgradeItem := range upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgradeItem.Name,
			upgradeItem.Upgrade,
		)

		if upgradeInfo.Name == upgradeItem.Name && !isSkipHeight {
			// The line below is essential. If `&upgradeItem.StoreUpgrades` is passed to `UpgradeStoreLoader`
			// directly, then due to how `for` loop works in go, the `StoreUpgrades` of the last defined upgrade plan is
			// always used. To overcome this, here we make a copy of the store upgrades before taking a pointer.
			storeUpgrades := upgradeItem.StoreUpgrades
			// configure store loader that checks if version == upgradeHeight and applies store upgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// Initialize pinned codes in wasmvm as they are not persisted there
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(errors.Wrapf(err, "failed initialize wasmp pinned codes").Error())
		}
	}

	return app
}

// Name returns the name of the App.
func (app *App) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application.
func (app *App) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// BeginBlocker application updates every begin block.
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.ModuleManager.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block.
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.ModuleManager.EndBlock(ctx, req)
}

// Configurator returns the app Configurator.
func (app *App) Configurator() module.Configurator {
	return app.configurator
}

// InitChainer application update at chain initialization.
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height.
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry.
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns TxConfig.
func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *App) DefaultGenesis() map[string]json.RawMessage {
	return ModuleBasics.DefaultGenesis(app.appCodec)
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface.
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *serverapi.Server, _ serverconfig.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// TODO(v5) remove alongside the cnft module
	// Regsiter cnft routes.
	// We register the tx and query handlers here, since we don't want to introduce a new module to the
	// list of app.Modules where we have to handle genesis registration and migraitons. we only need to
	// keep these deprecated handlers around to give time to users to migrate.
	cnftmodule.
		NewAppModule(app.AppCodec(), cnftkeeper.NewKeeper(app.NFTKeeper)).
		RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register app's OpenAPI routes.
	apiSvr.Router.Handle("/static/openapi.json", http.FileServer(http.FS(docs.Docs)))
	apiSvr.Router.HandleFunc("/", openapi.Handler(Name, "/static/openapi.json"))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// RegisterNodeService registers the app node service.
func (app *App) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// GetMaccPerms returns a copy of the module account permissions
//
// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}

	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces.
func initParamsKeeper(
	appCodec codec.BinaryCodec,
	legacyAmino *codec.LegacyAmino,
	key, tkey storetypes.StoreKey,
) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// TODO(v5): Remove once all of modules migrate to the new param management mechanism.
	// For now we keep for legacy proposals to work properly.
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName).WithKeyTable(packetforwardtypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(wasmtypes.ModuleName)
	paramsKeeper.Subspace(feemodeltypes.ModuleName)
	paramsKeeper.Subspace(customparamstypes.CustomParamsStaking)
	paramsKeeper.Subspace(assetfttypes.ModuleName)
	paramsKeeper.Subspace(assetnfttypes.ModuleName)

	return paramsKeeper
}

func excludeModules(modules map[string]interface{}, modulesToExclude []string) map[string]interface{} {
	filteredModules := make(map[string]interface{}, 0)
	modulesToExcludeMap := lo.SliceToMap(modulesToExclude, func(k string) (string, struct{}) {
		return k, struct{}{}
	})
	for n, m := range modules {
		if _, ok := modulesToExcludeMap[n]; ok {
			continue
		}
		filteredModules[n] = m
	}

	return filteredModules
}
