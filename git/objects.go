package git

type Object interface {
}

const (
	OBJ_COMMIT    = 1
	OBJ_TREE      = 2
	OBJ_BLOB      = 3
	OBJ_TAG       = 4
	OBJ_OFS_DELTA = 6
	OBJ_REF_DELTA = 7
)

type Commit struct {
	Content []byte
}

type Tree struct {
	Content []byte
}

type Blob struct {
	Content []byte
}

type Tag struct {
	Content []byte
}
