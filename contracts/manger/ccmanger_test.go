package manger

import (
	"testing"
	"golang.org/x/net/context"
	"github.com/stretchr/testify/assert"
	"github.com/palletone/go-palletone/contracts/rwset"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/common"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/utils"
	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/palletone/go-palletone/core/vmContractPub/mocks/samplesyscc"
	"github.com/spf13/viper"
	"net"
	"google.golang.org/grpc"
	"time"
	"os"
	"github.com/palletone/go-palletone/contracts/accesscontrol"
)

type mocksupt struct {}

func (*mocksupt) GetTxSimulator(chainid string, txid string) (*rwset.TxSimulator, error) {
	return nil, nil
}
func (*mocksupt) IsSysCC(name string) bool {
	return true
}

func (*mocksupt) Execute(ctxt context.Context, cid, name, version, txid string, syscc bool, signedProp *pb.SignedProposal, prop *pb.Proposal, spec interface{}) (*pb.Response, *pb.ChaincodeEvent, error) {
	return nil, nil, nil
}

func TestChaincodeMgrPeerProcess(t *testing.T) {
	var mksupt Support = &SupportImpl{}
	es := NewEndorserServer(mksupt)
	//signedProp := getSignedProp("ccid", "0", t)
	_, err := es.ProcessProposal(context.Background(), nil, nil, "", "", nil)
	assert.Error(t, err)
}
//
//func singedPro(chid, ccid, ccver string, ccargs [][]byte) *pb.SignedProposal {
//	spec := &pb.ChaincodeSpec{Type: 1, ChaincodeId: &pb.ChaincodeID{Name: ccid, Version: ccver}, Input: &pb.ChaincodeInput{Args: ccargs}}
//
//	cis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}
//
//	creator, err := signer.Serialize()
//	prop, _, err := utils.CreateChaincodeProposal(common.HeaderType_ENDORSER_TRANSACTION, chid, cis, creator)
//	propBytes, err := utils.GetBytesProposal(prop)
//	signature, err := signer.Sign(propBytes)
//
//	return &pb.SignedProposal{ProposalBytes: propBytes, Signature: signature}
//
//
//	sprop, prop := putils.MockSignedEndorserProposalOrPanic(chainID, spec, creator, []byte("msg1"))
//	cccid := ccprovider.NewCCContext(chainID, cdInvocationSpec.ChaincodeSpec.ChaincodeId.Name, version, uuid, false, sprop, prop)
//	retval, ccevt, err = ExecuteWithErrorFilter(ctx, cccid, cdInvocationSpec)
//	if err != nil {
//		return nil, uuid, nil, fmt.Errorf("Error invoking chaincode: %s", err)
//	}
//}
//

func getSignedPropWithCHIdAndArgs(chid, ccid, ccver string, ccargs [][]byte, t *testing.T) *pb.SignedProposal {
	spec := &pb.ChaincodeSpec{Type: 1, ChaincodeId: &pb.ChaincodeID{Name: ccid, Version: ccver}, Input: &pb.ChaincodeInput{Args: ccargs}}
	cis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	//creator, err := signer.Serialize()
	creator := []byte("glh")
	prop, _, err := utils.CreateChaincodeProposal(common.HeaderType_ENDORSER_TRANSACTION, chid, cis, creator)
	assert.NoError(t, err)
	propBytes, err := utils.GetBytesProposal(prop)
	assert.NoError(t, err)

	//todo ,tmp!!!!!!
	signature := propBytes
	//signature, err := signer.Sign(propBytes)
	assert.NoError(t, err)
	return &pb.SignedProposal{ProposalBytes: propBytes, Signature: signature}
}

func TestEndorserDeployExecSysCC(t *testing.T) {
	SysCCMap := make(map[string]struct{})
	deployedCCName := "sample_syscc"
	SysCCMap[deployedCCName] = struct{}{}
	creator := []byte("glh")
	var mksupt Support = &SupportImpl{}

	peerInit()

	t.Logf("TestEndorserDeployExecSysCC run, cc name[%s]", deployedCCName)

	chainID := util.GetTestChainID()
	es := NewEndorserServer(mksupt)

	f := "putval"
	args := util.ToChaincodeArgs(f, "greeting", "hey there")

	//signedProp := getSignedPropWithCHIdAndArgs(util.GetTestChainID(), "lscc", "0", [][]byte{[]byte("deploy"), []byte("a"), cds}, t)
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{Name: deployedCCName},
		Type:        pb.ChaincodeSpec_GOLANG,
		Input:       &pb.ChaincodeInput{Args: args},
	}
	cid := &pb.ChaincodeID{
		Path: "/home/glh/project/pallet/src/common/mocks/samplesyscc/samplesyscc", ///home/glh/project/pallet/src/common/mocks/samplesyscc
		Name: "sample_syscc",
		Version:"ptn001",
	}

	sprop, prop := MockSignedEndorserProposalOrPanic(chainID, spec, creator, []byte("msg1"))
	rsp, err := es.ProcessProposal(context.Background(), sprop, prop, chainID, "txid001", cid)
	if err != nil {
		logger.Errorf("ProcessProposal error[%v]", err)
	}
	logger.Infof("ProcessProposal rsp=%v", rsp)
}

//type oldSysCCInfo struct {
//	origSystemCC       []*scc.SystemChaincode
//	origSysCCWhitelist map[string]string
//}

//func (osyscc *oldSysCCInfo) reset() {
//	scc.MockResetSysCCs(osyscc.origSystemCC)
//	viper.Set("chaincode.system", osyscc.origSysCCWhitelist)
//}

func peerMockInitialize() {
//ledgermgmt.InitializeTestEnvWithCustomProcessors(ConfigTxProcessors)
chains.list = nil
chains.list = make(map[string]*chain)
//chainInitializer = func(string) { return }
}
func peerMockCreateChain(cid string) error {
	chains.Lock()
	defer chains.Unlock()

	chains.list[cid] = &chain{
		cs: &chainSupport{
			//Resources: &mockchannelconfig.Resources{
			//	PolicyManagerVal: &mockpolicies.Manager{
			//		Policy: &mockpolicies.Policy{},
			//	},
			//	ConfigtxValidatorVal: &mockconfigtx.Validator{},
			//},
			//ledger: ledger},
		},
	}

	return nil
}

func peerInitSysCCTests() (*oldSysCCInfo, net.Listener, error) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	viper.Set("peer.fileSystemPath", "/home/glh/tmp/chaincodes")
	viper.Set("peer.address", "127.0.0.1:12345")
	viper.Set("chaincode.executetimeout", 20*time.Second)

	defer os.RemoveAll("/home/glh/tmp/chaincodes")

	peerMockInitialize()

	peerAddress := "0.0.0.0:21726"
	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		return nil, nil, err
	}

	ccStartupTimeout := time.Duration(5000) * time.Millisecond
	ca, _ := accesscontrol.NewCA()
	pb.RegisterChaincodeSupportServer(grpcServer, core.NewChaincodeSupport(peerAddress, false, ccStartupTimeout, ca))

	go grpcServer.Serve(lis)

	//set systemChaincodes to sample
	sysccs := []*scc.SystemChaincode{
		{
			Enabled:   true,
			Name:      "sample_syscc",
			Path:      "/home/glh/project/pallet/src/common/mocks/samplesyscc/samplesyscc",
			InitArgs:  [][]byte{},
			Chaincode: &samplesyscc.SampleSysCC{},
		},
	}

	sysccinfo := &oldSysCCInfo{origSysCCWhitelist: viper.GetStringMapString("chaincode.system")}

	// System chaincode has to be enabled
	viper.Set("chaincode.system", map[string]string{"sample_syscc": "true"})

	sysccinfo.origSystemCC = scc.MockRegisterSysCCs(sysccs)

	/////^^^ system initialization completed ^^^
	return sysccinfo, lis, nil
}

func peerInit() {
	_, _, err := peerInitSysCCTests() //lis
	if err != nil {
		return
	}

	chainID := util.GetTestChainID()
	peerMockCreateChain(chainID)

	scc.DeploySysCCs(chainID)
	//defer scc.DeDeploySysCCs(chainID)
}


func TestExecSysCC(t *testing.T) {
	// System chaincode has to be enabled
	viper.Set("chaincode.system", map[string]string{"sample_syscc": "true"})

	chainID := util.GetTestChainID()
	f := "putval"
	args := util.ToChaincodeArgs(f, "greeting", "hey there")

	Init()
	ContractInvoke(chainID, "sample_syscc",  args)
}
