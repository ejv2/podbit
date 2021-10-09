package sound

import (
	"github.com/ethanv2/podbit/data"
)

// Singleton state
var queue []data.QueueItem

// Low-level enqueue routine
// Enqueues a raw QueueItem for playback
func Enqueue(item data.QueueItem) {
	queue = append(queue, item)
}

// Searches data sources for episodes under <url>
// Remember to download before playing!
// If you know episode is downloaded, use EnqueueByTitle - it's faster
func EnqueueByUrl(url string) {
	for _, elem := range data.Q.Items {
		if url == elem.Url {
			Enqueue(elem)
			return
		}
	}
}

// Searches data sources for episodes under the title <title>
// Only the first match (if multiple are somehow present) is used
//
// The availability of a title implies presence in cache, so don't bother to download the episode
func EnqueueByTitle(title string) {
	for _, elem := range data.Q.Items {
		ent, ok := data.Caching.Query(elem.Path)
		if ok && ent.Title == title {
			Enqueue(elem)
			return
		}
	}
}

// Bulk enqueue episodes by the name/url of their parent podcast <ident>
// Does not check or care about download status
//
// If <ident> is empty or does not exist, no action is taken
//
// If a url is provided, a lookup is performed to find the friendly name
// If a friendly name could not be found, the url is used instead
// If a friendly name is provided, no lookup is performed
func EnqueueByPodcast(ident string) {
	comp := data.DB.GetFriendlyName(ident)

	for _, elem := range data.Q.Items {
		name := data.DB.GetFriendlyName(elem.Url)
		if name == comp {
			Enqueue(elem) // Do not return: we are queueing in bulk
		}
	}
}

// Truncates the queue to zero items
func ClearQueue() {
	queue = queue[:0]
}

// Return the raw queue in QueueItem slice form
func GetQueue() []data.QueueItem {
	return queue
}
