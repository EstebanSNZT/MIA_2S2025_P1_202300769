package main

import (
	"fmt"
	"server/analyzer"
)

func main() {
	result, err := analyzer.Analyzer("fdisk -name=TamalExtPart -size=1 -unit=M -type=E -path=/home/esteban/Documentos/Projects/MIA_2S2025_P1_202300769/server/disks/Disk1.mia -fit=WF")
	if err != nil {
		fmt.Println("Error", err)
	}

	fmt.Println(result)
}
