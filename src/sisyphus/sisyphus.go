package sisyphus

import (
	"fmt"
	"reflect"
	"strings"
	"syscall/js"
)

type Table struct {
	width      float64
	height     float64
	CellHeight int
	Total      int
	Rows       []interface{}
	Columns    []Column
}

func (t *Table) Mount() {
	var (
		width           float64
		height          float64
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

	width = viewport.Get("clientWidth").Float()
	height = viewport.Get("clientHeight").Float()

	t.width = width
	t.height = height

	canvasEl.Set("width", width)
	canvasEl.Set("height", height)

	virtualScroller = VirtualScroller{Height: float64(t.CellHeight * t.Total), ScrollerEl: scroller, ScrollHeighterEl: scrollHeighter}

	var renderFrame js.Callback

	renderFrame = js.NewCallback(func(args []js.Value) {
		// Resize canvas if needed
		curBodyW := viewport.Get("clientWidth").Float()
		curBodyH := viewport.Get("clientHeight").Float()
		if curBodyW != width || curBodyH != height {
			width, height = curBodyW, curBodyH
			canvasEl.Set("width", width)
			canvasEl.Set("height", height)
		}

		t.Render(ctx, virtualScroller.GetOffset())

		js.Global().Call("requestAnimationFrame", renderFrame)
	})

	defer renderFrame.Release()

	virtualScroller.Mount()

	// Start loop
	js.Global().Call("requestAnimationFrame", renderFrame)

	<-done
}

func (t *Table) Render(ctx js.Value, offset float64) {
	ctx.Call("clearRect", 0, 0, t.width, t.height)
	cellHeight := t.CellHeight
	rowsLen := len(t.Rows)

	for i := 0; i < t.Total; i++ {
		y := i * cellHeight
		if i < rowsLen {
			rowData := t.Rows[i]
			row := Row{Width: int(t.width), Height: t.CellHeight, Columns: t.Columns, Data: rowData, Y: t.CellHeight + y - int(offset)}
			row.Render(ctx)
		} else {
			row := Row{Width: int(t.width), Height: t.CellHeight, Columns: t.Columns, Y: t.CellHeight + y - int(offset)}
			row.Render(ctx)
		}

		// x, y := 0, i*cellHeight+5
		// row := Row{Width: int(t.width), Height: t.CellHeight, Columns: t.Columns, Data: }
		// color := "red"

		// if i%2 == 0 {
		// 	color = "black"
		// }

		// ctx.Set("fillStyle", color)
		// ctx.Call("fillRect", x, float64(t.CellHeight)+float64(y)-offset, cellWidth, cellHeight)
	}

	// Render header *after* rows, so it is *above*
	headerRow := Row{Width: int(t.width), Height: t.CellHeight, Columns: t.Columns, IsHeader: true, Y: 0}
	headerRow.Render(ctx)
}

type VirtualScroller struct {
	Height           float64
	ScrollerEl       js.Value
	ScrollHeighterEl js.Value
}

func (v *VirtualScroller) Mount() {
	v.Update()
}

func (v *VirtualScroller) Update() {
	v.ScrollHeighterEl.Set("style", fmt.Sprintf("height: %gpx", v.Height))
}

func (v *VirtualScroller) GetOffset() float64 {
	return v.ScrollerEl.Get("scrollTop").Float()
}

type CellData struct {
	Value string
}

type Column struct {
	Property string
}

func (c *Column) GetData(row interface{}) string {
	return getField(row, c.Property)
}

func getField(row interface{}, field string) string {
	r := reflect.ValueOf(row)
	f := reflect.Indirect(r).FieldByName(field)
	return f.String()
}

type Row struct {
	Width    int
	Height   int
	Y        int
	Columns  []Column
	IsHeader bool
	Data     interface{}
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
		// x, y := i*widthPerColumn+(widthPerColumn/2), r.Height/2

		if i == 0 {
			ctx.Set("textAlign", "start")
		} else {
			ctx.Set("textAlign", "end")
		}

		if r.IsHeader {
			ctx.Call("fillText", strings.ToUpper(r.Columns[i].Property), x, r.Y+y)
		} else {
			if r.Data == nil {
				ctx.Call("fillText", "*", x, r.Y+y)
			} else {
				ctx.Call("fillText", r.Columns[i].GetData(r.Data), x, r.Y+y)
			}
		}

		// ctx.Call("strokeText", h.Columns[i].Property, x, y)
	}

	ctx.Call("fillRect", 0, r.Y+r.Height-1, r.Width, 1)
}
