// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package builder

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	properties "github.com/arduino/go-properties-orderedmap"
)

// ArduinoPreprocessorProperties are the platform properties needed to run arduino-preprocessor
var ArduinoPreprocessorProperties = properties.NewFromHashmap(map[string]string{
	// Ctags
	"tools.arduino-preprocessor.path":     "{runtime.tools.arduino-preprocessor.path}",
	"tools.arduino-preprocessor.cmd.path": "{path}/arduino-preprocessor",
	"tools.arduino-preprocessor.pattern":  `"{cmd.path}" "{source_file}" "{codecomplete}" -- -std=gnu++11`,

	"preproc.macros.flags": "-w -x c++ -E -CC",
})

type PreprocessSketchArduino struct{}

func (s *PreprocessSketchArduino) Run(ctx *types.Context) error {
	sourceFile := ctx.SketchBuildPath.Join(ctx.Sketch.MainFile.Name.Base() + ".cpp")
	commands := []types.Command{
		&ArduinoPreprocessorRunner{},
	}

	if err := ctx.PreprocPath.MkdirAll(); err != nil {
		return i18n.WrapError(err)
	}

	GCCPreprocRunner(ctx, sourceFile, ctx.PreprocPath.Join(constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E), ctx.IncludeFolders)

	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			return i18n.WrapError(err)
		}
	}

	var err error
	if ctx.CodeCompleteAt != "" {
		err = new(OutputCodeCompletions).Run(ctx)
	} else {
		err = bldr.SketchSaveItemCpp(ctx.Sketch.MainFile.Name.String(), []byte(ctx.Source), ctx.SketchBuildPath.String())
	}

	return err
}

type ArduinoPreprocessorRunner struct{}

func (s *ArduinoPreprocessorRunner) Run(ctx *types.Context) error {
	buildProperties := ctx.BuildProperties
	targetFilePath := ctx.PreprocPath.Join(constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E)
	logger := ctx.GetLogger()

	properties := buildProperties.Clone()
	toolProps := buildProperties.SubTree("tools").SubTree("arduino-preprocessor")
	properties.Merge(toolProps)
	properties.SetPath(constants.BUILD_PROPERTIES_SOURCE_FILE, targetFilePath)
	if ctx.CodeCompleteAt != "" {
		if runtime.GOOS == "windows" {
			//use relative filepath to avoid ":" escaping
			splt := strings.Split(ctx.CodeCompleteAt, ":")
			if len(splt) == 3 {
				//all right, do nothing
			} else {
				splt[1] = filepath.Base(splt[0] + ":" + splt[1])
				ctx.CodeCompleteAt = strings.Join(splt[1:], ":")
			}
		}
		properties.Set("codecomplete", "-output-code-completions="+ctx.CodeCompleteAt)
	} else {
		properties.Set("codecomplete", "")
	}

	pattern := properties.Get(constants.BUILD_PROPERTIES_PATTERN)
	if pattern == constants.EMPTY_STRING {
		return i18n.ErrorfWithLogger(logger, constants.MSG_PATTERN_MISSING, "arduino-preprocessor")
	}

	commandLine := properties.ExpandPropsInString(pattern)
	command, err := utils.PrepareCommand(commandLine, logger, "")
	if err != nil {
		return i18n.WrapError(err)
	}

	if runtime.GOOS == "windows" {
		// chdir in the uppermost directory to avoid UTF-8 bug in clang (https://github.com/arduino/arduino-preprocessor/issues/2)
		command.Dir = filepath.VolumeName(command.Args[0]) + "/"
		//command.Args[0], _ = filepath.Rel(command.Dir, command.Args[0])
	}

	verbose := ctx.Verbose
	if verbose {
		fmt.Println(commandLine)
	}

	buf, err := command.Output()
	if err != nil {
		return errors.New(i18n.WrapError(err).Error() + string(err.(*exec.ExitError).Stderr))
	}

	result := utils.NormalizeUTF8(buf)

	//fmt.Printf("PREPROCESSOR OUTPUT:\n%s\n", output)
	if ctx.CodeCompleteAt != "" {
		ctx.CodeCompletions = string(result)
	} else {
		ctx.Source = string(result)
	}
	return nil
}

type OutputCodeCompletions struct{}

func (s *OutputCodeCompletions) Run(ctx *types.Context) error {
	if ctx.CodeCompletions == "" {
		// we assume it is a json, let's make it compliant at least
		ctx.CodeCompletions = "[]"
	}
	ctx.GetLogger().Println(constants.LOG_LEVEL_INFO, ctx.CodeCompletions)
	return nil
}
