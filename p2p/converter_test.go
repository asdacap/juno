package p2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nsf/jsondiff"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/juno/blockchain"
	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/juno/db"
	"github.com/NethermindEth/juno/db/pebble"
	"github.com/NethermindEth/juno/encoder"
	"github.com/NethermindEth/juno/utils"
)

type mapClassProvider struct {
	classes             map[felt.Felt]*core.DeclaredClass
	providerToIntercept ClassProvider
}

func (c *mapClassProvider) GetClass(hash *felt.Felt) (*core.DeclaredClass, error) {
	if c.providerToIntercept != nil {
		cls, err := c.providerToIntercept.GetClass(hash)
		if err == nil {
			c.classes[*hash] = cls
		}

		return cls, err
	}

	class, ok := c.classes[*hash]
	if !ok {
		return nil, fmt.Errorf("unable to get class of hash %s", hash.String())
	}
	return class, nil
}

func (c *mapClassProvider) Intercept(provider ClassProvider) {
	c.providerToIntercept = provider
}

func (c *mapClassProvider) Load() {
	bytes, err := os.ReadFile("converter_tests/classes.dat")
	if err != nil {
		panic(err)
	}

	err = encoder.Unmarshal(bytes, &c.classes)
	if err != nil {
		panic(err)
	}
}

func (c *mapClassProvider) Save() {
	bytes, err := encoder.Marshal(c.classes)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("converter_tests/classes.dat", bytes, 0o666)
	if err != nil {
		panic(err)
	}
}

// Used to dump blocks
//
//nolint:all
func dumpBlock(blockNum uint64, d db.DB) {
	bc := blockchain.New(d, utils.MAINNET, utils.NewNopZapLogger()) // Needed because class loader need encoder to be registered

	block, err := bc.BlockByNumber(blockNum)
	if err != nil {
		panic(err)
	}

	asbyte, err := encoder.Marshal(block)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(fmt.Sprintf("converter_tests/blocks/%d.dat", blockNum), asbyte, 0o666)
	if err != nil {
		panic(err)
	}
}

func TestEncodeDecodeBlocks(t *testing.T) {
	interceptClassDB, _ := os.LookupEnv("P2P_TEST_SOURCE_DB")
	var d db.DB
	if interceptClassDB != "" {
		d, _ = pebble.New(interceptClassDB, utils.NewNopZapLogger())
	} else {
		d, _ = pebble.NewMem()
	}
	bc := blockchain.New(d, utils.MAINNET, utils.NewNopZapLogger()) // Needed because class loader need encoder to be registered

	classProvider := &mapClassProvider{
		classes: map[felt.Felt]*core.DeclaredClass{},
	}

	if interceptClassDB != "" {
		classProvider.Intercept(&blockchainClassProvider{blockchain: bc})
	} else {
		classProvider.Load()
	}

	c := converter{
		classprovider: classProvider,
	}

	globed, err := filepath.Glob("converter_tests/blocks/*dat")
	if err != nil {
		panic(err)
	}

	for _, filename := range globed {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		decoder := encoder.NewDecoder(f)
		block := core.Block{}
		err = decoder.Decode(&block)
		if err != nil {
			panic(err)
		}

		t.Run(filename, func(t *testing.T) {
			runEncodeDecodeBlockTest(t, &c, bc.Network(), &block)
		})
	}

	if interceptClassDB != "" {
		classProvider.Save()
	}
}

func runEncodeDecodeBlockTest(t *testing.T, c *converter, network utils.Network, originalBlock *core.Block) {
	err := testBlockEncoding(originalBlock, c, network, false)
	if err != nil {
		t.Fatalf("error on block encoding test %s", err)
	}
}

// This was originally meant to be a util code that run on all blocks.
func testBlockEncoding(originalBlock *core.Block, c *converter, network utils.Network, saveOnFailure bool) error {
	originalBlock.ProtocolVersion = ""

	protoheader, err := c.coreBlockToProtobufHeader(originalBlock)
	if err != nil {
		return err
	}

	protoBody, err := c.coreBlockToProtobufBody(originalBlock)
	if err != nil {
		return err
	}

	newCoreBlock, classes, err := c.protobufHeaderAndBodyToCoreBlock(protoheader, protoBody)
	if err != nil {
		return err
	}

	for key, class := range classes {
		declaredClass, err := c.classprovider.GetClass(&key)
		if err != nil {
			return err
		}

		currentClass := declaredClass.Class
		if v, ok := currentClass.(*core.Cairo1Class); ok {
			panic(fmt.Sprintf("Got cairo1 at block %d\n", originalBlock.Number))
			v.Compiled = nil
		}

		if !compareAndPrintDiff(currentClass, class) {
			return errors.New("class mismatch")
		}
	}

	newCoreBlock.ProtocolVersion = ""

	normalizeBlock(originalBlock)
	normalizeBlock(newCoreBlock)

	gatewayjson, err := json.MarshalIndent(originalBlock, "", "    ")
	if err != nil {
		return err
	}

	reencodedblockjson, err := json.MarshalIndent(newCoreBlock, "", "    ")
	if err != nil {
		return err
	}

	if !bytes.Equal(gatewayjson, reencodedblockjson) {

		updateBytes, err := encoder.Marshal(originalBlock)
		if err != nil {
			return err
		}

		if saveOnFailure {
			err = os.WriteFile(fmt.Sprintf("p2p/converter_tests/blocks/%d.dat", originalBlock.Number), updateBytes, 0o666)
			if err != nil {
				return err
			}
		}

		for i, receipt := range originalBlock.Receipts {
			tx := originalBlock.Transactions[i]

			tx2 := newCoreBlock.Transactions[i]
			receipt2 := newCoreBlock.Receipts[i]

			if !compareAndPrintDiff(tx, tx2) {
				return errors.New("tx mismatch.")
			}

			if !compareAndPrintDiff(receipt, receipt2) {
				return errors.New("receipt mismatch")
			}
		}

		txCommit, err := originalBlock.CalculateTransactionCommitment()
		if err != nil {
			return err
		}

		eCommit, err := originalBlock.CalculateEventCommitment()
		if err != nil {
			return err
		}

		headeragain, _ := c.coreBlockToProtobufHeader(originalBlock)
		txCommit2 := fieldElementToFelt(headeragain.TransactionCommitment)
		eCommit2 := fieldElementToFelt(headeragain.EventCommitment)
		if !txCommit.Equal(txCommit2) {
			return errors.New("Tx commit not match")
		}
		if !eCommit.Equal(eCommit2) {
			return errors.New("Event commit not match")
		}

		err = core.VerifyBlockHash(originalBlock, network)
		if err != nil {
			return err
		}

		compareAndPrintDiff(originalBlock, newCoreBlock)
		return errors.New("Mismatch")
	}

	return nil
}

func compareAndPrintDiff(item1, item2 interface{}) bool {
	item1json, _ := json.MarshalIndent(item1, "", "    ")
	item2json, _ := json.MarshalIndent(item2, "", "    ")

	opt := jsondiff.DefaultConsoleOptions()
	diff, strdiff := jsondiff.Compare(item1json, item2json, &opt)

	if diff == jsondiff.FullMatch {
		return true
	}

	fmt.Printf("Mismatch\n")
	fmt.Println(strdiff)

	return false
}

func normalizeBlock(originalBlock *core.Block) {
	for _, transaction := range originalBlock.Transactions {
		if invokeTx, ok := transaction.(*core.InvokeTransaction); ok {
			senderAddress := invokeTx.SenderAddress
			if senderAddress == nil {
				senderAddress = invokeTx.ContractAddress
			}
			invokeTx.SenderAddress = senderAddress
			invokeTx.ContractAddress = senderAddress
		}
	}
}

func testStateDiff(stateDiff *core.StateUpdate) error {
	oriBlockHash := stateDiff.BlockHash
	stateDiff.BlockHash = nil
	stateDiff.NewRoot = nil
	stateDiff.OldRoot = nil

	protobuff := coreStateUpdateToProtobufStateUpdate(stateDiff)

	reencodedStateDiff := protobufStateUpdateToCoreStateUpdate(protobuff)

	before, err := encoder.Marshal(stateDiff)
	if err != nil {
		panic(err)
	}
	after, err := encoder.Marshal(reencodedStateDiff)
	if err != nil {
		panic(err)
	}

	if bytes.Equal(before, after) {
		return nil
	}

	updateBytes, err := encoder.Marshal(stateDiff)
	if err != nil {
		return err
	}
	//nolint:gomnd
	err = os.WriteFile(fmt.Sprintf("p2p/converter_tests/state_updates/%s.dat", oriBlockHash.String()), updateBytes, 0o600)
	if err != nil {
		return err
	}

	oriSD := stateDiff.StateDiff
	rSD := reencodedStateDiff.StateDiff

	for key, diffs := range oriSD.StorageDiffs {
		odiff, ok := rSD.StorageDiffs[key]
		if !ok {
			return fmt.Errorf("missing entry %s", key.String())
		}

		if !compareAndPrintDiff(diffs, odiff) {
			return errors.New("Wrong diff")
		}
	}

	if !compareAndPrintDiff(stateDiff.StateDiff.DeclaredV0Classes, reencodedStateDiff.StateDiff.DeclaredV0Classes) {
		return errors.New("Unable to compare")
	}

	if !compareAndPrintDiff(stateDiff.StateDiff.DeclaredV1Classes, reencodedStateDiff.StateDiff.DeclaredV1Classes) {
		return errors.New("Unable to compare")
	}

	if !compareAndPrintDiff(stateDiff.StateDiff.ReplacedClasses, reencodedStateDiff.StateDiff.ReplacedClasses) {
		return errors.New("Unable to compare")
	}

	return errors.New("mismatch")
}

func TestEncodeDecodeStateUpdate(t *testing.T) {
	d, _ := pebble.NewMem()
	blockchain.New(d, utils.MAINNET, utils.NewNopZapLogger()) // So that the encoder get registered
	globed, err := filepath.Glob("converter_tests/state_updates/*dat")
	if err != nil {
		panic(err)
	}

	for i, filename := range globed {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		decoder := encoder.NewDecoder(f)
		stateUpdate := core.StateUpdate{}
		err = decoder.Decode(&stateUpdate)
		if err != nil {
			panic(err)
		}

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			runEncodeDecodeStateUpdateTest(t, &stateUpdate)
		})
	}
}

func runEncodeDecodeStateUpdateTest(t *testing.T, stateUpdate *core.StateUpdate) {
	// Convert original struct to protobuf struct
	err := testStateDiff(stateUpdate)
	if err != nil {
		t.Fatalf("error on state encoding test %s", err)
	}
}
