package sisyphus

import (
	"fmt"
	// "reflect"
	"github.com/patrickmn/go-cache"
	"math"
	"strings"
	"syscall/js"
	"time"
)

type Table struct {
	width         int
	height        int
	CellHeight    int
	Total         int
	rows          []js.Value
	Columns       []Column
	currentOffset int
	hasMounted    bool
	triggerRender bool
	store         *cache.Cache
	pages         []Page
	PageSize      int
	api           js.Value
}

func (t *Table) Mount() {
	var (
		width           int
		height          int
		ctx             js.Value
		virtualScroller VirtualScroller
	)
	t.api = js.Global().Get("api")

	t.fillPages()
	t.store = cache.New(60*time.Minute, 10*time.Minute)
	t.hydratePage(0)

	done := make(chan struct{}, 0)
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "surface")
	viewport := doc.Call("getElementById", "viewport")
	scroller := doc.Call("getElementById", "scroller")
	scrollHeighter := doc.Call("getElementById", "scrollHeighter")
	ctx = canvasEl.Call("getContext", "2d")

	width = viewport.Get("clientWidth").Int()
	height = viewport.Get("clientHeight").Int()

	t.width = width
	t.height = height

	canvasEl.Set("width", width)
	canvasEl.Set("height", height)

	virtualScroller = VirtualScroller{Height: t.CellHeight * t.Total, ScrollerEl: scroller, ScrollHeighterEl: scrollHeighter}

	var renderFrame js.Callback

	renderFrame = js.NewCallback(func(args []js.Value) {
		nextScrollOffset := virtualScroller.GetOffset()

		if !t.shouldRender(nextScrollOffset) {
			js.Global().Call("requestAnimationFrame", renderFrame)
			return
		}

		// Resize canvas if needed
		curBodyW := viewport.Get("clientWidth").Int()
		curBodyH := viewport.Get("clientHeight").Int()
		if curBodyW != width || curBodyH != height {
			width, height = curBodyW, curBodyH
			canvasEl.Set("width", width)
			canvasEl.Set("height", height)
		}

		t.Render(ctx, nextScrollOffset)

		js.Global().Call("requestAnimationFrame", renderFrame)
	})

	defer renderFrame.Release()

	virtualScroller.Mount()
	// Initial render
	t.Render(ctx, 0)

	// Start loop
	js.Global().Call("requestAnimationFrame", renderFrame)

	<-done
}

func (t *Table) fillPages() {
	pages := float64(t.Total) / float64(t.PageSize)

	f := RoundUpToWholeNumber(pages)

	for i := 0; i < f; i++ {
		t.pages = append(t.pages, Page{Offset: i * t.PageSize, Size: t.PageSize})
	}
}

func RoundUpToWholeNumber(x float64) int {
	t := math.Trunc(x)

	if math.Abs(x-t) > 0 {
		return int(t + math.Copysign(1, x))
	}
	return int(t)
}

func (t *Table) getStoredRowCount() int {
	var count int
	rx, found := t.store.Get("count")

	if found {
		count = rx.(int)
	}

	return count
}

func (t *Table) Render(ctx js.Value, offset int) {
	// Reset any forced renders
	t.triggerRender = false

	ctx.Call("clearRect", 0, 0, t.width, t.height)
	cellHeight := t.CellHeight
	t.currentOffset = offset

	// count := t.getStoredRowCount()

	for i := 0; i < t.Total; i++ {
		y := i * cellHeight
		yConsideringHeaderOffset := y + cellHeight
		yConsideringScrollOffset := yConsideringHeaderOffset - offset
		rowVisible := yConsideringScrollOffset < t.height

		if !rowVisible {
			continue
		}

		// doing a lot of work here... probably a bad idea
		if t.shouldFetchNextPage(i) {
			currentPageIndex := i / t.PageSize

			// @todo convert to async
			t.hydratePage(currentPageIndex + 1)

			t.triggerRender = true
		}

		var rowData *js.Value
		rxd, foundRow := t.store.Get(fmt.Sprintf("record-%d", i))

		if foundRow {
			rowData = rxd.(*js.Value)
		}

		row := Row{Width: t.width, Height: t.CellHeight, Columns: t.Columns, Data: rowData, Y: yConsideringScrollOffset}
		row.Render(ctx)
	}

	// Render header *after* rows, so it is *above*
	headerRow := Row{Width: int(t.width), Height: t.CellHeight, Columns: t.Columns, IsHeader: true, Y: 0}
	headerRow.Render(ctx)
}

func (t *Table) hydratePage(pageIndex int) {
	page := &t.pages[pageIndex]

	var nextData js.Value
	nextData = t.api.Call("onMoreData", page.Offset, page.Size)
	nextDataLength := nextData.Length()

	for idx := 0; idx < nextDataLength; idx++ {
		v := nextData.Index(idx)
		t.store.Set(fmt.Sprintf("record-%d", page.Offset+idx), &v, cache.NoExpiration)
	}

	page.Fulfilled = true
}

func (t *Table) shouldRender(nextOffset int) bool {
	return t.getCurrentOffset() != nextOffset || t.triggerRender
}

func (t *Table) getCurrentOffset() int {
	return t.currentOffset
}

func (t *Table) shouldFetchNextPage(currentIndex int) bool {
	currentPageIndex := currentIndex / t.PageSize
	percentOverPage := float32(currentIndex)/float32(t.PageSize) - float32(currentPageIndex)
	overthresholdForNextPage := percentOverPage >= 0.8
	currentPageFulfilled := t.pages[currentPageIndex].Fulfilled

	var nextPageFulfilled bool

	if currentPageIndex+1 > len(t.pages)-1 {
		nextPageFulfilled = true
	} else {
		nextPageFulfilled = t.pages[currentPageIndex+1].Fulfilled
	}

	return !currentPageFulfilled || overthresholdForNextPage && !nextPageFulfilled
}

type VirtualScroller struct {
	Height           int
	ScrollerEl       js.Value
	ScrollHeighterEl js.Value
}

func (v *VirtualScroller) Mount() {
	v.Update()
}

func (v *VirtualScroller) Update() {
	v.ScrollHeighterEl.Set("style", fmt.Sprintf("height: %dpx", v.Height))
}

func (v *VirtualScroller) GetOffset() int {
	return v.ScrollerEl.Get("scrollTop").Int()
}

type Page struct {
	Offset    int
	Size      int
	Fulfilled bool
}

type Column struct {
	Property string
}

func (c *Column) GetData(row *js.Value) string {
	if row == nil {
		return "*"
	}

	return row.Get(c.Property).String()
}

type Row struct {
	Width    int
	Height   int
	Y        int
	Columns  []Column
	IsHeader bool
	Data     *js.Value
}

func (r *Row) Render(ctx js.Value) {
	widthPerColumn := r.Width / len(r.Columns)

	ctx.Set("fillStyle", "#FFF")
	ctx.Call("fillRect", 0, r.Y, r.Width, r.Y+r.Height)
	ctx.Set("fillStyle", "#111")
	ctx.Set("strokeStyle", "#888")
	ctx.Set("font", "16px sans-serif")
	for i := 0; i < len(r.Columns); i++ {
		x, y := i*widthPerColumn+(widthPerColumn/3), r.Height/2

		if i == 0 {
			ctx.Set("textAlign", "start")
		} else {
			ctx.Set("textAlign", "end")
		}

		if r.IsHeader {
			ctx.Call("fillText", strings.ToUpper(r.Columns[i].Property), x, r.Y+y)
		} else {
			ctx.Call("fillText", r.Columns[i].GetData(r.Data), x, r.Y+y)
		}
	}

	ctx.Call("fillRect", 0, r.Y+r.Height-1, r.Width, 1)
}
