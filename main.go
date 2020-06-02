package main

import (
	"fmt"
	"os"
	"syscall"
	"os/exec"
	"path/filepath"
	"io/ioutil"
	"strconv"
)
//go run main.go run echo hello

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help?")
	}
}

func run(){
	fmt.Printf("running in new UTS namespace %v as %d\n", os.Args[2:],os.Getpid())
	
	cmd := exec.Command("/proc/self/exe",append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stderr= os.Stderr
	cmd.Stdout = os.Stdout

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,//We do this because we don't want namespace to share the info with the host
	}

		if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

}
func child(){
	fmt.Printf("running in new UTS namespace %v as %d\n", os.Args[2:],os.Getpid())

	cg()

	syscall.Sethostname([]byte("inside-container"))
	must(syscall.Chroot("/root/rootfs"))
	// Change directory after chroot
	must(syscall.Chdir("/"))
	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(os.Args[2],os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr= os.Stderr
	cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	
	syscall.Unmount("/proc",0)




}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "cgg"), 0755)
	must(ioutil.WriteFile(filepath.Join(pids, "cgg/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	must(ioutil.WriteFile(filepath.Join(pids, "cgg/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "cgg/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error){
	if err !=nil{
		panic(err)
	}
