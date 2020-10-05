// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"reflect"
	"testing"
)

func Test_releases(t *testing.T) {
	type args struct {
		packageFiles []string
	}
	tests := []struct {
		name    string
		args    args
		want    []Release
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				packageFiles: []string{
					`README`,
					`circonus-agent-1.0.10-1.el6.x86_64.rpm`,
					`circonus-agent-1.0.10-1.el7.x86_64.rpm`,
					`circonus-agent-1.0.10-1.el8.x86_64.rpm`,
					`circonus-agent-1.0.10-1.freebsd.11.3_amd64.tgz`,
					`circonus-agent-1.0.10-1.freebsd.12.1_amd64.tgz`,
					`circonus-agent-1.0.10-1.ubuntu.14.04_x86_64.deb`,
					`circonus-agent-1.0.10-1.ubuntu.16.04_x86_64.deb`,
					`circonus-agent-1.0.10-1.ubuntu.18.04_x86_64.deb`,
					`circonus-agent-1.0.10-1.ubuntu.20.04_x86_64.deb`,
					`circonus-agent-1.0.6-1.el6.x86_64.rpm`,
					`circonus-agent-1.0.6-1.el7.x86_64.rpm`,
					`circonus-agent-1.0.6-1.el8.x86_64.rpm`,
					`circonus-agent-1.0.6-1.freebsd.11.3_amd64.tgz`,
					`circonus-agent-1.0.6-1.freebsd.12.1_amd64.tgz`,
					`circonus-agent-1.0.6-1.ubuntu.14.04_x86_64.deb`,
					`circonus-agent-1.0.6-1.ubuntu.16.04_x86_64.deb`,
					`circonus-agent-1.0.6-1.ubuntu.18.04_x86_64.deb`,
					`circonus-agent-1.0.6-1.ubuntu.20.04_x86_64.deb`,
					`circonus-agent-1.0.13-1.el6.x86_64.rpm`,
					`circonus-agent-1.0.13-1.el7.x86_64.rpm`,
					`circonus-agent-1.0.13-1.el8.x86_64.rpm`,
					`circonus-agent-1.0.13-1.freebsd.11.3_amd64.tgz`,
					`circonus-agent-1.0.13-1.freebsd.12.1_amd64.tgz`,
					`circonus-agent-1.0.13-1.ubuntu.14.04_x86_64.deb`,
					`circonus-agent-1.0.13-1.ubuntu.16.04_x86_64.deb`,
					`circonus-agent-1.0.13-1.ubuntu.18.04_x86_64.deb`,
					`circonus-agent-1.0.13-1.ubuntu.20.04_x86_64.deb`,
					`circonus-agent-1.0.5-1.el6.x86_64.rpm`,
					`circonus-agent-1.0.5-1.el7.x86_64.rpm`,
					`circonus-agent-1.0.5-1.el8.x86_64.rpm`,
					`circonus-agent-1.0.5-1.freebsd.11.3_amd64.tgz`,
					`circonus-agent-1.0.5-1.freebsd.12.1_amd64.tgz`,
					`circonus-agent-1.0.5-1.ubuntu.14.04_x86_64.deb`,
					`circonus-agent-1.0.5-1.ubuntu.16.04_x86_64.deb`,
					`circonus-agent-1.0.5-1.ubuntu.18.04_x86_64.deb`,
					`circonus-agent-1.0.5-1.ubuntu.20.04_x86_64.deb`,
					`circonus-agent-1.0.7-1.el6.x86_64.rpm`,
					`circonus-agent-1.0.7-1.el7.x86_64.rpm`,
					`circonus-agent-1.0.7-1.el8.x86_64.rpm`,
					`circonus-agent-1.0.7-1.freebsd.11.3_amd64.tgz`,
					`circonus-agent-1.0.7-1.freebsd.12.1_amd64.tgz`,
					`circonus-agent-1.0.7-1.ubuntu.14.04_x86_64.deb`,
					`circonus-agent-1.0.7-1.ubuntu.16.04_x86_64.deb`,
					`circonus-agent-1.0.7-1.ubuntu.18.04_x86_64.deb`,
					`circonus-agent-1.0.7-1.ubuntu.20.04_x86_64.deb`,
					`circonus-agentd`,
					`metrics_20200716_120112.json`,
					`circonus-agent-0.19.5-1.el6.x86_64.rpm`,
					`circonus-agent-0.19.5-1.el7.centos.x86_64.rpm`,
					`circonus-agent-0.19.5-1.freebsd.11.2_amd64.tgz`,
					`circonus-agent-0.19.5-1.ubuntu.16.04_x86_64.deb`,
					`circonus-agent-0.19.5-1.ubuntu.18.04_x86_64.deb`,
					`circonus-agent-1.0.9-1.el6.x86_64.rpm`,
					`circonus-agent-1.0.9-1.el7.x86_64.rpm`,
					`circonus-agent-1.0.9-1.el8.x86_64.rpm`,
					`circonus-agent-1.0.9-1.freebsd.11.3_amd64.tgz`,
					`circonus-agent-1.0.9-1.freebsd.12.1_amd64.tgz`,
					`circonus-agent-1.0.9-1.ubuntu.14.04_x86_64.deb`,
					`circonus-agent-1.0.9-1.ubuntu.16.04_x86_64.deb`,
					`circonus-agent-1.0.9-1.ubuntu.18.04_x86_64.deb`,
					`circonus-agent-1.0.9-1.ubuntu.20.04_x86_64.deb`,
				},
			},
			want: []Release{
				{Version: "1.0.13", Packages: []string{"circonus-agent-1.0.13-1.el6.x86_64.rpm", "circonus-agent-1.0.13-1.el7.x86_64.rpm", "circonus-agent-1.0.13-1.el8.x86_64.rpm", "circonus-agent-1.0.13-1.freebsd.11.3_amd64.tgz", "circonus-agent-1.0.13-1.freebsd.12.1_amd64.tgz", "circonus-agent-1.0.13-1.ubuntu.14.04_x86_64.deb", "circonus-agent-1.0.13-1.ubuntu.16.04_x86_64.deb", "circonus-agent-1.0.13-1.ubuntu.18.04_x86_64.deb", "circonus-agent-1.0.13-1.ubuntu.20.04_x86_64.deb"}},
				{Version: "1.0.10", Packages: []string{"circonus-agent-1.0.10-1.el6.x86_64.rpm", "circonus-agent-1.0.10-1.el7.x86_64.rpm", "circonus-agent-1.0.10-1.el8.x86_64.rpm", "circonus-agent-1.0.10-1.freebsd.11.3_amd64.tgz", "circonus-agent-1.0.10-1.freebsd.12.1_amd64.tgz", "circonus-agent-1.0.10-1.ubuntu.14.04_x86_64.deb", "circonus-agent-1.0.10-1.ubuntu.16.04_x86_64.deb", "circonus-agent-1.0.10-1.ubuntu.18.04_x86_64.deb", "circonus-agent-1.0.10-1.ubuntu.20.04_x86_64.deb"}},
				{Version: "1.0.9", Packages: []string{"circonus-agent-1.0.9-1.el6.x86_64.rpm", "circonus-agent-1.0.9-1.el7.x86_64.rpm", "circonus-agent-1.0.9-1.el8.x86_64.rpm", "circonus-agent-1.0.9-1.freebsd.11.3_amd64.tgz", "circonus-agent-1.0.9-1.freebsd.12.1_amd64.tgz", "circonus-agent-1.0.9-1.ubuntu.14.04_x86_64.deb", "circonus-agent-1.0.9-1.ubuntu.16.04_x86_64.deb", "circonus-agent-1.0.9-1.ubuntu.18.04_x86_64.deb", "circonus-agent-1.0.9-1.ubuntu.20.04_x86_64.deb"}},
				{Version: "1.0.7", Packages: []string{"circonus-agent-1.0.7-1.el6.x86_64.rpm", "circonus-agent-1.0.7-1.el7.x86_64.rpm", "circonus-agent-1.0.7-1.el8.x86_64.rpm", "circonus-agent-1.0.7-1.freebsd.11.3_amd64.tgz", "circonus-agent-1.0.7-1.freebsd.12.1_amd64.tgz", "circonus-agent-1.0.7-1.ubuntu.14.04_x86_64.deb", "circonus-agent-1.0.7-1.ubuntu.16.04_x86_64.deb", "circonus-agent-1.0.7-1.ubuntu.18.04_x86_64.deb", "circonus-agent-1.0.7-1.ubuntu.20.04_x86_64.deb"}},
				{Version: "1.0.6", Packages: []string{"circonus-agent-1.0.6-1.el6.x86_64.rpm", "circonus-agent-1.0.6-1.el7.x86_64.rpm", "circonus-agent-1.0.6-1.el8.x86_64.rpm", "circonus-agent-1.0.6-1.freebsd.11.3_amd64.tgz", "circonus-agent-1.0.6-1.freebsd.12.1_amd64.tgz", "circonus-agent-1.0.6-1.ubuntu.14.04_x86_64.deb", "circonus-agent-1.0.6-1.ubuntu.16.04_x86_64.deb", "circonus-agent-1.0.6-1.ubuntu.18.04_x86_64.deb", "circonus-agent-1.0.6-1.ubuntu.20.04_x86_64.deb"}},
				{Version: "1.0.5", Packages: []string{"circonus-agent-1.0.5-1.el6.x86_64.rpm", "circonus-agent-1.0.5-1.el7.x86_64.rpm", "circonus-agent-1.0.5-1.el8.x86_64.rpm", "circonus-agent-1.0.5-1.freebsd.11.3_amd64.tgz", "circonus-agent-1.0.5-1.freebsd.12.1_amd64.tgz", "circonus-agent-1.0.5-1.ubuntu.14.04_x86_64.deb", "circonus-agent-1.0.5-1.ubuntu.16.04_x86_64.deb", "circonus-agent-1.0.5-1.ubuntu.18.04_x86_64.deb", "circonus-agent-1.0.5-1.ubuntu.20.04_x86_64.deb"}},
				{Version: "0.19.5", Packages: []string{"circonus-agent-0.19.5-1.el6.x86_64.rpm", "circonus-agent-0.19.5-1.el7.centos.x86_64.rpm", "circonus-agent-0.19.5-1.freebsd.11.2_amd64.tgz", "circonus-agent-0.19.5-1.ubuntu.16.04_x86_64.deb", "circonus-agent-0.19.5-1.ubuntu.18.04_x86_64.deb"}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := releases(tt.args.packageFiles)
			if (err != nil) != tt.wantErr {
				t.Errorf("releases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("releases() = %#v, want %v", got, tt.want)
			}
		})
	}
}
