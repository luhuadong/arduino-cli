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

package daemon

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/rpc/commands"
	"io"
	"os/exec"
)

// DebugService implements the `Debug` service
type DebugService struct{}

// StreamingOpen returns a stream response that can be used to fetch data from the
// Debug target. The first message passed through the `StreamingOpenReq` must
// contain Debug configuration params, not data.
func (s *DebugService) StreamingOpen(stream commands.Debug_StreamingOpenServer) error {
	// grab the first message
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	// ensure it's a config message and not data
	compileReq := msg.GetCompileReq()
	if compileReq == nil {
		return fmt.Errorf("first message must contain a CompileReq, not data")
	}
	// compile the sketch
	var compileOut bytes.Buffer
	compileOutBuf := bufio.NewWriter(&compileOut)
	var compileErr bytes.Buffer
	compileErrBuf := bufio.NewWriter(&compileErr)
	_, err = compile.Compile(stream.Context(), compileReq, compileOutBuf, compileErrBuf, false)
	if err != nil {
		return (err)
	}
	// re-hydrate debug recipe using compile params to identify the recipe
	getDebugCommand(compileReq)
	// launch debug recipe attaching stdin and out to grpc streaming
	cmd := exec.Command("gdb")
	in, err := cmd.StdinPipe()
	if err != nil {
		return (err)
	}
	defer in.Close()

	out, err := cmd.StdoutPipe()
	if err != nil {
		return (err)
	}
	defer out.Close()

	err = cmd.Start()
	if err != nil {
		fmt.Println("%v\n", err)
		return err
	}

	// we'll use these channels to communicate with the goroutines
	// handling the stream and the target respectively
	streamClosed := make(chan error)
	targetClosed := make(chan error)

	// now we can read the other commands and re-route to the Debug Client...
	go func() {
		for {
			command, err := stream.Recv()
			if err == io.EOF {
				// stream was closed
				streamClosed <- nil
				break
			}

			if err != nil {
				// error reading from stream
				streamClosed <- err
				break
			}

			if _, err := in.Write(command.GetData()); err != nil {
				// error writing to target
				targetClosed <- err
				fmt.Println("in.Write(command.GetData())")
				break
			}
		}
	}()

	// ...and read from the Debug and forward to the output stream
	go func() {
		buf := make([]byte, 8)
		for {
			n, err := out.Read(buf)
			if err != nil {
				// error reading from target
				targetClosed <- err
				break
			}

			if n == 0 {
				// target was closed
				targetClosed <- nil
				break
			}

			err = stream.Send(&commands.StreamingOpenResp{
				Data: buf[:n],
			})
			if err != nil {
				// error sending to stream
				streamClosed <- err
				break
			}
		}
	}()

	// let goroutines route messages from/to the Debug
	// until either the client closes the stream or the
	// Debug target is closed
	for {
		select {
		case err := <-streamClosed:
			fmt.Println("streamClosed")
			cmd.Process.Kill()
			cmd.Wait()
			return err
		case err := <-targetClosed:
			fmt.Println("targetClosed")
			return err
		}
	}
}

func getDebugCommand(req *commands.CompileReq) {

}
