package main

import (
	"FileBackup/internal/core"
	"FileBackup/internal/log"
	"os"
)

func main() {
	log.Init(os.Stdout, log.DEBUG)

	a := core.Initialize()
	err := a.ReadLocal("testA")
	if err != nil {
		log.Debug(err)
		return
	}
	err = a.WriteIndex("testA.idx")
	if err != nil {
		log.Debug(err)
		return
	}

	b := core.Initialize()
	err = b.ReadLocal("testB")
	if err != nil {
		log.Debug(err)
		return
	}
	err = b.WriteIndex("testB.idx")
	if err != nil {
		log.Debug(err)
		return
	}

	diff, err := core.Compare(a, b)
	if err != nil {
		log.Debug(err)
		return
	}
	log.Debug("aa:\n", diff.A)
	log.Debug("bb:\n", diff.B)
	log.Debug("diff:\n", diff.Diff)

	//fmt.Println(backup)
	//
	//curr := backup.GetNodeAtPath(".")
	//if curr != nil {
	//	fmt.Println(curr)
	//	for _, bytes := range curr.File {
	//		fmt.Printf("%x", bytes)
	//	}
	//}

	//backup, err = core.ReadIndex("index.idx")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//err = backup.WriteIndex("index2.idx")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
}
