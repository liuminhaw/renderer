package renderer

import (
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
)

const longPollTimeout = 5 * time.Second

type requestInfo struct {
	start        time.Time
	resourceType network.ResourceType
	url          string
	// If ignore is true, the request will not be counted
	ignore bool
}

type interactiveTime struct {
	done chan struct{}
}

func newInteractiveTime() *interactiveTime {
	return &interactiveTime{
		done: make(chan struct{}, 1),
	}
}

// networkIdle is a struct to keep track of requests state
type networkIdle struct {
	mu sync.Mutex
	// request records to check if network is idle
	active map[network.RequestID]*requestInfo
	// record document requests by loader ID
	byLoader map[cdp.LoaderID]network.RequestID
	// network is considered idle when there are less than or equal to maxInflight
	// requests in action in the last `idle` time duration window.
	idleTimer   *time.Timer
	idle        time.Duration
	maxInflight int
	done        chan struct{}
	stopped     bool
}

func newNetworkIdle(idle time.Duration, maxInflight int) *networkIdle {
	return &networkIdle{
		mu:          sync.Mutex{},
		active:      make(map[network.RequestID]*requestInfo),
		byLoader:    make(map[cdp.LoaderID]network.RequestID),
		done:        make(chan struct{}, 1),
		idle:        idle,
		maxInflight: maxInflight,
	}
}

func (n *networkIdle) countActive() int {
	count := 0
	for _, info := range n.active {
		if !info.ignore {
			count++
		}
	}
	return count
}

func (n *networkIdle) startOrResetTimer() {
	if n.idleTimer != nil {
		if n.idleTimer.Stop() {
			// stopped before firing; okay
		}
	}
	n.idleTimer = time.AfterFunc(n.idle, func() {
		if n.stopped {
			return
		}
		if n.countActive() <= n.maxInflight {
			select {
			case n.done <- struct{}{}:
			default:
			}
		}
	})
}

func (n *networkIdle) maybeArm() {
	if n.countActive() <= n.maxInflight {
		n.startOrResetTimer()
	} else if n.idleTimer != nil {
		_ = n.idleTimer.Stop() // let it be re-armed when things quiet down
	}
}

func (n *networkIdle) promoteLongPoll() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.stopped {
		return
	}
	now := time.Now()
	for _, info := range n.active {
		if (info.resourceType == network.ResourceTypeXHR ||
			info.resourceType == network.ResourceTypeFetch) &&
			!info.ignore &&
			now.Sub(info.start) > longPollTimeout {
			info.ignore = true // ignore this request
		}
	}
	n.maybeArm()
}

func (n *networkIdle) add(id network.RequestID, resourceType network.ResourceType, url string) int {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.active[id] = &requestInfo{
		start:        time.Now(),
		resourceType: resourceType,
		url:          url,
		ignore:       false,
	}
	n.maybeArm()

	return len(n.active)
}

func (n *networkIdle) remove(id network.RequestID) int {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.active, id)
	n.maybeArm()

	return len(n.active)
}

func (n *networkIdle) addByLoader(
	loaderID cdp.LoaderID,
	id network.RequestID,
	resourceType network.ResourceType,
	url string,
) int {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.active[id] = &requestInfo{
		start:        time.Now(),
		resourceType: resourceType,
		url:          url,
		ignore:       false,
	}
	n.byLoader[loaderID] = id
	n.maybeArm()

	return len(n.active)
}

// isNoisyRequest checks if the request is considered noisy and should be ignored
// based on its resource type.
func isNoisyRequest(e *network.EventRequestWillBeSent) bool {
	switch e.Type {
	case network.ResourceTypeEventSource,
		network.ResourceTypeWebSocket,
		network.ResourceTypeMedia,
		network.ResourceTypeTextTrack,
		network.ResourceTypePing,
		network.ResourceTypeManifest:
		return true
	}

	return false
}
