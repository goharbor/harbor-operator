package v1beta1

import (
	"encoding/json"
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"testing"
)

func TestHarborCluster_ConvertTo(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       HarborClusterSpec
		Status     HarborClusterStatus
	}
	type args struct {
		dstRaw conversion.Hub
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "validatingHarborClusterConversion",
			fields: fields{
				TypeMeta: v1.TypeMeta{
					Kind:       "HarborCluster",
					APIVersion: "goharbor.io/v1beta1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "sample",
				},
				Spec: HarborClusterSpec{
					HarborSpec: HarborSpec{},
					Cache:      nil,
					Database: &Database{
						Kind: "PostgreSQL",
						Spec: DatabaseSpec{
							PostgreSQL: &PostgreSQLSpec{
								HarborDatabaseSpec{
									PostgresCredentials: harbormetav1.PostgresCredentials{
										Username:    "postgres",
										PasswordRef: "harbor-database-password",
									},
									Hosts: []harbormetav1.PostgresHostSpec{
										{
											Host: "harbor-database-postgresql",
											Port: 5432,
										},
									},
									SSLMode: "disable",
									Prefix:  "",
								},
							},
						},
					},
					Storage: nil,
				},
				Status: HarborClusterStatus{},
			},
			args: args{
				dstRaw: &v1alpha3.HarborCluster{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := &HarborCluster{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if err := src.ConvertTo(tt.args.dstRaw); (err != nil) != tt.wantErr {
				t.Errorf("ConvertTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			d, _ := json.Marshal(tt.args.dstRaw)
			t.Logf("v1alpha3: %s", d)

		})
	}
}
