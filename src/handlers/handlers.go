package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"weatherdump/src/handlers/interfaces"
	npoessDecoder "weatherdump/src/protocols/hrd/decoder"
	npoessProcessor "weatherdump/src/protocols/hrd/processor"
	meteorDecoder "weatherdump/src/protocols/lrpt/decoder"
	meteorProcessor "weatherdump/src/protocols/lrpt/processor"
)

// AvailableDecoders shows the currently available decoders for this build.
var AvailableDecoders = interfaces.DecoderMakers{
	"lrpt": {
		"soft": meteorDecoder.NewDecoder,
	},
	"hrd": {
		"soft": npoessDecoder.NewSoftSymbolDecoder,
		"cadu": npoessDecoder.NewCaduDecoder,
		"asm":  npoessDecoder.NewAsmDecoder,
	},
}

// AvailableProcessors shows the currently available processors for this build.
var AvailableProcessors = interfaces.ProcessorMakers{
	"lrpt": meteorProcessor.NewProcessor,
	"hrd":  npoessProcessor.NewProcessor,
}

// GenerateDirectories takes user paths and returns the standard output scheme.
func GenerateDirectories(inputFile string, outputPath string) (string, string) {
	inputFileName := filepath.Base(inputFile)
	inputFileName = strings.TrimSuffix(inputFileName, filepath.Ext(inputFile))
	workingPath := filepath.Dir(inputFile)

	if outputPath != "" {
		workingPath = outputPath
		if _, err := os.Stat(workingPath); os.IsNotExist(err) {
			os.Mkdir(workingPath, os.ModePerm)
		}
	}

	if !strings.Contains(inputFile, "/OUTPUT_") {
		workingPath = fmt.Sprintf("%s/OUTPUT_%s", workingPath, strings.ToUpper(inputFileName))
	}

	if _, err := os.Stat(workingPath); os.IsNotExist(err) {
		os.Mkdir(workingPath, os.ModePerm)
	}

	return workingPath, inputFileName
}
