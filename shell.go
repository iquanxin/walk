package walk

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/iquanxin/walk/win"
)

// ShellExecuteNoWait is version of ShellExecuteEx which don't want process
func ShellExecuteNowait(hwnd win.HWND, lpOperation, lpFile, lpParameters, lpDirectory string, nShowCmd int) error {
	var lpctstrVerb, lpctstrParameters, lpctstrDirectory uintptr
	if len(lpOperation) != 0 {
		lpctstrVerb = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpOperation)))
	}
	if len(lpParameters) != 0 {
		lpctstrParameters = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpParameters)))
	}
	if len(lpDirectory) != 0 {
		lpctstrDirectory = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpDirectory)))
	}
	i := &win.SHELLEXECUTEINFO{
		FMask:        win.SEE_MASK_DEFAULT, // _SEE_MASK_NOCLOSEPROCESS
		Hwnd:         hwnd,
		LpVerb:       lpctstrVerb,
		LpFile:       uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpFile))),
		LpParameters: lpctstrParameters,
		LpDirectory:  lpctstrDirectory,
		NShow:        nShowCmd,
	}
	i.CbSize = uint32(unsafe.Sizeof(*i))
	return _ShellExecuteEx(i)
}

// ShellExecuteEx is Windows API
func _ShellExecuteEx(pExecInfo *win.SHELLEXECUTEINFO) error {
	ret := win.ShellExecuteEx(pExecInfo)
	if ret && pExecInfo.FMask&win.SEE_MASK_NOCLOSEPROCESS != 0 {
		s, e := syscall.WaitForSingleObject(syscall.Handle(pExecInfo.HProcess), syscall.INFINITE)
		switch s {
		case syscall.WAIT_OBJECT_0:
			break
		case syscall.WAIT_FAILED:
			return os.NewSyscallError("WaitForSingleObject", e)
		default:
			return errors.New("unexpected result from wait for single object")
		}
	}
	errorMsg := ""
	if pExecInfo.HInstApp != 0 && pExecInfo.HInstApp <= 32 {
		switch int(pExecInfo.HInstApp) {
		case win.SE_ERR_FNF:
			errorMsg = "The specified file was not found"
		case win.SE_ERR_PNF:
			errorMsg = "The specified path was not found"
		case win.ERROR_BAD_FORMAT:
			errorMsg = "The .exe file is invalid (non-Win32 .exe or error in .exe image)"
		case win.SE_ERR_ACCESSDENIED:
			errorMsg = "The operating system denied access to the specified file"
		case win.SE_ERR_ASSOCINCOMPLETE:
			errorMsg = "The file name association is incomplete or invalid"
		case win.SE_ERR_DDEBUSY:
			errorMsg = "The DDE transaction could not be completed because other DDE transactions were being processed"
		case win.SE_ERR_DDEFAIL:
			errorMsg = "The DDE transaction failed"
		case win.SE_ERR_DDETIMEOUT:
			errorMsg = "The DDE transaction could not be completed because the request timed out"
		case win.SE_ERR_DLLNOTFOUND:
			errorMsg = "The specified DLL was not found"
		case win.SE_ERR_NOASSOC:
			errorMsg = "There is no application associated with the given file name extension"
		case win.SE_ERR_OOM:
			errorMsg = "There was not enough memory to complete the operation"
		case win.SE_ERR_SHARE:
			errorMsg = "A sharing violation occurred"
		default:
			errorMsg = fmt.Sprintf("Unknown error occurred with error code %v", pExecInfo.HInstApp)
		}
	} else {
		return nil
	}
	return errors.New(errorMsg)
}
