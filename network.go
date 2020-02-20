package darknet

// #include <stdlib.h>
//
// #include <darknet.h>
//
// #include "network.h"
import "C"
import (
	"errors"
	"time"
	"unsafe"
)

// YOLONetwork represents a neural network using YOLO.
type YOLONetwork struct {
	GPUDeviceIndex           int
	DataConfigurationFile    string
	NetworkConfigurationFile string
	WeightsFile              string
	Threshold                float32

	ClassNames []string
	Classes    int

	cNet                *C.network
	hierarchalThreshold float32
	nms                 float32
}

var (
	errNetworkNotInit      = errors.New("network not initialised")
	errUnableToInitNetwork = errors.New("unable to initialise")
)

// Init the network.
func (n *YOLONetwork) Init() error {

	nCfg := C.CString(n.NetworkConfigurationFile)
	defer C.free(unsafe.Pointer(nCfg))
	wFile := C.CString(n.WeightsFile)
	defer C.free(unsafe.Pointer(wFile))

	// data := C.CString(n.DataConfigurationFile)
	// defer C.free(unsafe.Pointer(data))

	// inptim := C.CString("/home/dimitrii/Downloads/mega.jpg")
	// defer C.free(unsafe.Pointer(inptim))

	// outputim := C.CString("/home/dimitrii/work/src/github.com/LdDl/go-darknet/example/out-sample.png")
	// defer C.free(unsafe.Pointer(outputim))

	// C.test_detector(data, nCfg, wFile, inptim, 0.4, 0.5, 1, 1, 0, outputim, 0, 0)

	// // GPU device ID must be set before `load_network()` is invoked.
	C.cuda_set_device(C.int(n.GPUDeviceIndex))
	n.cNet = C.load_network(nCfg, wFile, 0)

	if n.cNet == nil {
		return errUnableToInitNetwork
	}

	// C.set_batch_network(n.cNet, 1)
	C.srand(2222222)

	n.hierarchalThreshold = 0.5
	n.nms = 0.45

	metadata := C.get_metadata(nCfg)
	n.Classes = int(metadata.classes)
	n.ClassNames = makeClassNames(metadata.names, n.Classes)

	return nil
}

// Close and release resources.
func (n *YOLONetwork) Close() error {
	if n.cNet == nil {
		return errNetworkNotInit
	}

	C.free_network(*n.cNet)
	n.cNet = nil
	return nil
}

// Detect specified image
func (n *YOLONetwork) Detect(img *DarknetImage) (*DetectionResult, error) {
	if n.cNet == nil {
		return nil, errNetworkNotInit
	}
	startTime := time.Now()
	result := C.perform_network_detect(n.cNet, &img.image, C.int(n.Classes), C.float(n.Threshold), C.float(n.hierarchalThreshold), C.float(n.nms), C.int(0))
	endTime := time.Now()
	defer C.free_detections(result.detections, result.detections_len)
	ds := makeDetections(img, result.detections, int(result.detections_len), n.Threshold, n.Classes, n.ClassNames)
	endTimeOverall := time.Now()
	out := DetectionResult{
		Detections:           ds,
		NetworkOnlyTimeTaken: endTime.Sub(startTime),
		OverallTimeTaken:     endTimeOverall.Sub(startTime),
	}
	return &out, nil
}
