package sound

import (
	"github.com/ethanv2/podbit/data"
)

// Singleton state
var head int
var queue []*data.QueueItem

// Enqueue is the low-level enqueue routine
// Enqueues a raw QueueItem for playback
func Enqueue(item *data.QueueItem) {
	queue = append(queue, item)
}

// EnqueueByURL searches data sources for episodes under <url>
// Remember to download before playing!
// If you know episode is downloaded, use EnqueueByTitle - it's faster
func EnqueueByURL(url string) {
	item := data.Q.GetEpisodeByURL(url)
	if item != nil {
		Enqueue(item)
	}
}

// EnqueueByTitle searches data sources for episodes under the title <title>
// Only the first match (if multiple are somehow present) is used
//
// The availability of a title implies presence in cache, so don't bother to download the episode
func EnqueueByTitle(title string) {
	item := data.Q.GetEpisodeByTitle(title)
	if item != nil {
		Enqueue(item)
	}
}

// EnqueueByPodcast bulk enqueues episodes by the name/url of their parent podcast <ident>
// Does not check or care about download status
//
// If <ident> is empty or does not exist, no action is taken
//
// If a url is provided, a lookup is performed to find the friendly name
// If a friendly name could not be found, the url is used instead
// If a friendly name is provided, no lookup is performed
func EnqueueByPodcast(ident string) {
	comp := data.DB.GetFriendlyName(ident)

	data.Q.Range(func (i int, elem *data.QueueItem) bool {
		name := data.DB.GetFriendlyName(elem.URL)
		if name == comp {
			Enqueue(&data.Q.Items[i]) // Do not return: we are queueing in bulk
		}

		return true
	})
}

// ClearQueue truncates the queue to zero items
func ClearQueue() {
	queue = queue[:0]
}

// GetQueue returns the raw queue in QueueItem slice form
// You should not edit the returned values, as this looses
// all thread protection
func GetQueue() []*data.QueueItem {
	return queue
}

// GetQueueHead returns the head of the queue, increasing
// the head beyond the end (if present). The boolean returned
// indicates if the player should stop - which occurs in the
// case of the end of the queue or an empty queue
func GetQueueHead() (*data.QueueItem, bool) {
	if len(queue) > 0 {
		if head >= len(queue) {
			head = 0
			return nil, true
		}

		i := queue[head]
		head++

		return i, false
	}

	return nil, true
}
