#!/usr/bin/env python
# Copyright (c) 2012 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

""" This script is a wrapper for running GM on an Android device.  First, we
prepare the device by clearing and creating the output image directory, then we
run GM, and then we pull the output images from the device to the host. """

from android_build_step import AndroidBuildStep
from build_step import BuildStep
from run_gm import RunGM
from utils import android_utils
import os
import shutil
import sys


BINARY_NAME = 'gm'
ANDROID_GM_ARGS = ['--nopdf']


class AndroidRunGM(AndroidBuildStep, RunGM):
  def _PullImages(self, serial):
    """ Pull images generated by gm from the device to the host.

    serial: string indicating the serial number of the target device
    """

    # Here we make the assumption that nobody else is messing with this area of
    # the file system.
    try:
      shutil.rmtree(self._gm_actual_dir)
    except Exception:
      pass
    os.makedirs(self._gm_actual_dir)
    android_utils.RunADB(serial, ['pull', '%s/%s' % (self._device_dirs.GMDir(),
                                                     self._gm_image_subdir),
                                  self._gm_actual_dir])
    android_utils.RunADB(serial, ['shell', 'rm', '-r', '%s/%s' % (
                                      self._device_dirs.GMDir(),
                                      self._gm_image_subdir)
                                  ])

  def _Run(self):
    self._PreGM()
    try:
      android_utils.RunADB(self._serial, ['shell', 'rm', '-r',
                                          '%s/%s' % (self._device_dirs.GMDir(),
                                                     self._gm_image_subdir)])
    except Exception:
      pass
    android_utils.RunADB(self._serial, ['shell', 'mkdir', '-p',
                                        '%s/%s' % (self._device_dirs.GMDir(),
                                                   self._gm_image_subdir)])
    arguments = (ANDROID_GM_ARGS +
                 ['-w', '%s/%s' % (self._device_dirs.GMDir(),
                                   self._gm_image_subdir)]
                 + self._gm_args)
    try:
      android_utils.StopShell(self._serial)
      android_utils.RunADB(self._serial, ['logcat', '-c'])
      cmd = [android_utils.PATH_TO_ADB, '-s', self._serial, 'shell',
             'skia_launcher', 'gm'] + arguments
      self._RunModulo(cmd)
    finally:
      android_utils.RunADB(self._serial, ['logcat', '-d', '-v', 'time'])
      self._PullImages(self._serial)


if '__main__' == __name__:
  sys.exit(BuildStep.RunBuildStep(AndroidRunGM))
