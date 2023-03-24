package model

import "github.com/oswaldoooo/octools/toolsbox"

// the root dir is rootpath+"/conf/"
func ParseFileMap(filesmallpath string) (resmap map[string]string) {
	resmap, err := toolsbox.ParseList(ROOTPATH + "/conf/" + filesmallpath)
	if err == nil {
		for ke, ve := range resmap {
			if len(ve) != 6 || len(ve) != 4 {
				//only left fileid
				delete(resmap, ke)
			}
		}
	}
	return
}
func ParseDirMap() (resmap map[string]string) {
	resmap, err := toolsbox.ParseList(ROOTPATH + "/conf/filemap")
	if err == nil {
		for ke, ve := range resmap {
			if len(ve) != 3 {
				//only left dirid link
				delete(resmap, ke)
			}
		}
	}
	return
}
