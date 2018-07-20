// +build linux

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless ruired by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package process

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestAdaptRootfsPath(t *testing.T) {
	rootfs := "/foo"
	path := "/bin:/usr/bin"
	expected := "/foo/bin:/foo/usr/bin"

	res := adaptRootfsPath(rootfs, path)

	if res != expected {
		t.Fatalf("expected path to be prefixed; received %s", res)
	}
}

func TestAdaptRootfsLDLibraryPath(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "containerd-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	testdir1 := "/foo"
	testdir2 := "/bar"
	confPath := filepath.Join(tmpdir, ldConfPath, "test")
	if err := os.MkdirAll(filepath.Dir(confPath), 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(confPath)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("# this should be skipped\n"))
	f.Write([]byte(fmt.Sprintf("%s\n", testdir1)))
	f.Write([]byte(fmt.Sprintf("%s\n", testdir2)))
	f.Close()

	res, err := resolveRootfsLDLibraryPath(tmpdir)
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(tmpdir, testdir1) + ":" + filepath.Join(tmpdir, testdir2)
	if res != expected {
		t.Fatalf("expected %s; received %s", expected, res)
	}
}

func TestAdaptRootfsLDLibraryPathNoPresent(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "containerd-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	res, err := resolveRootfsLDLibraryPath(tmpdir)
	if err != nil {
		t.Fatal(err)
	}

	expected := ""
	if res != expected {
		t.Fatalf("expected %s; received %s", expected, res)
	}
}
