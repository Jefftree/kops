/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package etcdmanager

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestJoin(t *testing.T) {
	tests := []struct {
		base        string
		other1      string
		other2      string
		expectedStr string
	}{
		{
			base:        "/test",
			other1:      "z1",
			other2:      "/z2",
			expectedStr: "/test/z1/z2",
		},
		{
			base:        "test/",
			other1:      "z1",
			other2:      "/z2",
			expectedStr: "test/z1/z2",
		},
	}
	for _, test := range tests {
		result := join(test.base, test.other1, test.other2)
		if test.expectedStr != result {
			t.Errorf("Expected %s, got %s", test.expectedStr, result)
		}
	}
}

func TestEtcdVersionsForCluster(t *testing.T) {
	findVersion := func(versions []etcdVersion, v string) *etcdVersion {
		for i := range versions {
			if versions[i].Version == v {
				return &versions[i]
			}
		}
		return nil
	}

	t.Run("no override returns the supported set verbatim", func(t *testing.T) {
		got := etcdVersionsForCluster(kops.EtcdClusterSpec{Version: "3.6.11"})
		want := etcdSupportedVersions()
		if len(got) != len(want) {
			t.Fatalf("expected %d versions, got %d", len(want), len(got))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("index %d: got %+v, want %+v", i, got[i], want[i])
			}
		}
	})

	t.Run("override replaces image for a supported version and drops symlinks targeting it", func(t *testing.T) {
		const customImage = "registry.k8s.io/etcd:v3.6.11"
		got := etcdVersionsForCluster(kops.EtcdClusterSpec{Version: "3.6.11", Image: customImage})

		v := findVersion(got, "3.6.11")
		if v == nil {
			t.Fatalf("expected version 3.6.11 in result")
		}
		if v.Image != customImage {
			t.Errorf("expected image %q, got %q", customImage, v.Image)
		}
		if v.SymlinkToVersion != "" {
			t.Errorf("expected overridden version to have empty SymlinkToVersion, got %q", v.SymlinkToVersion)
		}
		for _, e := range got {
			if e.SymlinkToVersion == "3.6.11" {
				t.Errorf("did not expect any entry to symlink to overridden version 3.6.11: %+v", e)
			}
		}
	})

	t.Run("override adds an entry for an unsupported version", func(t *testing.T) {
		const customImage = "registry.k8s.io/etcd:v3.7.0-beta.0"
		got := etcdVersionsForCluster(kops.EtcdClusterSpec{Version: "v3.7.0-beta.0", Image: customImage})

		v := findVersion(got, "3.7.0-beta.0")
		if v == nil {
			t.Fatalf("expected version 3.7.0-beta.0 in result")
		}
		if v.Image != customImage {
			t.Errorf("expected image %q, got %q", customImage, v.Image)
		}
		if v.SymlinkToVersion != "" {
			t.Errorf("expected overridden version to have empty SymlinkToVersion, got %q", v.SymlinkToVersion)
		}
	})

	t.Run("override with empty version falls back to the bundled set", func(t *testing.T) {
		got := etcdVersionsForCluster(kops.EtcdClusterSpec{Image: "registry.k8s.io/etcd:v3.7.0-beta.0"})
		want := etcdSupportedVersions()
		if len(got) != len(want) {
			t.Fatalf("expected %d versions, got %d", len(want), len(got))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("index %d: got %+v, want %+v", i, got[i], want[i])
			}
		}
	})
}
