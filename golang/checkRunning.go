package run

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

var (
	LockDir      = "/tmp/run/"
	BuildVersion string
	BuildDate    string
)

// check and set file lock
func FileLock(fd *os.File) (err error) {
	err = syscall.Flock(int(fd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) //this method just work on unix like system
	return err
}

func ShowVersion() {
	fmt.Printf("Git commit:%s\n", BuildVersion)
	fmt.Printf("Build time:%s\n", BuildDate)
}

// check server , if already running return ,otherwise get current pid and write to file.
func CheckRun(name string) (err error) {
	fmt.Println("CheckRun.......")
	ShowVersion()

	pidFile := LockDir + name + ".pid"

	fd, err := os.OpenFile(pidFile, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		fmt.Println("open file error:", err)
		return err
	}
	//defer fd.Close()
	err = FileLock(fd)
	if err != nil {
		fmt.Println(err)
		return err
	}

	npid := os.Getpid()

	_ = fd.Truncate(0)

	_, err = fd.WriteString(strconv.Itoa(npid))
	_ = fd.Sync() //flush to disk
	fmt.Println("current pid:", npid)
	return
}
