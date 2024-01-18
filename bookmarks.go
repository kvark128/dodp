package dodp

import (
	"encoding/xml"
)

type Bookmark struct {
	XMLName    xml.Name `xml:"bookmark"`
	NcxRef     string   `xml:"ncxRef"`
	URI        string   `xml:"URI"`
	TimeOffset string   `xml:"timeOffset"`
	CharOffset string   `xml:"charOffset"`
	Note       Note
	Label      string `xml:"label,attr"`
	Lang       string `xml:"lang,attr"`
}

type BookmarkSet struct {
	XMLName  xml.Name `xml:"http://www.daisy.org/z3986/2005/bookmark/ bookmarkSet"`
	Title    Title
	UID      string `xml:"uid"`
	Lastmark Lastmark
	Bookmark []Bookmark `xml:"bookmark,omitempty"`
	Hilite   []Hilite   `xml:"hilite,omitempty"`
}

type Hilite struct {
	XMLName     xml.Name `xml:"hilite"`
	HiliteStart HiliteStart
	HiliteEnd   HiliteEnd
	Note        Note
	Label       string `xml:"label,attr"`
}

type HiliteEnd struct {
	XMLName    xml.Name `xml:"hiliteEnd"`
	NcxRef     string   `xml:"ncxRef"`
	URI        string   `xml:"URI"`
	TimeOffset string   `xml:"timeOffset"`
	CharOffset string   `xml:"charOffset"`
}

type HiliteStart struct {
	XMLName    xml.Name `xml:"hiliteStart"`
	NcxRef     string   `xml:"ncxRef"`
	URI        string   `xml:"URI"`
	TimeOffset string   `xml:"timeOffset"`
	CharOffset string   `xml:"charOffset"`
}

type Lastmark struct {
	XMLName    xml.Name `xml:"lastmark,omitempty"`
	NcxRef     string   `xml:"ncxRef"`
	URI        string   `xml:"URI"`
	TimeOffset string   `xml:"timeOffset"`
	CharOffset string   `xml:"charOffset"`
}

type Note struct {
	XMLName xml.Name `xml:"note"`
	Text    string   `xml:"text,omitempty"`
}

type Title struct {
	XMLName xml.Name `xml:"title"`
	Text    string   `xml:"text"`
}
