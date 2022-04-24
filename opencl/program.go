package opencl

// #include "opencl.h"
import "C"
import (
	"strings"
	"unsafe"
)

type Program struct {
	program C.cl_program
}

func createProgramWithSource(context Context, programCode string) (Program, error) {
	cs := C.CString(programCode)
	defer C.free(unsafe.Pointer(cs))

	var errInt clError
	program := C.clCreateProgramWithSource(
		context.context,
		1,
		&cs,
		nil,
		(*C.cl_int)(&errInt),
	)
	if errInt != clSuccess {
		return Program{}, clErrorToError(errInt)
	}

	return Program{program}, nil
}

func (p Program) Build(device Device, log *string) error {
	emptyString := C.CString("\x00")
	defer C.free(unsafe.Pointer(emptyString))

	var errInt clError = clError(C.clBuildProgram(
		p.program,
		1,
		&device.deviceID,
		emptyString,
		nil,
		nil,
	))
	if errInt == clSuccess {
		return nil
	}

	// If there was a log provided, get the compiler log. Otherwise just return the error
	if log == nil {
		return clErrorToError(errInt)
	}

	var sizeptr C.ulong
	var buildInfoError = clError(C.clGetProgramBuildInfo(
		p.program,
		device.deviceID,
		C.CL_PROGRAM_BUILD_LOG,
		0,
		nil,
		&sizeptr))

	if buildInfoError == clSuccess {

		size := uint64(sizeptr)
		compilerLog := make([]byte, size)

		buildInfoError = clError(C.clGetProgramBuildInfo(
			p.program,
			device.deviceID,
			C.CL_PROGRAM_BUILD_LOG,
			C.size_t(size),
			unsafe.Pointer(&compilerLog[0]),
			nil))

		*log = strings.TrimRight(string(compilerLog), "\x00")
	}

	return clErrorToError(errInt)
}

func (p Program) Release() {
	C.clReleaseProgram(p.program)
}

func (p Program) CreateKernel(kernelName string) (Kernel, error) {
	return createKernel(p, kernelName)
}
