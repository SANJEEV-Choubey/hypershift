package hostedcontrolplane

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	hyperv1 "github.com/openshift/hypershift/api/v1alpha1"
	"github.com/openshift/hypershift/control-plane-operator/controllers/hostedapicache"
	"github.com/openshift/hypershift/support/capabilities"
	"github.com/openshift/hypershift/support/releaseinfo"
	"github.com/openshift/hypershift/support/upsert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type releaseProvider struct{}

func (r releaseProvider) Lookup(ctx context.Context, image string, pullSecret []byte) (*releaseinfo.ReleaseImage, error) {
	hcp := &HostedControlPlaneReconciler{}

	return hcp.ReleaseProvider.Lookup(ctx, image, pullSecret)
}

type hostedAPICache struct{}

func (h hostedAPICache) Get(ctx context.Context, key types.NamespacedName, obj client.Object) error {
	return nil
}

func (h hostedAPICache) Events() <-chan event.GenericEvent {
	return nil
}
func (h hostedAPICache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}
func TestReconcileKubeadminPassword(t *testing.T) {
	fakeClient := fake.NewClientBuilder().Build()

	type fields struct {
		Client                        client.Client
		ManagementClusterCapabilities *capabilities.ManagementClusterCapabilities
		SetDefaultSecurityContext     bool
		Log                           logr.Logger
		ReleaseProvider               releaseinfo.Provider
		HostedAPICache                hostedapicache.HostedAPICache
		CreateOrUpdateProvider        upsert.CreateOrUpdateProvider
		EnableCIDebugOutput           bool
	}
	type args struct {
		ctx                 context.Context
		hcp                 *hyperv1.HostedControlPlane
		explicitOauthConfig bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "global config oauth nil",
			fields: fields{
				Client:                        fakeClient,
				ManagementClusterCapabilities: &capabilities.ManagementClusterCapabilities{},
				SetDefaultSecurityContext:     false,
				Log:                           ctrl.LoggerFrom(context.TODO()),
				ReleaseProvider:               &releaseProvider{},
				HostedAPICache:                &hostedAPICache{},
				CreateOrUpdateProvider:        upsert.New(false),
				EnableCIDebugOutput:           false,
			},
			args: args{
				ctx: context.TODO(),
				hcp: &hyperv1.HostedControlPlane{
					TypeMeta: metav1.TypeMeta{
						Kind:       "HostedControlPlane",
						APIVersion: hyperv1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "master-cluster1",
						Name:      "cluster1",
					},
				},
				explicitOauthConfig: true,
			},
			wantErr: false,
		},

		{
			name: "global config oauth not nil",
			fields: fields{
				Client:                        fakeClient,
				ManagementClusterCapabilities: &capabilities.ManagementClusterCapabilities{},
				SetDefaultSecurityContext:     false,
				Log:                           ctrl.LoggerFrom(context.TODO()),
				ReleaseProvider:               &releaseProvider{},
				HostedAPICache:                &hostedAPICache{},
				CreateOrUpdateProvider:        upsert.New(false),
				EnableCIDebugOutput:           false,
			},
			args: args{
				ctx: context.TODO(),
				hcp: &hyperv1.HostedControlPlane{
					TypeMeta: metav1.TypeMeta{
						Kind:       "HostedControlPlane",
						APIVersion: hyperv1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "master-cluster1",
						Name:      "cluster1",
					},
				},
				explicitOauthConfig: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HostedControlPlaneReconciler{
				Client:                        tt.fields.Client,
				ManagementClusterCapabilities: tt.fields.ManagementClusterCapabilities,
				SetDefaultSecurityContext:     tt.fields.SetDefaultSecurityContext,
				Log:                           tt.fields.Log,
				ReleaseProvider:               tt.fields.ReleaseProvider,
				HostedAPICache:                tt.fields.HostedAPICache,
				CreateOrUpdateProvider:        tt.fields.CreateOrUpdateProvider,
				EnableCIDebugOutput:           tt.fields.EnableCIDebugOutput,
			}
			if err := r.reconcileKubeadminPassword(tt.args.ctx, tt.args.hcp, tt.args.explicitOauthConfig); (err != nil) != tt.wantErr {
				t.Errorf("ReconcileKubeadminPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
