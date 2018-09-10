package main

import (
	"sisyphus"
)

var columns = []sisyphus.Column{
	sisyphus.Column{Property: "name"},
	sisyphus.Column{Property: "email"},
	sisyphus.Column{Property: "count"},
}

func main() {
	table := sisyphus.Table{Total: 500, Columns: columns, CellHeight: 50, PageSize: 25}
	table.Mount()
}
