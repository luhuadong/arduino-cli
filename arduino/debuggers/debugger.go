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

package debuggers

const (
	defaultBaudRate = 9600
)

// SerialMonitor is a monitor for serial ports
type Debugger struct {
}

// OpenDebugger creates a debugger instance for a debug session
func OpenDebugger() (*Debugger, error) {

	return &Debugger{}, nil
}

// Close the connection
func (mon *Debugger) Close() error {
	return nil
}

// Read bytes from the stdin
func (mon *Debugger) Read(bytes []byte) (int, error) {
	return 0, nil
}

// Write bytes to the stdout
func (mon *Debugger) Write(bytes []byte) (int, error) {
	return 0, nil
}
