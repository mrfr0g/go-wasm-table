package main

import (
	"sisyphus"
)

type Row struct {
	name  string
	date  string
	count string
}

var rows = []Row{
	Row{
		name:  "Bob Johnson",
		date:  "01/20/2020",
		count: "50",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
	Row{
		name:  "Billy Jo Johnson",
		date:  "01/20/2020",
		count: "150",
	},
}

var columns = []sisyphus.Column{
	sisyphus.Column{Property: "name"},
	sisyphus.Column{Property: "date"},
	sisyphus.Column{Property: "count"},
}

func main() {
	s := make([]interface{}, len(rows))
	for i, v := range rows {
		s[i] = v
	}

	table := sisyphus.Table{Total: 100, Columns: columns, CellHeight: 50}
	table.Mount()
}
