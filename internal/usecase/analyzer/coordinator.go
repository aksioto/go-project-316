package analyzer

import (
	"context"
	"sync"

	"code/internal/domain"
	"code/internal/usecase/analyzer/filter"
	"code/internal/usecase/analyzer/normalizer"
)

type pageTask struct {
	url   string
	depth int
}

type pageResult struct {
	page          domain.Page
	links         []domain.Link
	depth         int
	shouldInclude bool
}

type pageProcessor interface {
	ProcessPage(ctx context.Context, url string, depth int) pageResult
}

type Coordinator struct {
	processor    pageProcessor
	domainFilter filter.DomainFilter
	rateLimiter  Waiter
	maxDepth     int
	workers      int

	mu      sync.Mutex
	visited map[string]struct{}
	pending int32
	pages   []domain.Page
}

func NewCoordinator(
	processor pageProcessor,
	domainFilter filter.DomainFilter,
	rateLimiter Waiter,
	maxDepth int,
	workers int,
) *Coordinator {
	if workers < 1 {
		workers = 1
	}
	return &Coordinator{
		processor:    processor,
		domainFilter: domainFilter,
		rateLimiter:  rateLimiter,
		maxDepth:     maxDepth,
		workers:      workers,
		visited:      make(map[string]struct{}),
	}
}

func (c *Coordinator) Crawl(ctx context.Context, startURL string) []domain.Page {
	normalized := normalizer.NormalizeURL(startURL)
	c.visited[normalized] = struct{}{}

	if err := c.rateLimiter.Wait(ctx); err != nil {
		return c.pages
	}

	result := c.processor.ProcessPage(ctx, startURL, 0)
	if result.shouldInclude {
		c.pages = append(c.pages, result.page)
	}

	if ctx.Err() != nil || c.maxDepth <= 1 {
		return c.pages
	}

	queue := c.filterLinks(result.links, 1)
	if len(queue) == 0 {
		return c.pages
	}

	c.pending = int32(len(queue))
	c.runWorkers(ctx, queue)

	return c.pages
}

func (c *Coordinator) tryVisit(linkURL string) (string, bool) {
	normalized := normalizer.NormalizeURL(linkURL)
	if !c.domainFilter.IsSameDomain(normalized) {
		return "", false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, seen := c.visited[normalized]; seen {
		return "", false
	}
	c.visited[normalized] = struct{}{}
	return normalized, true
}

func (c *Coordinator) filterLinks(links []domain.Link, depth int) []pageTask {
	tasks := make([]pageTask, 0, len(links))
	for _, link := range links {
		normalized, ok := c.tryVisit(link.URL)
		if !ok {
			continue
		}
		tasks = append(tasks, pageTask{url: normalized, depth: depth})
	}
	return tasks
}

func (c *Coordinator) runWorkers(ctx context.Context, initialQueue []pageTask) {
	tasks := make(chan pageTask, c.workers*100)
	results := make(chan pageResult, c.workers*100)
	done := make(chan struct{})

	var closeTasks sync.Once
	var closeDone sync.Once
	closeTasksFn := func() { closeTasks.Do(func() { close(tasks) }) }
	closeDoneFn := func() { closeDone.Do(func() { close(done) }) }

	var wg sync.WaitGroup
	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go c.worker(ctx, &wg, tasks, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go c.sendInitialTasks(ctx, tasks, closeTasksFn, closeDoneFn, initialQueue)
	go c.collectResults(ctx, tasks, results, closeTasksFn, closeDoneFn)

	<-done
}

func (c *Coordinator) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	tasks <-chan pageTask,
	results chan<- pageResult,
) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tasks:
			if !ok {
				return
			}
			if ctx.Err() != nil {
				c.decrementPending()
				continue
			}
			if err := c.rateLimiter.Wait(ctx); err != nil {
				c.decrementPending()
				continue
			}
			result := c.processor.ProcessPage(ctx, task.url, task.depth)
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *Coordinator) sendInitialTasks(
	ctx context.Context,
	tasks chan<- pageTask,
	closeTasksFn func(),
	closeDoneFn func(),
	queue []pageTask,
) {
	for _, item := range queue {
		select {
		case tasks <- item:
		case <-ctx.Done():
			closeTasksFn()
			closeDoneFn()
			return
		}
	}
}

func (c *Coordinator) collectResults(
	ctx context.Context,
	tasks chan pageTask,
	results <-chan pageResult,
	closeTasksFn func(),
	closeDoneFn func(),
) {
	for {
		select {
		case <-ctx.Done():
			closeTasksFn()
			closeDoneFn()
			return
		case res, ok := <-results:
			if !ok {
				closeDoneFn()
				return
			}

			if res.shouldInclude {
				c.mu.Lock()
				c.pages = append(c.pages, res.page)
				c.mu.Unlock()
			}

			c.enqueueNewLinks(ctx, tasks, res)

			c.mu.Lock()
			c.pending--
			if c.pending == 0 {
				c.mu.Unlock()
				closeTasksFn()
				closeDoneFn()
				return
			}
			c.mu.Unlock()
		}
	}
}

func (c *Coordinator) enqueueNewLinks(
	ctx context.Context,
	tasks chan<- pageTask,
	res pageResult,
) {
	if res.depth+1 >= c.maxDepth {
		return
	}

	for _, link := range res.links {
		normalized, ok := c.tryVisit(link.URL)
		if !ok {
			continue
		}

		c.mu.Lock()
		c.pending++
		c.mu.Unlock()

		select {
		case tasks <- pageTask{url: normalized, depth: res.depth + 1}:
		case <-ctx.Done():
			return
		}
	}
}

func (c *Coordinator) decrementPending() {
	c.mu.Lock()
	c.pending--
	c.mu.Unlock()
}
