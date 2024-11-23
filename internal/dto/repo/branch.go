package repo

type BranchInit struct {
	Name     string `json:"name" validate:"required,min=1,max=100,excludesall= "`
	ParentId int32  `json:"parentId" validate:"required,numeric"`
}
