package main

import (
	"strconv"
)

type Usage struct {
	Users *UsageStat
	Bookmarks *UsageStat
}

type UsageStat struct {
	Values []int
	Titles []string
	Title string
	Max int
}

type Bargraph struct {
	Data []DataPoint
	Title string
}

type DataPoint struct {
	Value int
	Title string
	Width string
}

func (u *UsageStat) AsBarGraph() (b Bargraph) {

	b.Title = u.Title
	var d DataPoint
	for i, value := range u.Values {
		ratio := float64(value) / float64(u.Max)

		d.Value = value
		d.Width = strconv.Itoa(int(ratio*100)) + "%"
		d.Title = u.Titles[i]
		b.Data = append(b.Data, d)
	}
	return b
}
