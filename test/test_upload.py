# This file is part of arduino-cli.
#
# Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
#
# This software is released under the GNU General Public License version 3,
# which covers the main part of arduino-cli.
# The terms of this license can be found at:
# https://www.gnu.org/licenses/gpl-3.0.en.html
#
# You can be released from the requirements of the above licenses by purchasing
# a commercial license. Buying such a license is mandatory if you want to modify or
# otherwise use the software for commercial activities involving the Arduino
# software without disclosing the source code of your own applications. To purchase
# a commercial license, send an email to license@arduino.cc.
import os

import pytest

from .common import running_on_ci

# Skip this module when running in CI environments
pytestmark = pytest.mark.skipif(running_on_ci(), reason="VMs have no serial ports")


def test_upload(run_command, data_dir, detected_boards):
    # Init the environment explicitly
    assert run_command("core update-index")

    for board in detected_boards:
        # Download core
        assert run_command("core install {}".format(board.core))
        # Create a sketch
        sketch_path = os.path.join(data_dir, "foo")
        assert run_command("sketch new {}".format(sketch_path))
        # Build sketch
        assert run_command(
            "compile -b {fqbn} {sketch_path}".format(
                fqbn=board.fqbn, sketch_path=sketch_path
            )
        )
        # Upload
        assert run_command(
            "upload -b {fqbn} -p {port} {sketch_path}".format(
                sketch_path=sketch_path, fqbn=board.fqbn, port=board.address
            )
        )
