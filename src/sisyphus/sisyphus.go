package sisyphus

import (
	"fmt"
	// "reflect"
	"strings"
	"syscall/js"
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
}

func (t *Table) Mount() {
	var (
		width           int
		height          int
		ctx             js.Value
		virtualScroller VirtualScroller
	)

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

func (t *Table) Render(ctx js.Value, offset int) {
	// Reset any forced renders
	t.triggerRender = false

	ctx.Call("clearRect", 0, 0, t.width, t.height)
	cellHeight := t.CellHeight
	rowsLen := len(t.rows)
	t.currentOffset = offset

	var api js.Value
	api = js.Global().Get("api")

	for i := 0; i < t.Total; i++ {
		y := i * cellHeight
		yConsideringHeaderOffset := y + cellHeight
		yConsideringScrollOffset := yConsideringHeaderOffset - offset
		rowVisible := yConsideringScrollOffset < t.height

		if !rowVisible {
			continue
		}

		if t.shouldFetchNextPage(i) {
			var nextData js.Value

			// @todo convert to async
			nextData = api.Call("onMoreData", i, i+25)

			for i := 0; i < nextData.Length(); i++ {
				v := nextData.Index(i)
				t.rows = append(t.rows, v)
			}

			t.triggerRender = true
		}

		if i < rowsLen {
			rowData := t.rows[i]
			row := Row{Width: t.width, Height: t.CellHeight, Columns: t.Columns, Data: rowData, Y: yConsideringScrollOffset}
			row.Render(ctx)
		} else {
			row := Row{Width: t.width, Height: t.CellHeight, Columns: t.Columns, Y: yConsideringScrollOffset}
			row.Render(ctx)
		}
	}

	// Render header *after* rows, so it is *above*
	headerRow := Row{Width: int(t.width), Height: t.CellHeight, Columns: t.Columns, IsHeader: true, Y: 0}
	headerRow.Render(ctx)
}

func (t *Table) shouldRender(nextOffset int) bool {
	return t.getCurrentOffset() != nextOffset || t.triggerRender
}

func (t *Table) getCurrentOffset() int {
	return t.currentOffset
}

func (t *Table) shouldFetchNextPage(currentIndex int) bool {
	rowsLen := len(t.rows)
	buffer := 25
	nextSet := currentIndex + buffer

	return rowsLen < nextSet && t.Total >= nextSet
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

type CellData struct {
	Value string
}

type Column struct {
	Property string
}

func (c *Column) GetData(row js.Value) string {
	return row.Get(c.Property).String()
}

type Row struct {
	Width    int
	Height   int
	Y        int
	Columns  []Column
	IsHeader bool
	Data     js.Value
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
			if (r.Data == js.Value{}) {
				ctx.Call("fillText", "*", x, r.Y+y)
			} else {
				ctx.Call("fillText", r.Columns[i].GetData(r.Data), x, r.Y+y)
			}
		}
	}

	ctx.Call("fillRect", 0, r.Y+r.Height-1, r.Width, 1)
}
