package sync

import (
	"context"
	"errors"
	"fmt"
	"github.com/NethermindEth/juno/db"
	"github.com/NethermindEth/juno/starknetdata"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/errgroup"
	big "math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NethermindEth/juno/blockchain"
	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/crypto"
	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/juno/core/trie"
	"github.com/NethermindEth/juno/service"
	"github.com/NethermindEth/juno/utils"
)

type MutableStorage interface {
	SetClasss(path *felt.Felt, classHash *felt.Felt, class core.Class) error
	SetAddress(paths []*felt.Felt, nodeHashes []*felt.Felt, classHashes []*felt.Felt, nonces []*felt.Felt) error
	SetStorage(storagePath *felt.Felt, paths []*felt.Felt, values []*felt.Felt) error
	GetStateRoot() (*felt.Felt, error)
	ApplyStateUpdate(update *core.StateUpdate, validate bool) error
}

type SnapSyncher struct {
	baseSync     service.Service
	starknetData starknetdata.StarknetData
	snapServer   *reliableSnapServer
	blockchain   *blockchain.Blockchain
	log          utils.Logger

	startingBlock    *core.Header
	lastBlock        *core.Header
	currentStateRoot *felt.Felt
	currentClassRoot *felt.Felt

	addressRangeDone chan interface{}
	storageRangeDone chan interface{}
	phase1Done       chan interface{}

	storageRangeJobCount int32
	storageRangeJob      chan *blockchain.StorageRangeRequest
	storageRangeJobRetry chan *blockchain.StorageRangeRequest

	largeStorageRangeJobCount int32
	largeStorageRangeJob      chan *blockchain.StorageRangeRequest
	largeStorageStoreJob      chan *largeStorageStoreJob

	classesJob chan *felt.Felt

	// Three lock priority lock
	mtxM *sync.Mutex
	mtxN *sync.Mutex
	mtxL *sync.Mutex
}

type largeStorageStoreJob struct {
	storagePath *felt.Felt
	changes     chan core.StorageDiff
}

func NewSnapSyncer(
	baseSyncher service.Service,
	consensus starknetdata.StarknetData,
	server blockchain.SnapServer,
	blockchain *blockchain.Blockchain,
	log utils.Logger,
) *SnapSyncher {
	return &SnapSyncher{
		baseSync:     baseSyncher,
		starknetData: consensus,
		snapServer: &reliableSnapServer{
			innerServer: server,
		},
		blockchain: blockchain,
		log:        log,
	}
}

var (
	addressDurations = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "juno_address_durations",
		Help:    "Time in address get",
		Buckets: prometheus.ExponentialBuckets(1.0, 1.7, 30),
	}, []string{"phase"})
	storageDurations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "juno_storage_durations",
		Help: "Time in address get",
	}, []string{"phase"})
	storageStoreSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "juno_storage_store_size",
		Help: "Time in address get",
	})
	storageStoreSizeTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "juno_storage_store_size_total",
		Help: "Time in address get",
	})

	largeStorageDurations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "juno_large_storage_durations",
		Help: "Time in address get",
	}, []string{"phase"})
	largeStorageStoreSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "juno_large_storage_store_size",
		Help: "Time in address get",
	})

	largeStorageStoreJobSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "juno_large_storage_store_job_size",
		Help: "Time in address get",
	})
	rangeProgress = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "juno_range_progress",
		Help: "Time in address get",
	})

	largeStorageStoreJobSizeTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "juno_large_storage_store_job_size_total",
		Help: "Time in address get",
	})
)

const (
	storageJobQueueSize             = 100000
	highPriorityStorageJobThreshold = 10000
	storageJobRetryQueueSize        = 10000000
	largeStorageJobQueueSize        = 0

	storageJobWorker      = 8
	largeStorageJobWorker = 64 // Large storage are largest and most parallelizable. So we want a lot of this.

	classRangeMaxNodes   = 10000
	addressRangeMaxNodes = 5000

	// Smaller value here seems to help quite a lot
	storageBatchSize = 400
	storageMaxNodes  = 2000

	largeStorageMaxNodes       = 5000
	largeStorageStoreQueueSize = 0

	classesJobQueueSize = 128

	maxPivotDistance = 2
)

func (s *SnapSyncher) initState(ctx context.Context) error {
	head, err := s.starknetData.BlockLatest(ctx)
	if err != nil {
		return errors.Join(err, errors.New("error getting current head"))
	}

	s.startingBlock = head.Header
	s.lastBlock = head.Header

	rootInfo, err := s.snapServer.GetTrieRootAt(ctx, s.startingBlock)
	if err != nil {
		return errors.Join(err, errors.New("error getting trie root"))
	}
	s.currentStateRoot = rootInfo.StorageRoot
	s.currentClassRoot = rootInfo.ClassRoot

	s.storageRangeJobCount = 0
	s.storageRangeJob = make(chan *blockchain.StorageRangeRequest, storageJobQueueSize)
	s.storageRangeJobRetry = make(chan *blockchain.StorageRangeRequest, storageJobRetryQueueSize)
	s.largeStorageRangeJobCount = 0
	s.largeStorageRangeJob = make(chan *blockchain.StorageRangeRequest, largeStorageJobQueueSize)
	s.largeStorageStoreJob = make(chan *largeStorageStoreJob, largeStorageStoreQueueSize)

	s.addressRangeDone = make(chan interface{})
	s.storageRangeDone = make(chan interface{})
	s.classesJob = make(chan *felt.Felt, classesJobQueueSize)
	s.phase1Done = make(chan interface{})

	s.mtxM = &sync.Mutex{}
	s.mtxN = &sync.Mutex{}
	s.mtxL = &sync.Mutex{}

	return nil
}

func (s *SnapSyncher) Run(ctx context.Context) error {
	err := s.runPhase1(ctx)
	if err != nil {
		return err
	}

	for i := s.startingBlock.Number; i <= s.lastBlock.Number; i++ {
		stateUpdate, err := s.starknetData.StateUpdate(ctx, uint64(i))
		if err != nil {
			return errors.Join(err, errors.New("error fetching state update"))
		}

		s.log.Infow("applying block", "blockNumber", i)

		err = s.ApplyStateUpdate(uint64(i), stateUpdate, false)
		if err != nil {
			return errors.Join(err, errors.New("error applying state update"))
		}
	}

	s.log.Infow("delegating to standard synchronizer")
	return s.baseSync.Run(ctx)
}

func (s *SnapSyncher) runPhase1(ctx context.Context) error {
	// 1. Get the current head
	// 2. Start the snap sync with pivot set to that head
	// 3. If at any moment, if:
	//    a. The current head is too new (more than 64 block let say)
	//    b. Too many missing node
	//    then reset the pivot.
	// 4. Once finished, replay state update from starting pivot to the latest pivot.
	// 5. Then do some cleanup, mark things and complete and such.
	// 6. Probably download old state updato/bodies too
	// 7. Send back control to base sync.

	starttime := time.Now()

	err := s.initState(ctx)
	if err != nil {
		return errors.Join(err, errors.New("error initializing snap syncer state"))
	}

	eg := &errgroup.Group{}

	eg.Go(func() error {
		defer func() {
			if err := recover(); err != nil {
				s.log.Errorw("latest block pool paniced", "err", err)
			}
		}()

		return s.poolLatestBlock(ctx)
	})

	eg.Go(func() error {
		defer func() {
			if err := recover(); err != nil {
				s.log.Errorw("class range paniced", "err", err)
			}
		}()

		err := s.runClassRangeWorker(ctx)
		if err != nil {
			s.log.Errorw("error in class range worker", "err", err)
		}

		return err
	})

	eg.Go(func() error {
		defer func() {
			if err := recover(); err != nil {
				s.log.Errorw("address range paniced", "err", err)
			}
		}()

		err := s.runAddressRangeWorker(ctx)
		if err != nil {
			s.log.Errorw("error in address range worker", "err", err)
		}

		return err
	})

	storageEg := &errgroup.Group{}
	for i := 0; i < storageJobWorker; i++ {
		i := i
		storageEg.Go(func() error {
			defer func() {
				if err := recover(); err != nil {
					s.log.Errorw("storage worker paniced", "err", err)
				}
			}()

			err := s.runStorageRangeWorker(ctx, i)
			if err != nil {
				s.log.Errorw("error in storage range worker", "err", err)
			}
			s.log.Infow("Storage worker completed", "workerId", i)

			return err
		})
	}

	// For notifying that storage range is done
	eg.Go(func() error {
		err := storageEg.Wait()
		if err != nil {
			return err
		}

		fmt.Printf("Storage range range completed")
		close(s.storageRangeDone)
		return nil
	})

	lStorageEg := &errgroup.Group{}
	for i := 0; i < largeStorageJobWorker; i++ {
		i := i
		lStorageEg.Go(func() error {
			defer func() {
				if err := recover(); err != nil {
					s.log.Errorw("large storage worker paniced", "err", err)
				}
			}()

			err := s.runLargeStorageRangeWorker(ctx, i)
			if err != nil {
				s.log.Errorw("error in large storage range worker", "err", err)
			}
			s.log.Infow("Large storage worker completed", "workerId", i)
			return err
		})
	}

	eg.Go(func() error {
		err := lStorageEg.Wait()
		close(s.largeStorageStoreJob)
		return err
	})

	eg.Go(func() error {
		err := s.runLargeStorageStore(ctx)
		if err != nil {
			s.log.Errorw("large storage store failed", "err", err)
		}
		return err
	})

	eg.Go(func() error {
		err := s.runFetchClassJob(ctx)
		if err != nil {
			s.log.Errorw("fetch class failed", "err", err)
		}
		return err
	})

	close(s.phase1Done)
	err = eg.Wait()
	if err != nil {
		return err
	}

	state, closer, err := s.blockchain.HeadState()
	if err != nil {
		return err
	}
	sroot, croot, err := state.(*core.State).StateAndClassRoot()
	if err != nil {
		return err
	}
	err = closer()
	if err != nil {
		return err
	}

	if !sroot.Equal(s.currentStateRoot) {
		return fmt.Errorf("state root mismatch %s vs %s", sroot.String(), s.currentStateRoot.String())
	}
	if !croot.Equal(s.currentClassRoot) {
		return fmt.Errorf("state root mismatch %s vs %s", sroot.String(), s.currentStateRoot.String())
	}

	s.log.Infow("first phase completed", "duration", time.Now().Sub(starttime).String())

	return nil
}

func calculatePercentage(f *felt.Felt) uint64 {
	maxint := big.NewInt(1)
	maxint.Lsh(maxint, 251)

	theint := f.BigInt(big.NewInt(0))
	theint.Mul(theint, big.NewInt(100))
	theint.Div(theint, maxint)

	return theint.Uint64()
}

func (s *SnapSyncher) runClassRangeWorker(ctx context.Context) error {
	totaladded := 0
	startAddr := &felt.Zero
	hasNext := true
	for hasNext {
		classRoot := s.currentClassRoot
		if classRoot == nil || classRoot.IsZero() {
			s.log.Infow("no class root", "progress", calculatePercentage(startAddr))
			return nil
		}

		s.log.Infow("class range progress", "progress", calculatePercentage(startAddr))

		newHasNext, response, err := s.snapServer.GetClassRange(ctx, classRoot, startAddr, nil, classRangeMaxNodes)
		if err != nil {
			return errors.Join(err, errors.New("error get address range"))
		}

		hasNext = newHasNext

		totaladded += len(response.Paths)
		err = s.SetClasss(response.Paths, response.ClassCommitments)
		if err != nil {
			return errors.Join(err, errors.New("error setting class"))
		}

		for _, path := range response.Paths {
			queued := false
			for !queued {
				select {
				case s.classesJob <- path:
					queued = true
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
					s.log.Infow("address queue stall on class")
				}
			}
		}

		startAddr = response.Paths[len(response.Paths)-1]
	}

	s.log.Infow("class range completed", "totalClass", totaladded)
	return nil
}

func (s *SnapSyncher) runAddressRangeWorker(ctx context.Context) error {
	defer func() {
		close(s.classesJob)
		close(s.addressRangeDone)
	}()

	startAddr := &felt.Zero
	hasNext := true
	for hasNext {
		curstateroot := s.currentStateRoot
		s.log.Infow("snap range progress", "progress", calculatePercentage(startAddr))
		rangeProgress.Set(float64(calculatePercentage(startAddr)))

		newHasNext, response, err := s.snapServer.GetAddressRange(ctx, curstateroot, startAddr, nil, addressRangeMaxNodes) // Verify is slow.
		if err != nil {
			return errors.Join(err, errors.New("error get address range"))
		}
		hasNext = newHasNext

		classHashes := make([]*felt.Felt, 0)
		nonces := make([]*felt.Felt, 0)
		for i := range response.Paths {
			classHashes = append(classHashes, response.Leaves[i].ClassHash)
			nonces = append(nonces, response.Leaves[i].Nonce)
		}

		// TODO: l0 class not in trie
		starttime := time.Now()
		err = s.SetAddress(response.Paths, response.Hashes, classHashes, nonces)
		addressDurations.WithLabelValues("set").Observe(float64(time.Now().Sub(starttime).Microseconds()))
		starttime = time.Now()
		if err != nil {
			return errors.Join(err, errors.New("error setting address"))
		}

		starttime = time.Now()
		for i, path := range response.Paths {
			if response.Leaves[i].ContractStorageRoot == nil {
				return errors.New("storage root is nil")
			}

			queued := false
			for !queued {
				select {
				case s.storageRangeJob <- &blockchain.StorageRangeRequest{
					Path:      path,
					Hash:      response.Leaves[i].ContractStorageRoot,
					StartAddr: &felt.Zero,
				}:
					queued = true
					atomic.AddInt32(&s.storageRangeJobCount, 1)
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
					s.log.Infow("address queue stall")
				}
			}

			queued = false
			for !queued {
				select {
				case s.classesJob <- response.Leaves[i].ClassHash:
					queued = true
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
					s.log.Infow("address queue stall on class")
				}
			}

		}

		startAddr = response.Paths[len(response.Paths)-1]
		addressDurations.WithLabelValues("queueing").Observe(float64(time.Now().Sub(starttime).Microseconds()))
	}

	s.log.Infow("address range completed")

	return nil
}

func (s *SnapSyncher) runStorageRangeWorker(ctx context.Context, workerIdx int) error {
	totalprocessed := 0
	for {
		requests := make([]*blockchain.StorageRangeRequest, 0)

	requestloop:
		for len(requests) < storageBatchSize {
			addressdonechecker := s.addressRangeDone
			if s.storageRangeJobCount > 0 {
				addressdonechecker = nil // So that it never complete
			}

			// Take from retry first, or there can be a deadlock
			// TODO: use a loop
			select {
			case job := <-s.storageRangeJobRetry:
				requests = append(requests, job)
				continue
			default:
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second * 10):
				if len(requests) > 0 {
					break requestloop
				}
			case <-addressdonechecker:
				// Its done...
				return nil
			case job := <-s.storageRangeJob:
				requests = append(requests, job)
			}
		}

		curstateroot := s.currentStateRoot

		starttime := time.Now()
		responses, err := s.snapServer.GetContractRange(curstateroot, requests, storageMaxNodes)
		storageDurations.WithLabelValues("get").Add(float64(time.Now().Sub(starttime).Microseconds()))
		if err != nil {
			return err
		}

		for i := len(responses); i < len(requests); i++ {
			unprocessedRequest := requests[i]
			select {
			case s.storageRangeJobRetry <- unprocessedRequest:
				break
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		totalSize := 0
		allDiffs := map[felt.Felt][]core.StorageDiff{}
		largeStorageRequests := make([]*blockchain.StorageRangeRequest, 0)

		for i, response := range responses {
			request := requests[i]

			// TODO: it could be nil if its updated and therefore require a refresh
			if len(response.Paths) == 0 {
				if !request.Hash.Equal(&felt.Zero) {
					return fmt.Errorf("empty path got non zero hash")
				}
				// TODO: need to check if its really empty
				atomic.AddInt32(&s.storageRangeJobCount, -1)
				totalprocessed++
				continue
			}

			starttime := time.Now()
			hasNext, err := trie.VerifyTrie(request.Hash, response.Paths, response.Values, response.Proofs, crypto.Pedersen)
			storageDurations.WithLabelValues("verify").Add(float64(time.Now().Sub(starttime).Microseconds()))
			if err != nil {
				fmt.Printf("Verification failed\n")
				fmt.Printf("Request %s %s\n", request.Hash.String(), request.Path.String())
				for i, path := range response.Paths {
					fmt.Printf("S %s -> %s\n", path.String(), response.Values[i].String())
				}

				return err
			}

			diffs := make([]core.StorageDiff, 0)
			for i, path := range response.Paths {
				diffs = append(diffs, core.StorageDiff{
					Key:   path,
					Value: response.Values[i],
				})
			}
			totalSize += len(diffs)

			allDiffs[*request.Path] = diffs
			if hasNext {
				largeStorageRequests = append(largeStorageRequests, request)
			}

			atomic.AddInt32(&s.storageRangeJobCount, -1)
			totalprocessed++
		}

		storageStoreSize.Set(float64(len(allDiffs)))

		starttime = time.Now()
		err = s.SetStorage(allDiffs, s.storageRangeJobCount > highPriorityStorageJobThreshold, false)
		storageDurations.WithLabelValues("set").Add(float64(time.Now().Sub(starttime).Microseconds()))
		starttime = time.Now()

		// Need to be after SetStorage or some contract would not be deployed yet.
		for _, request := range largeStorageRequests {
			queued := false
			for !queued {
				select {
				case s.largeStorageRangeJob <- request:
					queued = true
					atomic.AddInt32(&s.largeStorageRangeJobCount, 1)
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
					fmt.Printf("Storage queue stall\n")
				}
			}
		}
		storageDurations.WithLabelValues("queueing").Add(float64(time.Now().Sub(starttime).Microseconds()))

		if err != nil {
			return err
		}

	}
}

func (s *SnapSyncher) runLargeStorageRangeWorker(ctx context.Context, workerIdx int) error {
	for {
		storageRangeDone := s.storageRangeDone
		if s.largeStorageRangeJobCount > 0 {
			storageRangeDone = nil // So that it never complete
		}

		var job *blockchain.StorageRangeRequest

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-storageRangeDone:
			// Its done...
			return nil
		case job = <-s.largeStorageRangeJob:
		case <-time.After(time.Second * 5):
			continue
		}

		s.log.Infow("new large storage job", "path", job.Path.String(), "remainingJob", s.largeStorageRangeJobCount)

		err := s.fetchLargeStorageSlot(ctx, workerIdx, job)
		if err != nil {
			return err
		}

		atomic.AddInt32(&s.largeStorageRangeJobCount, -1)
	}
}

func (s *SnapSyncher) fetchLargeStorageSlot(ctx context.Context, workerIdx int, job *blockchain.StorageRangeRequest) error {
	outchan := make(chan core.StorageDiff)
	defer func() {
		close(outchan)
	}()

	select {
	case s.largeStorageStoreJob <- &largeStorageStoreJob{
		storagePath: job.Path,
		changes:     outchan,
	}:
	case <-ctx.Done():
	}

	startAddr := job.StartAddr
	hasNext := true
	for hasNext {
		s.log.Infow("large storage", "workerId", workerIdx, "path", job.Path, "percentage", calculatePercentage(startAddr))

		curstateroot := s.currentStateRoot
		job.StartAddr = startAddr
		starttime := time.Now()
		responses, err := s.snapServer.GetContractRange(curstateroot, []*blockchain.StorageRangeRequest{job}, largeStorageMaxNodes)
		largeStorageDurations.WithLabelValues("get").Add(float64(time.Now().Sub(starttime).Microseconds()))
		starttime = time.Now()
		if err != nil {
			return err
		}

		response := responses[0] // TODO: it can return nothing

		// TODO: Verify hashes
		hasNext, err = trie.VerifyTrie(job.Hash, response.Paths, response.Values, response.Proofs, crypto.Pedersen)
		largeStorageDurations.WithLabelValues("verify").Add(float64(time.Now().Sub(starttime).Microseconds()))
		starttime = time.Now()
		if err != nil {
			s.log.Warnw("trie verification failed in large store")
			return err
		}

		for i, path := range response.Paths {
			select {
			case outchan <- core.StorageDiff{
				Key:   path,
				Value: response.Values[i],
			}:
			case <-ctx.Done():
			}
		}

		starttime = time.Now()
		largeStorageDurations.WithLabelValues("queue").Add(float64(time.Now().Sub(starttime).Microseconds()))

		startAddr = response.Paths[len(response.Paths)-1]
	}

	return nil
}

func (s *SnapSyncher) runLargeStorageStore(ctx context.Context) error {

	rwlock := &sync.RWMutex{}
	curmap := map[felt.Felt][]core.StorageDiff{}

	type perPathJobs struct {
		storagePath *felt.Felt
		changes     []core.StorageDiff
	}
	centralfeeder := make(chan perPathJobs)

	feederThrottleThreshold := 30000
	perContractBatchSize := 10000
	maxTotalJobTrigger := 320000
	maxConcurrentContractTrigger := 32
	feederThrottleDelay := time.Millisecond

	eg := &errgroup.Group{}

	eg.Go(func() error {
		defer func() {
			close(centralfeeder)
		}()
		eg2 := &errgroup.Group{}

		for newJob := range s.largeStorageStoreJob {
			newJob := newJob

			eg2.Go(func() error {
				batch := make([]core.StorageDiff, 0)

				for job := range newJob.changes {
					batch = append(batch, job)

					if len(batch) > 1000 {
						for {
							rwlock.RLock()
							length := len(curmap[*newJob.storagePath])
							rwlock.RUnlock()

							if length > feederThrottleThreshold {
								select {
								case <-time.After(feederThrottleDelay):
									continue
								case <-s.storageRangeDone:
								}
							}

							centralfeeder <- perPathJobs{
								storagePath: newJob.storagePath,
								changes:     batch,
							}
							batch = make([]core.StorageDiff, 0)
							break
						}
					}
				}

				if len(batch) > 0 {
					centralfeeder <- perPathJobs{
						storagePath: newJob.storagePath,
						changes:     batch,
					}
				}

				return nil
			})
		}

		return eg2.Wait()
	})

	counter := 0
	for job := range centralfeeder {
		counter += len(job.changes)
		rwlock.Lock()
		curmap[*job.storagePath] = append(curmap[*job.storagePath], job.changes...)
		length := len(curmap)
		rwlock.Unlock()

		for counter > maxTotalJobTrigger || length > maxConcurrentContractTrigger {

			// In an effort to improve parallelism, we try to limit each contract to a batch size so that no
			// single contract take too much time, delaying other contract.
			rwlock.Lock()

			tostore := map[felt.Felt][]core.StorageDiff{}
			for k, diffs := range curmap {
				if len(diffs) > perContractBatchSize {
					tostore[k] = diffs[:perContractBatchSize]
				} else {
					tostore[k] = diffs
				}
			}

			jobsize := 0
			for k := range tostore {
				jobsize += len(tostore[k])
				if len(curmap[k]) > perContractBatchSize {
					curmap[k] = curmap[k][perContractBatchSize:]
				} else {
					delete(curmap, k)
				}
			}
			rwlock.Unlock()

			largeStorageStoreSize.Set(float64(len(tostore)))
			largeStorageStoreJobSize.Set(float64(jobsize))
			counter -= jobsize

			starttime := time.Now()
			err := s.SetStorage(tostore, false, true)
			largeStorageDurations.WithLabelValues("set").Add(float64(time.Now().Sub(starttime).Microseconds()))
			if err != nil {
				return errors.Join(err, errors.New("error storing large storage"))
			}
			s.log.Infow("large storage store", "size", len(curmap), "total", jobsize, "time", time.Now().Sub(starttime))
		}
	}

	s.log.Infow("large storage store job completed")

	err := eg.Wait()
	if err != nil {
		return nil
	}

	err = s.SetStorage(curmap, false, true)
	if err != nil {
		return errors.Join(err, errors.New("error storing large storage"))
	}

	return nil
}

func (s *SnapSyncher) poolLatestBlock(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second * 10):
			break
		case <-s.phase1Done:
			return nil
		}

		head, err := s.starknetData.BlockLatest(ctx)
		if err != nil {
			s.log.Infow("Error pooling latest block", "lastblock", s.lastBlock.Number, "err", err)
			return errors.Join(err, errors.New("error getting current head"))
		}

		// TODO: Race issue
		if head.Number-s.lastBlock.Number < maxPivotDistance {
			s.log.Infow("Not updating pivot yet", "lastblock", s.lastBlock.Number, "head", head.Number, "diff", head.Number-s.lastBlock.Number)
			continue
		}

		s.log.Infow("Switching snap pivot", "hash", head.Hash, "number", head.Number)
		s.lastBlock = head.Header

		rootInfo, err := s.snapServer.GetTrieRootAt(ctx, s.startingBlock)
		if err != nil {
			return errors.Join(err, errors.New("error getting trie root"))
		}
		s.currentStateRoot = rootInfo.StorageRoot
		s.currentClassRoot = rootInfo.ClassRoot
	}
}

func (s *SnapSyncher) ApplyStateUpdate(blockNumber uint64, update *core.StateUpdate, validate bool) error {
	ctx := context.Background()

	unknownClasses, err := s.fetchUnknownClasses(ctx, update)
	if err != nil {
		return err
	}

	block, err := s.starknetData.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return err
	}

	if validate {
		return s.blockchain.Store(block, update, unknownClasses)
	}

	return s.blockchain.ApplyNoVerify(block, update, unknownClasses)
}

func (s *SnapSyncher) fetchUnknownClasses(ctx context.Context, stateUpdate *core.StateUpdate) (map[felt.Felt]core.Class, error) {
	state, closer, err := s.blockchain.HeadState()
	if err != nil {
		// if err is db.ErrKeyNotFound we are on an empty DB
		if !errors.Is(err, db.ErrKeyNotFound) {
			return nil, err
		}
		closer = func() error {
			return nil
		}
	}

	newClasses := make(map[felt.Felt]core.Class)
	fetchIfNotFound := func(classHash *felt.Felt) error {
		if _, ok := newClasses[*classHash]; ok {
			return nil
		}

		stateErr := db.ErrKeyNotFound
		if state != nil {
			_, stateErr = state.Class(classHash)
		}

		if errors.Is(stateErr, db.ErrKeyNotFound) {
			class, fetchErr := s.starknetData.Class(ctx, classHash)
			if fetchErr == nil {
				newClasses[*classHash] = class
			}
			return fetchErr
		}
		return stateErr
	}

	for _, deployedContract := range stateUpdate.StateDiff.DeployedContracts {
		if err = fetchIfNotFound(deployedContract.ClassHash); err != nil {
			return nil, db.CloseAndWrapOnError(closer, err)
		}
	}
	for _, classHash := range stateUpdate.StateDiff.DeclaredV0Classes {
		if err = fetchIfNotFound(classHash); err != nil {
			return nil, db.CloseAndWrapOnError(closer, err)
		}
	}
	for _, declaredV1 := range stateUpdate.StateDiff.DeclaredV1Classes {
		if err = fetchIfNotFound(declaredV1.ClassHash); err != nil {
			return nil, db.CloseAndWrapOnError(closer, err)
		}
	}

	return newClasses, db.CloseAndWrapOnError(closer, nil)
}

func (s *SnapSyncher) GetStateRoot() (*felt.Felt, error) {
	state, close, err := s.blockchain.HeadState()
	if err == db.ErrKeyNotFound {
		return &felt.Zero, nil
	}
	if err != nil {
		return nil, err
	}

	trie, closer2, err := state.(core.StateReaderStorage).StorageTrie()
	if err != nil {
		return nil, err
	}

	root, err := trie.Root()
	if err != nil {
		return nil, err
	}

	closer2()
	close()

	return root, nil
}

func (s *SnapSyncher) SetClasss(paths []*felt.Felt, classCommitments []*felt.Felt) error {
	s.mtxN.Lock()
	s.mtxM.Lock()
	defer s.mtxM.Unlock()
	s.mtxN.Unlock()

	return s.blockchain.StoreClassCommitments(paths, classCommitments)
}

func (s *SnapSyncher) SetAddress(paths []*felt.Felt, nodeHashes []*felt.Felt, classHashes []*felt.Felt, nonces []*felt.Felt) error {
	s.mtxN.Lock()
	s.mtxM.Lock()
	defer s.mtxM.Unlock()
	s.mtxN.Unlock()

	starttime := time.Now()
	err := s.blockchain.StoreDirect(paths, classHashes, nodeHashes, nonces)
	addressDurations.WithLabelValues("effective_set").Observe(float64(time.Now().Sub(starttime).Microseconds()))
	return err
}

func (s *SnapSyncher) SetStorage(diffs map[felt.Felt][]core.StorageDiff, higherPriority bool, isLargeStore bool) error {
	// if !higherPriority {
	s.mtxL.Lock()
	defer s.mtxL.Unlock()
	// }

	s.mtxN.Lock()
	s.mtxM.Lock()
	defer s.mtxM.Unlock()
	s.mtxN.Unlock()

	starttime := time.Now()
	err := s.blockchain.StoreStorageDirect(diffs)

	if isLargeStore {
		jobsize := 0
		for _, storageDiffs := range diffs {
			jobsize += len(storageDiffs)
		}
		largeStorageDurations.WithLabelValues("effective_set").Add(float64(time.Now().Sub(starttime).Microseconds()))
		largeStorageStoreJobSizeTotal.Add(float64(jobsize))
	} else {
		jobsize := 0
		for _, storageDiffs := range diffs {
			jobsize += len(storageDiffs)
		}
		storageDurations.WithLabelValues("effective_set").Add(float64(time.Now().Sub(starttime).Microseconds()))
		storageStoreSizeTotal.Add(float64(jobsize))
	}

	return err

}

func (s *SnapSyncher) runFetchClassJob(ctx context.Context) error {

	keyBatches := make([]*felt.Felt, 0)
	for key := range s.classesJob {
		if key == nil || key.IsZero() {
			// Not sure why...
			continue
		}

		cls, err := s.blockchain.GetClasses([]*felt.Felt{key})
		if err != nil {
			s.log.Infow("error getting class", "err", err)
			return err
		}

		if cls[0] == nil {
			keyBatches = append(keyBatches, key)
		}

		if len(keyBatches) > 1000 {
			classes, err := s.snapServer.GetClasses(ctx, keyBatches)
			if err != nil {
				s.log.Infow("error getting class from outside", "err", err)
				return err
			}

			newBatch := make([]*felt.Felt, 0)
			newClassKeys := make([]*felt.Felt, 0)
			newClasses := make([]core.Class, 0)
			for i, class := range classes {
				if class == nil {
					s.log.Warnw("class %s not found", keyBatches[i])
					newBatch = append(newBatch, keyBatches[i])
					continue
				}

				newClassKeys = append(newClassKeys, keyBatches[i])
				newClasses = append(newClasses, class)
			}

			err = s.blockchain.StoreClasses(newClassKeys, newClasses)
			if err != nil {
				s.log.Infow("error storing class", "err", err)
				return err
			}

			keyBatches = newBatch
		}
	}

	return nil
}

var _ service.Service = (*SnapSyncher)(nil)
