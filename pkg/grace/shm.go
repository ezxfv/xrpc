package grace

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	// https://code.woboq.org/userspace/glibc/sysdeps/unix/sysv/linux/bits/ipc.h.html
	// Mode bits for `msgget', `semget', and `shmget'.
	IPC_CREAT  = 01000 // Create key if key does not exist.
	IPC_EXCL   = 02000 // Fail if key exists.
	IPC_NOWAIT = 04000 // Return error on wait.
	IPC_NORMAL = 00000

	// Control commands for `msgctl', `semctl', and `shmctl'.
	IPC_RMID = 0 // Remove identifier.
	IPC_SET  = 1 // Set `ipc_perm' options.
	IPC_STAT = 2 // Get `ipc_perm' options.
	IPC_INFO = 3 // See ipcs.

	// Special key values.
	IPC_PRIVATE = 0 // Private key. NOTE: this value is of type __key_t, i.e., ((__key_t) 0)
)

func ShmSet(id int, val int) uintptr{
	shmId := genShmId(id, IPC_CREAT)
	shmAddr := genShmAddr(shmId)
	defer syscall.Syscall(syscall.SYS_SHMDT, shmAddr, 0, 0)
	*(*int)(unsafe.Pointer(shmAddr)) = val
	return shmId
}

func ShmGet(id int) int{
	shmId := genShmId(id, IPC_NORMAL)
	shmAddr := genShmAddr(shmId)
	defer syscall.Syscall(syscall.SYS_SHMDT, shmAddr, 0, 0)
	return *(*int)(unsafe.Pointer(shmAddr))
}

func ShmDel(id int) string {
	shmId := genShmId(id, IPC_NORMAL)
	_, _, errno := syscall.Syscall(syscall.SYS_SHMCTL, shmId, 0, 0)
	return errno.Error()
}

func genShmId(id int, ipcFlag int) uintptr {
	shmId, _, err := syscall.Syscall(syscall.SYS_SHMGET, uintptr(2+id), 4, uintptr(ipcFlag|0600))
	if err != 0 {
		fmt.Printf("syscall error, err: %v\n", err)
		os.Exit(-1)
	}
	return shmId
}

func genShmAddr(shmId uintptr) uintptr {
	shmAddr, _, err := syscall.Syscall(syscall.SYS_SHMAT, shmId, 0, 0)
	if err != 0 {
		fmt.Printf("syscall error, err: %v\n", err)
		os.Exit(-2)
	}
	return shmAddr
}