package main

import (
	"fmt"
	"server/analyzer"
)

func main() {
	result, err := analyzer.Analyzer("mount -name=MiloPart -path=/home/esteban/Documentos/Projects/MIA_2S2025_P1_202300769/server/disks/Disk1.mia")
	if err != nil {
		fmt.Println("Error", err)
	}

	fmt.Println(result)
}
