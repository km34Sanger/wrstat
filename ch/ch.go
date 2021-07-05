/*******************************************************************************
 * Copyright (c) 2021 Genome Research Ltd.
 *
 * Author: Sendu Bala <sb10@sanger.ac.uk>
 *
 * Permission is hereby granted, free of charge, to any person obtaining
 * a copy of this software and associated documentation files (the
 * "Software"), to deal in the Software without restriction, including
 * without limitation the rights to use, copy, modify, merge, publish,
 * distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to
 * the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
 * CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/

// package ch is used to do chmod and chown on certain files, to correct for
// group and user permissions and ownership being wrong.

package ch

import (
	"io/fs"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
)

const modePermUser = 0700
const modePermGroup = 0070
const modePermUserToGroupShit = 3

// PathChecker is a callback used by Ch that will receive the absolute path to a
// file or directory and should return a boolean if this path is eligible for
// changing, and the desired group ID of this path.
type PathChecker func(path string) (change bool, gid int)

// Ch is used to chmod and chown files such that they match their desired group.
type Ch struct {
	pc     PathChecker
	logger log15.Logger
}

// New returns a Ch what will check your pc callback to see what work needs to
// be done on the paths this Ch will receive when Do() is called on it.
//
// Changes made will be logged to the given logger.
func New(pc PathChecker, logger log15.Logger) *Ch {
	return &Ch{
		pc:     pc,
		logger: logger,
	}
}

// Do is a github.com/wtsi-ssg/wrstat/stat Operation that passes path to our
// PathCheck callback, and if it returns true, does the following chmod and
// chown-type behaviours, making use of the supplied Lstat info to avoid doing
// unnecessary repeated work:
//
// 1. Ensures that the GID of the path is the returned GID.
// 2. If path is a directory, ensures it has setgid applied (group sticky).
// 3. Ensures that group permissions match user permissions.
//
// Any errors are returned without logging them. Any changes we do on disk are
// logged to our logger.
func (c *Ch) Do(path string, info fs.FileInfo) error {
	change, gid := c.pc(path)
	if !change {
		return nil
	}

	var merr error

	if err := c.chownGroup(path, getGIDFromFileInfo(info), gid); err != nil {
		merr = multierror.Append(merr, err)
	}

	if err := c.setgid(path, info); err != nil {
		merr = multierror.Append(merr, err)
	}

	if err := c.matchPermissions(path, info); err != nil {
		merr = multierror.Append(merr, err)
	}

	return merr
}

// getGIDFromFileInfo extracts the GID from a FileInfo. NB: this will only work
// on linux.
func getGIDFromFileInfo(info fs.FileInfo) int {
	return int(info.Sys().(*syscall.Stat_t).Gid)
}

// chownGroup chown's path to have newGID as its group owner, if newGID is
// different to origGID. If a change is made, logs it.
func (c *Ch) chownGroup(path string, origGID, newGID int) error {
	if origGID == newGID {
		return nil
	}

	if err := os.Chown(path, -1, newGID); err != nil {
		return err
	}

	origName, err := groupName(origGID)
	if err != nil {
		return err
	}

	newName, err := groupName(newGID)
	if err != nil {
		return err
	}

	c.logger.Info("changed group", "path", path, "orig", origName, "new", newName)

	return nil
}

// groupName returns the name of the group with the given GID.
func groupName(gid int) (string, error) {
	g, err := user.LookupGroupId(strconv.Itoa(gid))
	if err != nil {
		return "", err
	}

	return g.Name, err
}

// setgid sets group sticky bit on path if path is a dir and didn't already have
// group sticky bit set. If a change is made, logs it.
func (c *Ch) setgid(path string, info fs.FileInfo) error {
	if !info.IsDir() || setgidApplied(info) {
		return nil
	}

	err := os.Chmod(path, info.Mode()|os.ModeSetgid)
	if err != nil {
		return err
	}

	c.logger.Info("applied setgid", "path", path)

	return nil
}

// setgidApplied reports if the setgid bits are set on the given FileInfo.
func setgidApplied(info fs.FileInfo) bool {
	return (info.Mode() & os.ModeSetgid) != 0
}

// matchPermissions sets group permissions to match user permissions if they're
// different. If a change is made, logs it.
func (c *Ch) matchPermissions(path string, info fs.FileInfo) error {
	mode := info.Mode()
	userAsGroupPerms := extractUserAsGroupPermissions(mode)

	if userAsGroupPerms == extractGroupPermissions(mode) {
		return nil
	}

	err := os.Chmod(path, mode|userAsGroupPerms)
	if err != nil {
		return err
	}

	c.logger.Info("matched group permissions to user", "path", path)

	return nil
}

// extractUserAsGroupPermissions returns the user permission bits of the given
// mode, shifted as if they were group permissions. If there were no user
// permissions, treated as full permissions.
func extractUserAsGroupPermissions(mode fs.FileMode) fs.FileMode {
	user := mode & modePermUser
	if user == 0 {
		user = modePermUser
	}

	return user >> modePermUserToGroupShit
}

// extractGroupPermissions returns the user permission bits of the given mode.
func extractGroupPermissions(mode fs.FileMode) fs.FileMode {
	return mode & modePermGroup
}