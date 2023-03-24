package respository

import (
	"sfilesvr/body/model"

	"github.com/oswaldoooo/octools/datastore"
	"github.com/oswaldoooo/octools/toolsbox"
)

type AdvancedRepository interface {
	SearchFile(string) []string
}
type AdvancedRepositoryImp struct {
	Isprivate  bool
	PrivateDir string
}

// search service for public and private zone file link
func (s *AdvancedRepositoryImp) SearchFile(filename string) (res []string) {
	//usr search in private zone
	if s.Isprivate {
		dirmap := model.ParseDirMap()
		if dirid, ok := dirmap[s.PrivateDir]; ok {
			//if private dir exist
			resmap := model.ParseFileMap("privatemap/" + dirid)
			newreslist := toolsbox.ExportMapValue(resmap)
			res = datastore.BinarySearch(filename, newreslist)
		}
	} else {
		//usr search in public zone
		resmap := model.ParseFileMap("filemap")
		newreslist := toolsbox.ExportMapValue(resmap)
		res = datastore.BinarySearch(filename, newreslist)
	}
	return
}
