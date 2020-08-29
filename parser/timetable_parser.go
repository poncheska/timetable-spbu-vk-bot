package parser

import (
	"bytes"
	"fmt"
	"github.com/gocolly/colly"
	"regexp"
)

var regexNotSpace = regexp.MustCompile("(\\S+\\s?)+[^\\s]")

type Lesson struct {
	Time    string
	Type    string
	Place   string
	Teacher string
}

type Day struct {
	Date    string
	Lessons []Lesson
}

type Timetable struct {
	Days []Day
}

type ParseError struct {
	s string
}

func New(text string) error {
	return &ParseError{text}
}

func (e *ParseError) Error() string {
	return e.s
}

func ParseTimetable(link string) (*Timetable, error) {
	c := colly.NewCollector()

	tt := &Timetable{make([]Day, 0, 0)}

	c.OnHTML("div.panel-group div.panel-default", func(e *colly.HTMLElement) {
		times := regexNotSpace.FindAllString(e.DOM.Find("div.panel-default > ul.panel-collapse > li.row > "+
			"div:nth-child(1) div:nth-child(2)").Text(), -1)
		types := regexNotSpace.FindAllString(e.DOM.Find("div.panel-default > ul.panel-collapse > li.row > "+
			"div:nth-child(2) div:nth-child(2)").Text(), -1)
		places := regexNotSpace.FindAllString(e.DOM.Find("div.panel-default > ul.panel-collapse > li.row > "+
			"div:nth-child(3) div:nth-child(2)").Text(), -1)
		teachers := regexNotSpace.FindAllString(e.DOM.Find("div.panel-default > ul.panel-collapse > li.row > "+
			"div:nth-child(4) div:nth-child(2)").Text(), -1)
		date := regexNotSpace.FindString(e.DOM.Find("div.panel-default > div.panel-heading").Text())
		d := Day{date, make([]Lesson, 0, 0)}
		for i, _ := range times {
			l := Lesson{times[i], types[i], places[i], teachers[i]}
			d.Lessons = append(d.Lessons, l)
		}
		tt.Days = append(tt.Days, d)
	})

	err := c.Visit(link)

	return tt, err
}

func (tt *Timetable) GetString() string{
	buf := bytes.Buffer{}
	buf.WriteString("Расписание на неделю:\n")
	for _, day := range tt.Days{
		buf.WriteString(day.Date + "\n")
		for j, les := range day.Lessons{
			buf.WriteString(fmt.Sprintf("  %v.%v Время:%v\n  Место:%v Препод.:%v\n",
				j+1, les.Type, les.Time, les.Place, les.Teacher))
		}
	}
	return buf.String()
}