package sound

import (
	"sync"

	"github.com/ejv2/podbit/data"
)

// Singleton state.
var (
	mut sync.RWMutex

	head  int
	queue []*data.QueueItem
)

// Enqueue is the low-level enqueue routine.
// Enqueues a raw QueueItem for playback.
func Enqueue(item *data.QueueItem) {
	mut.Lock()
	defer mut.Unlock()

	queue = append(queue, item)
}

// EnqueueByURL searches data sources for episodes under <url>
// Remember to download before playing!
// If you know episode is downloaded, use EnqueueByTitle - it's faster.
func EnqueueByURL(url string) {
	item := data.Q.GetEpisodeByURL(url)
	if item != nil {
		Enqueue(item)
	}
}

// EnqueueByTitle searches data sources for episodes under the title <title>.
// Only the first match (if multiple are somehow present) is used.
//
// The availability of a title implies presence in cache, so don't bother to
// download the episode.
func EnqueueByTitle(title string) {
	item := data.Q.GetEpisodeByTitle(title)
	if item != nil {
		Enqueue(item)
	}
}

// EnqueueByPodcast bulk enqueues episodes by the name/url of their parent podcast <ident>.
// Does not check or care about download status.
//
// If <ident> is empty or does not exist, no action is taken.
//
// If a url is provided, a lookup is performed to find the friendly name.
// If a friendly name could not be found, the url is used instead.
// If a friendly name is provided, no lookup is performed.
func EnqueueByPodcast(ident string) {
	comp := data.DB.GetFriendlyName(ident)

	data.Q.Range(func(i int, elem *data.QueueItem) bool {
		name := data.DB.GetFriendlyName(elem.URL)
		if name == comp {
			Enqueue(data.Q.Items[i]) // Do not return: we are queueing in bulk
		}

		return true
	})
}

// JumpTo will force the head to the specified location.
// If the new location is invalid, no action is taken.
// After the jump, the player is instructed to play the
// new head.
func JumpTo(index int) {
	if index >= len(queue) {
		return
	}

	head = index

	Plr.Stop()
}

// PlayNow enqueues an episode and jumps to it, followed
// by a call to the player to play the new head.
func PlayNow(item *data.QueueItem) {
	Enqueue(item)
	JumpTo(len(queue) - 1)
}

// ClearQueue truncates the queue to zero items.
func ClearQueue() {
	queue = queue[:0]
	head = 0

	Plr.Stop()
}

// Dequeue removes an item from the queue at index.
// If index is invalid, no action is taken.
func Dequeue(index int) {
	mut.Lock()
	defer mut.Unlock()

	if index >= len(queue) {
		return
	}

	var after []*data.QueueItem
	prior := queue[:index]
	if index < len(queue)-1 {
		after = queue[index+1:]
	} else {
		after = make([]*data.QueueItem, 0)
	}

	queue = append(prior, after...)

	if index == head-1 {
		Plr.Stop()
		head--
	}

	if index < head-1 && head > 0 {
		head--
	}

	// Catch all - prevent crashes from spurious deletes
	if head < 0 {
		head = 0
	}
}

// GetQueue returns the raw queue in QueueItem slice form
// You should not edit the returned values, as this looses
// all thread protection.
func GetQueue() []*data.QueueItem {
	mut.RLock()
	defer mut.RUnlock()

	return queue
}

// PopHead returns the head of the queue, increasing
// the head beyond the end (if present). The boolean returned
// indicates if the player should stop - which occurs in the
// case of the end of the queue or an empty queue.
func PopHead() (*data.QueueItem, bool) {
	mut.Lock()
	defer mut.Unlock()

	if len(queue) > 0 {
		if head >= len(queue) {
			return nil, true
		}

		i := queue[head]
		head++

		return i, false
	}

	return nil, true
}

// GetHead returns the current queue item at the head of the
// queue (playing) and the index into the queue. This function
// does not pop any items (the head remains unchanged).
func GetHead() (h *data.QueueItem, pos int) {
	mut.RLock()
	defer mut.RUnlock()

	if len(queue) > 0 && len(queue) > head-1 && head-1 >= 0 {
		pos = head - 1
		h = queue[pos]
	}

	return
}

// GetNext returns the item which is enqueued to play next
// after the current entry has finished. If there is not an
// item enqueued next, returns nil and -1.
func GetNext() (*data.QueueItem, int) {
	if len(queue) > head {
		return queue[head], head
	}

	return nil, -1
}

// DownloadAtHead returns if the item next to play (before the head)
// is the expected ongoing download. Used by the sound system to detect
// download dequeues.
func DownloadAtHead(expect *data.QueueItem) bool {
	mut.RLock()
	defer mut.RUnlock()

	return len(queue) != 0 && head-1 >= 0 && queue[head-1] == expect
}
