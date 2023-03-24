package model

import "encoding/xml"

// this is configure file types collections
type CnfInfo struct {
	XMLName      xml.Name    `xml:"sfile_server"`
	DatabaseInfo DatabaseCnf `xml:"database_info"`
	PluginInfo   PluginCnf   `xml:"plugin_info"`
}
type DatabaseCnf struct {
	XMLName xml.Name `xml:"database_info"`
	Dbinfo  []Dbinfo `xml:"db"`
}
type Dbinfo struct {
	XMLName  xml.Name `xml:"db"`
	Class    string   `xml:"class"`
	User     string   `xml:"user"`
	Password string   `xml:"password"`
	Address  string   `xml:"address"`
	DBname   string   `xml:"database"`
	Charset  string   `xml:"charset"`
}
type PluginCnf struct {
	XMLName     xml.Name `xml:"plugin_info"`
	Plugin_Info []plugin `xml:"plugin"`
}
type plugin struct {
	XMLName   xml.Name `xml:"plugin"`
	ClassName string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
}
