package structures

import (
	"bytes"
	"fmt"
	"html"
	"strings"
	"unicode"
)

type FolderBlock struct {
	Content [4]FolderContent
}

type FolderContent struct {
	Name  [12]byte
	Inode int32
}

type FileBlock struct {
	Content [64]byte
}

type PointerBlock struct {
	Pointers [16]int32
}

func NewFileBlock(data []byte, start int32, blockSize int32) (*FileBlock, error) {
	fileSize := int32(len(data))
	if start >= fileSize {
		return nil, fmt.Errorf("inicio del índice fuera de rango")
	}

	end := min(start+blockSize, fileSize)
	fileBlock := &FileBlock{}
	copy(fileBlock.Content[:], data[start:end])
	return fileBlock, nil
}

func NewPointerBlock() *PointerBlock {
	return &PointerBlock{
		Pointers: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
	}
}

func NewFolderBlock() *FolderBlock {
	return &FolderBlock{
		Content: [4]FolderContent{
			{Name: [12]byte{'-'}, Inode: -1},
			{Name: [12]byte{'-'}, Inode: -1},
			{Name: [12]byte{'-'}, Inode: -1},
			{Name: [12]byte{'-'}, Inode: -1},
		},
	}
}

func (b *FolderBlock) GenerateTable(index int32) string {
	var rows strings.Builder
	for i, entry := range b.Content {
		name := strings.TrimRight(string(entry.Name[:]), "\x00")
		if name == "" {
			name = "-"
		}
		rows.WriteString(fmt.Sprintf(`<tr><td bgcolor="#ffe3fbff"><b>%s</b></td><td PORT="i%d" bgcolor="#fff2d0ff">%d</td></tr>`, name, i, entry.Inode))
	}

	return fmt.Sprintf(`
		block%d [label=<
			<table border="0" cellborder="1" cellspacing="0">
			<tr><td PORT="top" colspan="2" bgcolor="#ff6d59ff"><b>Bloque de Carpeta %d</b></td></tr>
			<tr><td bgcolor="#ff6eecff"><b>Nombre</b></td><td bgcolor="#ffd35cff">Inodo</td></tr>
			%s
			</table>
		>];
	`, index, index, rows.String())
}

func (b *FileBlock) GenerateTable(index int32) string {
	rawContent := string(bytes.TrimRight(b.Content[:], "\x00"))
	var cleanContentBuilder strings.Builder
	for _, r := range rawContent {
		if unicode.IsPrint(r) {
			cleanContentBuilder.WriteRune(r)
		}
	}
	contentSnippet := cleanContentBuilder.String()
	escapedContent := html.EscapeString(contentSnippet)

	if len(escapedContent) > 32 {
		escapedContent = escapedContent[:32] + "<br/>" + escapedContent[32:]
	}

	return fmt.Sprintf(`
		block%d [label=<
			<table border="0" cellborder="1" cellspacing="0">
			<tr><td PORT="top" bgcolor="#76ff76ff"><b>Bloque de Archivo %d</b></td></tr>
			<tr><td align="left">%s</td></tr>
			</table>
		>];
	`, index, index, escapedContent)
}

func (b *PointerBlock) GenerateTable(index int32) string {
	var rows strings.Builder
	for i, ptr := range b.Pointers {
		rows.WriteString(fmt.Sprintf(`<tr><td PORT="ptr%d">%d</td></tr>`, i, ptr))
	}

	return fmt.Sprintf(`
		block%d [label=<
			<table border="0" cellborder="1" cellspacing="0">
			<tr><td PORT="top" colspan="2" bgcolor="#ffe8c3"><b>Bloque de Punteros %d</b></td></tr>
			%s
			</table>
		>];
	`, index, index, rows.String())
}
