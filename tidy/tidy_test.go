/*
******************************************************************************
* Copyright (c) 2022 Genome Research Ltd.
*
* Author: Sendu Bala <sb10@sanger.ac.uk>
*         Kyle Mace <km34@sanger.ac.uk>
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

Given a srcDir of multi and a destDir of multi/final, and told to work on
go and perl folders, tidy produces:

multi
multi/final
multi/final/20220916_go.cci4au7nu1ibc2ta5j80.cci4au7nu1ibc2ta5j7g.stats.gz
multi/final/20220916_perl.cci4au7nu1ibc2ta5j8g.cci4au7nu1ibc2ta5j7g.stats.gz
multi/final/20220916_go.cci4au7nu1ibc2ta5j80.cci4au7nu1ibc2ta5j7g.byusergroup.gz
multi/final/20220916_perl.cci4au7nu1ibc2ta5j8g.cci4au7nu1ibc2ta5j7g.byusergroup.gz
multi/final/20220916_go.cci4au7nu1ibc2ta5j80.cci4au7nu1ibc2ta5j7g.bygroup
multi/final/20220916_perl.cci4au7nu1ibc2ta5j8g.cci4au7nu1ibc2ta5j7g.bygroup
multi/final/20220916_go.cci4au7nu1ibc2ta5j80.cci4au7nu1ibc2ta5j7g.logs.gz
multi/final/20220916_perl.cci4au7nu1ibc2ta5j8g.cci4au7nu1ibc2ta5j7g.logs.gz
multi/final/20220916_cci4au7nu1ibc2ta5j7g.basedirs
multi/final/20220916_cci4au7nu1ibc2ta5j7g.dgut.dbs
multi/final/20220916_cci4au7nu1ibc2ta5j7g.dgut.dbs/0
multi/final/20220916_cci4au7nu1ibc2ta5j7g.dgut.dbs/0/dgut.db
multi/final/20220916_cci4au7nu1ibc2ta5j7g.dgut.dbs/0/dgut.db.children
multi/final/20220916_cci4au7nu1ibc2ta5j7g.dgut.dbs/1
multi/final/20220916_cci4au7nu1ibc2ta5j7g.dgut.dbs/1/dgut.db
multi/final/20220916_cci4au7nu1ibc2ta5j7g.dgut.dbs/1/dgut.db.children
multi/final/.dgut.dbs.updated

Before running tidy, the srcDir looks like:

multi/cci4fafnu1ia052l75sg
multi/cci4fafnu1ia052l75sg/go
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/walk.1
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/walk.2
[...]
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/walk.1.log
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/walk.1.stats
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/walk.1.byusergroup
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/walk.1.bygroup
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/walk.1.dgut
[...]
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/combine.log.gz
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/combine.byusergroup.gz
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/combine.bygroup
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/combine.dgut.db
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/combine.dgut.db/dgut.db
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/combine.dgut.db/dgut.db.children
multi/cci4fafnu1ia052l75sg/go/cci4fafnu1ia052l75t0/combine.stats.gz
[...]
multi/cci4fafnu1ia052l75sg/perl
multi/cci4fafnu1ia052l75sg/perl/cci4fafnu1ia052l75tg
multi/cci4fafnu1ia052l75sg/perl/cci4fafnu1ia052l75tg/walk.1
[...]
multi/cci4fafnu1ia052l75sg/base.dirs

*****************************************************************************
*/
package tidy

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// modeRW are the read-write permission bits for user, group and other.
const modeRW = 0666

func TestTidy(t *testing.T) {
	date := "20220829"
	srcUniversal := "cci4fafnu1ia052l75sg"
	srcUniqueGo := "cci4fafnu1ia052l75t0"
	srcUniquePerl := "cci4fafnu1ia052l75tg"
	destUniversal := "cci4au7nu1ibc2ta5j7g"
	destUniqueGo := "cci4au7nu1ibc2ta5j80"
	destUniquePerl := "cci4au7nu1ibc2ta5j8g"

	Convey("Given existing source and dest dirs you can tidy the source", t, func() {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		destDir := filepath.Join(tmpDir, "dest")
		interestUniqueDir1 := createTestPath([]string{srcDir, srcUniversal, "go", srcUniqueGo})
		interestUniqueDir2 := createTestPath([]string{srcDir, srcUniversal, "perl", srcUniquePerl})

		buildSrcDir(srcDir, srcUniversal, srcUniqueGo, srcUniquePerl, interestUniqueDir1, interestUniqueDir2)

		createTestDirWithDifferentPerms(destDir)

		err := Up(srcDir, destDir, date)
		So(err, ShouldBeNil)

		Convey("And the combine files are moved from the source dir to the dest dir", func() {
			combineFileSuffixes := [4]string{".logs.gz", ".byusergroup.gz", ".bygroup", ".stats.gz"}

			for i := range combineFileSuffixes {
				final1 := filepath.Join(destDir, date+"_go."+destUniqueGo+"."+destUniversal+combineFileSuffixes[i])
				_, err = os.Stat(final1)
				So(err, ShouldBeNil)

				final2 := filepath.Join(destDir, date+"_perl."+destUniquePerl+"."+destUniversal+combineFileSuffixes[i])
				_, err = os.Stat(final2)
				So(err, ShouldBeNil)
			}
		})

		Convey("And the the contents of the .basedirs and .dgut.dbs dir exist", func() {
			dbsPath := filepath.Join(destDir, date+"_"+destUniversal)
			dbsSuffixes := [5]string{
				".basedirs",
				".dgut.dbs/0/dgut.db",
				".dgut.dbs/0/dgut.db.children",
				".dgut.dbs/1/dgut.db",
				".dgut.dbs/1/dgut.db.children"}

			for i := range dbsSuffixes {
				_, err = os.Stat(dbsPath + dbsSuffixes[i])
				So(err, ShouldBeNil)
			}
		})

		Convey("And the .dgut.dbs.updated file exists in the dest dir", func() {
			expectedFileName := filepath.Join(destDir, ".dgut.dbs.updated")

			_, err = os.Stat(expectedFileName)
			So(err, ShouldBeNil)
		})

		Convey("And the mtime of the .dgut.dbs file matches the oldest mtime of the walk log files", func() {
			newMtimeFile := filepath.Join(interestUniqueDir1, "walk.1.log")
			mTime := time.Date(2006, time.April, 1, 3, 4, 5, 0, time.UTC)
			aTime := time.Date(2007, time.March, 2, 4, 5, 6, 0, time.UTC)

			err = os.Chtimes(newMtimeFile, aTime, mTime)
			So(err, ShouldBeNil)

			err = Up(srcDir, destDir, date)
			So(err, ShouldBeNil)

			dbsFileMTime := getMTime(filepath.Join(destDir, "dgut.dbs.updated"))

			So(mTime, ShouldEqual, dbsFileMTime)
		})

		Convey("And the moved file permissions match those of the dest dir", func() {
			destDirPerm, errs := os.Stat(destDir)
			So(errs, ShouldBeNil)

			err = filepath.WalkDir(destDir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				pathPerm, err := os.Stat(path)
				if err != nil {
					return err
				}

				So(permissionsAndOwnershipSame(destDirPerm, pathPerm), ShouldBeTrue)

				return nil
			})
			So(err, ShouldBeNil)
		})

		Convey("Up deletes the source directory after the files have been moved", func() {
			_, err = os.Stat(srcDir)
			So(err, ShouldNotBeNil)
		})

		Convey("It also works if the dest dir doesn't exist", func() {
			err := os.RemoveAll(destDir)
			So(err, ShouldBeNil)

			err = os.RemoveAll(srcDir)
			So(err, ShouldBeNil)

			buildSrcDir(srcDir, srcUniversal, srcUniqueGo, srcUniquePerl, interestUniqueDir1, interestUniqueDir2)

			err = Up(srcDir, destDir, date)
			So(err, ShouldBeNil)

			_, err = os.Stat(destDir)
			So(err, ShouldBeNil)
		})

		Convey("It doesn't work if source dir doesn't exist", func() {
			err := os.RemoveAll(srcDir)
			So(err, ShouldBeNil)

			err = Up(srcDir, destDir, date)
			So(err, ShouldNotBeNil)

			_, err = os.Stat(srcDir)
			So(err, ShouldNotBeNil)
		})

		Convey("It doesn't work if source or dest is an incorrect relative path", func() {
			relDir := filepath.Join(tmpDir, "rel")
			err := os.Mkdir(relDir, modePermUser)
			So(err, ShouldBeNil)

			err = os.Chdir(relDir)
			So(err, ShouldBeNil)

			err = Up("../src", "../dest", date)
			So(err, ShouldBeNil)

			err = os.RemoveAll(relDir)
			So(err, ShouldBeNil)

			err = Up("../src", "../dest", date)
			So(err, ShouldNotBeNil)
		})
	})
}

func buildSrcDir(srcDir, srcUniversal, srcUniqueGo, srcUniquePerl, interestUniqueDir1, interestUniqueDir2 string) {
	walkFileSuffixes := [5]string{".log", ".stats", ".byusergroup", ".bygroup", ".dgut"}
	combineFileSuffixes := [4]string{"combine.log.gz", "combine.byusergroup.gz", "combine.bygroup", "combine.stats.gz"}

	for i := range walkFileSuffixes {
		createTestPath([]string{interestUniqueDir1}, "walk.1"+walkFileSuffixes[i])
		createTestPath([]string{interestUniqueDir2}, "walk.1"+walkFileSuffixes[i])
	}

	for i := range combineFileSuffixes {
		createTestPath([]string{interestUniqueDir1}, "walk.1"+combineFileSuffixes[i])
		createTestPath([]string{interestUniqueDir2}, "walk.1"+combineFileSuffixes[i])
	}

	goDBDir := []string{srcDir, srcUniversal, "go", srcUniqueGo, "combine.dgut.db"}
	perlDBDir := []string{srcDir, srcUniversal, "perl", srcUniquePerl, "combine.dgut.db"}
	combineDirSuffixes := [3]string{"dgut.db", "dgut.db.children", "combine.bygroup"}

	for i := range combineDirSuffixes {
		createTestPath(goDBDir, combineDirSuffixes[i])
		createTestPath(perlDBDir, combineDirSuffixes[i])
	}

	createTestPath([]string{srcDir, srcUniversal}, "base.dirs")
}

// createTestPath takes a set of subdirectory names and an optional file
// basename and creates a directory and empty file out of them. Returns the
// directory.
func createTestPath(dirs []string, basename ...string) string {
	wholeDir := filepath.Join(dirs...)

	err := os.MkdirAll(wholeDir, modePermUser)
	So(err, ShouldBeNil)

	if len(basename) == 1 {
		createFile(filepath.Join(wholeDir, basename[0]))
	}

	return wholeDir
}

// createFile creates an empty file in the path provided by the user.
func createFile(fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		So(err, ShouldBeNil)

		return
	}

	err = f.Close()
	So(err, ShouldBeNil)
}

// createTestDirWithDifferentPerms creates the given directory with different
// group ownership and rw permissions than normal.
func createTestDirWithDifferentPerms(dir string) {
	err := os.MkdirAll(dir, 0777)
	So(err, ShouldBeNil)

	destUID := os.Getuid()
	destGroups, err := os.Getgroups()
	So(err, ShouldBeNil)

	err = os.Lchown(dir, destUID, destGroups[1])
	So(err, ShouldBeNil)
}

// getMTime takes a filePath and returns its Mtime.
func getMTime(filePath string) time.Time {
	FileInfo, err := os.Stat(filePath)
	So(err, ShouldBeNil)

	fileMTime := FileInfo.ModTime()

	return fileMTime
}

// permissionsAndOwnershipSame takes two fileinfos and returns whether their permissions and ownerships are the same.
func permissionsAndOwnershipSame(a, b fs.FileInfo) bool {
	return readWritePermissionsSame(a, b) && userAndGroupOwnershipSame(a, b)
}

// userAndGroupOwnershipSame tests if the given fileinfos have the same UID and
// GID.
func userAndGroupOwnershipSame(a, b fs.FileInfo) bool {
	aUID, aGID := getUIDAndGID(a)
	bUID, bGID := getUIDAndGID(b)

	return aUID == bUID && aGID == bGID
}

// getUIDAndGID extracts the UID and GID from a FileInfo. NB: this will only
// work on linux.
func getUIDAndGID(info fs.FileInfo) (int, int) {
	return int(info.Sys().(*syscall.Stat_t).Uid), int(info.Sys().(*syscall.Stat_t).Gid) //nolint:forcetypeassert
}

// matchReadWrite ensures that the given file with the current fileinfo has the
// same user,group,other read&write permissions as the desired fileinfo.
func readWritePermissionsSame(a, b fs.FileInfo) bool {
	aMode := a.Mode()
	aRW := aMode & modeRW
	bRW := b.Mode() & modeRW

	return aRW == bRW
}
