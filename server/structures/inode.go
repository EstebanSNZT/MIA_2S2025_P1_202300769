package structures

import (
	"fmt"
	"strings"
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

func NewInode(uid int32, gid int32, size int32, blockType [1]byte, perm [3]byte) *Inode {
	return &Inode{
		UID:    uid,
		GID:    gid,
		Size:   size,
		Atime:  float32(time.Now().Unix()),
		Ctime:  float32(time.Now().Unix()),
		Mtime:  float32(time.Now().Unix()),
		Blocks: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		Type:   blockType,
		Perm:   perm,
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

func (i *Inode) GenerateTable(index int32) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`inode%d [label=<
	<table border="0" cellborder="1" cellspacing="0">
	<tr><td colspan="2" bgcolor="#f0e050ff"><b>Reporte Inodo %d</b></td></tr>`, index, index))

	tableRows := fmt.Sprintf(`
		<tr><td><b>i_uid</b></td><td>%d</td></tr>
        <tr><td bgcolor="#fff8beff"><b>i_gid</b></td><td>%d</td></tr>
        <tr><td bgcolor="#fff8beff"><b>i_size</b></td><td>%d</td></tr>
        <tr><td bgcolor="#fff8beff"><b>i_atime</b></td><td>%s</td></tr>
        <tr><td bgcolor="#fff8beff"><b>i_ctime</b></td><td>%s</td></tr>
        <tr><td bgcolor="#fff8beff"><b>i_mtime</b></td><td>%s</td></tr>
        <tr><td bgcolor="#fff8beff"><b>i_type</b></td><td>%c</td></tr>
        <tr><td bgcolor="#fff8beff"><b>i_perm</b></td><td>%s</td></tr>
		<tr><td colspan="2" bgcolor="#27b0ffff"><b>Bloques</b></td></tr>
	`,
		i.UID,
		i.GID,
		i.Size,
		time.Unix(int64(i.Atime), 0).Format("2006-01-02 15:04:05"),
		time.Unix(int64(i.Ctime), 0).Format("2006-01-02 15:04:05"),
		time.Unix(int64(i.Mtime), 0).Format("2006-01-02 15:04:05"),
		i.Type[0],
		string(i.Perm[:]),
	)
	sb.WriteString(tableRows)

	for j, block := range i.Blocks {
		if j < 12 {
			sb.WriteString(fmt.Sprintf(`<tr><td bgcolor="#bee7ffff"><b>Directo [%d]</b></td><td>%d</td></tr>`, j, block))
		} else if j < 13 {
			sb.WriteString(fmt.Sprintf(`<tr><td bgcolor="#4bdffdff"><b>Indirecto Simple [%d]</b></td><td>%d</td></tr>`, j, block))
		} else if j < 14 {
			sb.WriteString(fmt.Sprintf(`<tr><td bgcolor="#52ffc5ff"><b>Indirecto Doble [%d]</b></td><td>%d</td></tr>`, j, block))
		} else if j < 15 {
			sb.WriteString(fmt.Sprintf(`<tr><td bgcolor="#83ff6aff"><b>Indirecto Triple [%d]</b></td><td>%d</td></tr>`, j, block))
		}
	}

	sb.WriteString(`</table>>];`)
	return sb.String()
}
