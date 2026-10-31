package autostart

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// On, the system's record names this program and the tray; off, it's gone,
// and turning it off again is no error.
func TestSet(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the Run key is the real user's")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", home+"/.config")
	t.Setenv("APPIMAGE", "/opt/my apps/magpie.AppImage")
	// the image is taken as magpie's own only when magpie runs under its
	// mount, APPDIR; the test binary stands in for the mounted magpie
	if exe, err := os.Executable(); err == nil {
		if p, err := filepath.EvalSymlinks(exe); err == nil {
			exe = p
		}
		t.Setenv("APPDIR", filepath.Dir(exe))
	}
	if Enabled() != Off {
		t.Fatal("on before it was turned on")
	}
	if err := Set(true); err != nil {
		t.Fatal(err)
	}
	if Enabled() == Off {
		t.Fatal("not on after Set(true)")
	}
	path := record()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	arg := Arg
	if runtime.GOOS == "android" {
		arg = "serve"
	}
	for _, want := range []string{"/opt/my apps/magpie.AppImage", arg} {
		if !strings.Contains(string(b), want) {
			t.Errorf("%s lacks %q:\n%s", path, want, b)
		}
	}
	if err := Set(false); err != nil || Enabled() != Off {
		t.Fatalf("still on after Set(false): %v", err)
	}
	if err := Set(false); err != nil {
		t.Fatal(err)
	}
}

// A magpie started from a terminal inside another app's AppImage inherits
// that app's APPIMAGE; the record must still name magpie itself, not the
// other app (#4490).
func TestSelfTakesOnlyItsOwnAppImage(t *testing.T) {
	const foreign = "/opt/OpenChamber.AppImage"
	t.Setenv("APPIMAGE", foreign)
	t.Setenv("APPDIR", "/tmp/.mount_OpenChVGUuXr")
	exe, err := self()
	if err != nil {
		t.Fatal(err)
	}
	if exe == foreign {
		t.Fatalf("the other app's AppImage was taken for magpie's own: %q", exe)
	}
	want, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if p, err := filepath.EvalSymlinks(want); err == nil {
		want = p
	}
	if exe != want {
		t.Errorf("self %q, want this program %q", exe, want)
	}
}
