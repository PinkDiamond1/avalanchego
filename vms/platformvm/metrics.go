// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/utils/metric"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/avalanchego/vms/platformvm/transactions"
)

var errUnknownBlockType = errors.New("unknown block type")

type metrics struct {
	percentConnected prometheus.Gauge
	totalStake       prometheus.Gauge

	numAbortBlocks,
	numAtomicBlocks,
	numCommitBlocks,
	numProposalBlocks,
	numStandardBlocks prometheus.Counter

	numAddDelegatorTxs,
	numAddSubnetValidatorTxs,
	numAddValidatorTxs,
	numAdvanceTimeTxs,
	numCreateChainTxs,
	numCreateSubnetTxs,
	numExportTxs,
	numImportTxs,
	numRewardValidatorTxs prometheus.Counter

	apiRequestMetrics metric.APIInterceptor
}

func newBlockMetrics(namespace string, name string) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      fmt.Sprintf("%s_blks_accepted", name),
		Help:      fmt.Sprintf("Number of %s blocks accepted", name),
	})
}

func newTxMetrics(namespace string, name string) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      fmt.Sprintf("%s_txs_accepted", name),
		Help:      fmt.Sprintf("Number of %s transactions. accepted", name),
	})
}

// Initialize platformvm metrics
func (m *metrics) Initialize(
	namespace string,
	registerer prometheus.Registerer,
) error {
	m.percentConnected = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "percent_connected",
		Help:      "Percent of connected stake",
	})
	m.totalStake = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "total_staked",
		Help:      "Total amount of AVAX staked",
	})

	m.numAbortBlocks = newBlockMetrics(namespace, "abort")
	m.numAtomicBlocks = newBlockMetrics(namespace, "atomic")
	m.numCommitBlocks = newBlockMetrics(namespace, "commit")
	m.numProposalBlocks = newBlockMetrics(namespace, "proposal")
	m.numStandardBlocks = newBlockMetrics(namespace, "standard")

	m.numAddDelegatorTxs = newTxMetrics(namespace, "add_delegator")
	m.numAddSubnetValidatorTxs = newTxMetrics(namespace, "add_subnet_validator")
	m.numAddValidatorTxs = newTxMetrics(namespace, "add_validator")
	m.numAdvanceTimeTxs = newTxMetrics(namespace, "advance_time")
	m.numCreateChainTxs = newTxMetrics(namespace, "create_chain")
	m.numCreateSubnetTxs = newTxMetrics(namespace, "create_subnet")
	m.numExportTxs = newTxMetrics(namespace, "export")
	m.numImportTxs = newTxMetrics(namespace, "import")
	m.numRewardValidatorTxs = newTxMetrics(namespace, "reward_validator")

	apiRequestMetrics, err := metric.NewAPIInterceptor(namespace, registerer)
	m.apiRequestMetrics = apiRequestMetrics
	errs := wrappers.Errs{}
	errs.Add(
		err,

		registerer.Register(m.percentConnected),
		registerer.Register(m.totalStake),

		registerer.Register(m.numAbortBlocks),
		registerer.Register(m.numAtomicBlocks),
		registerer.Register(m.numCommitBlocks),
		registerer.Register(m.numProposalBlocks),
		registerer.Register(m.numStandardBlocks),

		registerer.Register(m.numAddDelegatorTxs),
		registerer.Register(m.numAddSubnetValidatorTxs),
		registerer.Register(m.numAddValidatorTxs),
		registerer.Register(m.numAdvanceTimeTxs),
		registerer.Register(m.numCreateChainTxs),
		registerer.Register(m.numCreateSubnetTxs),
		registerer.Register(m.numExportTxs),
		registerer.Register(m.numImportTxs),
		registerer.Register(m.numRewardValidatorTxs),
	)
	return errs.Err
}

func (m *metrics) AcceptBlock(b snowman.Block) error {
	switch b := b.(type) {
	case *AbortBlock:
		m.numAbortBlocks.Inc()
	case *AtomicBlock:
		m.numAtomicBlocks.Inc()
		return m.AcceptTx(&b.Tx)
	case *CommitBlock:
		m.numCommitBlocks.Inc()
	case *ProposalBlock:
		m.numProposalBlocks.Inc()
		return m.AcceptTx(&b.Tx)
	case *StandardBlock:
		m.numStandardBlocks.Inc()
		for _, tx := range b.Txs {
			if err := m.AcceptTx(tx); err != nil {
				return err
			}
		}
	default:
		return errUnknownBlockType
	}
	return nil
}

func (m *metrics) AcceptTx(tx *transactions.SignedTx) error {
	switch tx.UnsignedTx.(type) {
	case VerifiableUnsignedAddDelegatorTx:
		m.numAddDelegatorTxs.Inc()
	case VerifiableUnsignedAddSubnetValidatorTx:
		m.numAddSubnetValidatorTxs.Inc()
	case VerifiableUnsignedAddValidatorTx:
		m.numAddValidatorTxs.Inc()
	case VerifiableUnsignedAdvanceTimeTx:
		m.numAdvanceTimeTxs.Inc()
	case VerifiableUnsignedCreateChainTx:
		m.numCreateChainTxs.Inc()
	case VerifiableUnsignedCreateSubnetTx:
		m.numCreateSubnetTxs.Inc()
	case VerifiableUnsignedImportTx:
		m.numImportTxs.Inc()
	case VerifiableUnsignedExportTx:
		m.numExportTxs.Inc()
	case VerifiableUnsignedRewardValidatorTx:
		m.numRewardValidatorTxs.Inc()
	default:
		return errUnknownTxType
	}
	return nil
}
