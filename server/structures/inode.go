package structures

import (
	"fmt"
	"time"
)

type Inode struct {
	UID    int32
	GID    int32
	Size   int32
	Atime  float32
	Ctime  float32
	Mtime  float32
	Blocks [15]int32
	Type   [1]byte
	Perm   [3]byte
}

func NewInode(uid int32, gid int32, size int32, blockType [1]byte) *Inode {
	return &Inode{
		UID:    uid,
		GID:    gid,
		Size:   size,
		Atime:  float32(time.Now().Unix()),
		Ctime:  float32(time.Now().Unix()),
		Mtime:  float32(time.Now().Unix()),
		Blocks: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		Type:   blockType,
		Perm:   [3]byte{'7', '7', '7'},
	}
}

func (i *Inode) PushBlock(blockIndex int32) error {
	for j := 0; j < len(i.Blocks); j++ {
		if i.Blocks[j] == -1 {
			i.Blocks[j] = blockIndex
			return nil
		}
	}
	return fmt.Errorf("no hay espacio en el inodo para más bloques")
}

func (i *Inode) String() string {
	aTime := time.Unix(int64(i.Atime), 0).Format("2006-01-02 15:04:05")
	cTime := time.Unix(int64(i.Ctime), 0).Format("2006-01-02 15:04:05")
	mTime := time.Unix(int64(i.Mtime), 0).Format("2006-01-02 15:04:05")

	return fmt.Sprintf("------ Inode ------\n- Inode:\n- UID: %d\n- GID: %d\n- Size: %d\n- Atime: %s\n- Ctime: %s\n- Mtime: %s\n- Blocks: %v\n- Type: %s\n- Perm: %s\n",
		i.UID,
		i.GID,
		i.Size,
		aTime,
		cTime,
		mTime,
		i.Blocks,
		string(i.Type[:]),
		string(i.Perm[:]),
	)
}

func (i *Inode) UpdateAccessTime() {
	i.Atime = float32(time.Now().Unix())
}

func (i *Inode) UpdateModificationTime() {
	i.Mtime = float32(time.Now().Unix())
}
