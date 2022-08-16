package persistentvolumes

import (
	"testing"

	synccontext "github.com/loft-sh/vcluster/pkg/controllers/syncer/context"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/api/resource"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/loft-sh/vcluster/pkg/constants"
	generictesting "github.com/loft-sh/vcluster/pkg/controllers/syncer/testing"
	"github.com/loft-sh/vcluster/pkg/util/translate"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func newFakeSyncer(t *testing.T, ctx *synccontext.RegisterContext) (*synccontext.SyncContext, *persistentVolumeSyncer) {
	err := ctx.VirtualManager.GetFieldIndexer().IndexField(ctx.Context, &corev1.PersistentVolumeClaim{}, constants.IndexByPhysicalName, func(rawObj client.Object) []string {
		return []string{translate.ObjectPhysicalName(rawObj)}
	})
	assert.NilError(t, err)

	syncContext, object := generictesting.FakeStartSyncer(t, ctx, NewSyncer)
	return syncContext, object.(*persistentVolumeSyncer)
}

func TestSync(t *testing.T) {
	basePvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpvc",
			Namespace: "test",
		},
	}
	basePPvcReference := &corev1.ObjectReference{
		Name:            translate.PhysicalName("testpvc", "test"),
		Namespace:       "test",
		ResourceVersion: generictesting.FakeClientResourceVersion,
	}
	baseVPvcReference := &corev1.ObjectReference{
		Name:            "testpvc",
		Namespace:       "test",
		ResourceVersion: generictesting.FakeClientResourceVersion,
	}
	basePvObjectMeta := metav1.ObjectMeta{
		Name: "testpv",
		Annotations: map[string]string{
			HostClusterPersistentVolumeAnnotation: "testpv",
		},
	}
	basePPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef: basePPvcReference,
		},
	}
	baseVPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef: baseVPvcReference,
		},
	}
	wrongNsPPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef: &corev1.ObjectReference{
				Name:            "testpvc",
				Namespace:       "wrong",
				ResourceVersion: generictesting.FakeClientResourceVersion,
			},
		},
	}
	noPvcPPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef: &corev1.ObjectReference{
				Name:      "wrong",
				Namespace: "test",
			},
		},
	}
	backwardUpdatePPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef:         basePPvcReference,
			StorageClassName: "someStorageClass",
		},
		Status: corev1.PersistentVolumeStatus{
			Message: "someMessage",
		},
	}
	backwardUpdateVPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef:         baseVPvcReference,
			StorageClassName: "someStorageClass",
		},
		Status: corev1.PersistentVolumeStatus{
			Message: "someMessage",
		},
	}
	// Reclaim policy retain for virtual PV
	baseRPRetainVPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef:                      baseVPvcReference,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			StorageClassName:              "someStorageClass",
		},
	}
	baseRPRetainBoundVPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef:                      baseVPvcReference,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			StorageClassName:              "someStorageClass",
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: corev1.VolumeBound,
		},
	}

	baseRPRetainReleasedVPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef:                      baseVPvcReference,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			StorageClassName:              "someStorageClass",
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: corev1.VolumeReleased,
		},
	}

	// Reclaim policy delete for virtual PV
	baseRPDeleteVPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef:                      baseVPvcReference,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: corev1.VolumeReleased,
		},
	}

	// Reclaim policy retain for physical PV
	baseRPRetainPPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef:                      basePPvcReference,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			// StorageClassName:              translate.PhysicalName("someStorageClass", ""),
		},
	}
	// Reclaim policy retain for physical PV, with bound status
	baseRPRetainBoundPPv := baseRPRetainPPv.DeepCopy()
	baseRPRetainBoundPPv.Status = corev1.PersistentVolumeStatus{
		Phase: corev1.VolumeBound,
	}

	// Reclaim policy retain for physical PV, with released status
	baseRPRetainReleasedPPv := baseRPRetainPPv.DeepCopy()
	baseRPRetainReleasedPPv.Status = corev1.PersistentVolumeStatus{
		Phase: corev1.VolumeReleased,
	}

	// Reclaim policy delete for physical PV
	baseRPDeletePPv := &corev1.PersistentVolume{
		ObjectMeta: basePvObjectMeta,
		Spec: corev1.PersistentVolumeSpec{
			ClaimRef:                      basePPvcReference,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: corev1.VolumeReleased,
		},
	}
	basePvcSC := basePvc.DeepCopy()
	basePvcSC.Spec.StorageClassName = &baseRPRetainVPv.Spec.StorageClassName

	generictesting.RunTests(t, []*generictesting.SyncTest{
		{
			Name:                 "Update Status Bound To Released Upwards:",
			InitialVirtualState:  []runtime.Object{baseRPRetainBoundVPv},
			InitialPhysicalState: []runtime.Object{baseRPRetainReleasedPPv},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {baseRPRetainReleasedVPv},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {baseRPRetainReleasedPPv},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)
				_, err := syncer.Sync(syncContext, baseRPRetainReleasedPPv, baseRPRetainBoundVPv)
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Create ReclaimPolicy Retain Upwards:",
			InitialVirtualState:  []runtime.Object{basePvcSC},
			InitialPhysicalState: []runtime.Object{basePvcSC, baseRPRetainPPv},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"):      {baseRPRetainVPv},
				corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"): {basePvcSC},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"): {basePvcSC},
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"):      {baseRPRetainPPv},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)
				_, err := syncer.SyncUp(syncContext, baseRPRetainPPv)
				assert.NilError(t, err)
			},
		},
		{
			Name:                  "Delete vPv,pPv When ReclaimPolicy Delete And Volume Released:",
			InitialVirtualState:   []runtime.Object{basePvc, baseRPDeleteVPv},
			InitialPhysicalState:  []runtime.Object{basePvc, baseRPDeletePPv},
			ExpectedVirtualState:  map[schema.GroupVersionKind][]runtime.Object{},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)
				_, err := syncer.Sync(syncContext, baseRPDeletePPv, baseRPDeleteVPv)
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Create Backward",
			InitialVirtualState:  []runtime.Object{basePvc},
			InitialPhysicalState: []runtime.Object{basePPv},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"):      {baseVPv},
				corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"): {basePvc},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {basePPv},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)
				_, err := syncer.SyncUp(syncContext, basePPv)
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Don't Create Backward, wrong physical namespace",
			InitialVirtualState:  []runtime.Object{basePvc},
			InitialPhysicalState: []runtime.Object{wrongNsPPv},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"):      {},
				corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"): {basePvc},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {wrongNsPPv},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)
				_, err := syncer.SyncUp(syncContext, wrongNsPPv)
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Don't Create Backward, no virtual pvc",
			InitialVirtualState:  []runtime.Object{basePvc},
			InitialPhysicalState: []runtime.Object{noPvcPPv},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"):      {},
				corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"): {basePvc},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {noPvcPPv},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)
				_, err := syncer.SyncUp(syncContext, noPvcPPv)
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Update Backward",
			InitialVirtualState:  []runtime.Object{basePvc, baseVPv},
			InitialPhysicalState: []runtime.Object{backwardUpdatePPv},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"):      {backwardUpdateVPv},
				corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"): {basePvc},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {backwardUpdatePPv},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)
				backwardUpdatePPv := backwardUpdatePPv.DeepCopy()
				baseVPv := baseVPv.DeepCopy()
				_, err := syncer.Sync(syncContext, backwardUpdatePPv, baseVPv)
				assert.NilError(t, err)

				err = syncContext.VirtualClient.Get(ctx.Context, types.NamespacedName{Name: baseVPv.Name}, baseVPv)
				assert.NilError(t, err)

				err = syncContext.PhysicalClient.Get(ctx.Context, types.NamespacedName{Name: backwardUpdatePPv.Name}, backwardUpdatePPv)
				assert.NilError(t, err)

				_, err = syncer.Sync(syncContext, backwardUpdatePPv, baseVPv)
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Delete Backward by update backward",
			InitialVirtualState:  []runtime.Object{basePvc, baseVPv},
			InitialPhysicalState: []runtime.Object{noPvcPPv},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"):      {},
				corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"): {basePvc},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {noPvcPPv},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)
				_, err := syncer.Sync(syncContext, noPvcPPv, baseVPv)
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Delete Backward not needed",
			InitialVirtualState:  []runtime.Object{basePvc, baseVPv},
			InitialPhysicalState: []runtime.Object{basePPv},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"):      {baseVPv},
				corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"): {basePvc},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {basePPv},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)
				_, err := syncer.Sync(syncContext, basePPv, baseVPv)
				assert.NilError(t, err)
			},
		},
		{
			Name: "Sync PV Size",
			InitialVirtualState: []runtime.Object{
				&corev1.PersistentVolumeClaim{
					ObjectMeta: basePvc.ObjectMeta,
				},
				&corev1.PersistentVolume{
					ObjectMeta: baseVPv.ObjectMeta,
					Spec: corev1.PersistentVolumeSpec{
						Capacity: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("5Gi"),
						},
						ClaimRef: baseVPv.Spec.ClaimRef,
					},
				},
			},
			InitialPhysicalState: []runtime.Object{
				&corev1.PersistentVolume{
					ObjectMeta: basePPv.ObjectMeta,
					Spec: corev1.PersistentVolumeSpec{
						Capacity: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("20Gi"),
						},
						ClaimRef: basePPv.Spec.ClaimRef,
					},
				},
			},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {
					&corev1.PersistentVolume{
						ObjectMeta: baseVPv.ObjectMeta,
						Spec: corev1.PersistentVolumeSpec{
							Capacity: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("20Gi"),
							},
							ClaimRef: baseVPv.Spec.ClaimRef,
						},
					},
				},
				corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"): {
					&corev1.PersistentVolumeClaim{
						ObjectMeta: basePvc.ObjectMeta,
					},
				},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("PersistentVolume"): {
					&corev1.PersistentVolume{
						ObjectMeta: basePPv.ObjectMeta,
						Spec: corev1.PersistentVolumeSpec{
							Capacity: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("20Gi"),
							},
							ClaimRef: basePPv.Spec.ClaimRef,
						},
					},
				},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncContext, syncer := newFakeSyncer(t, ctx)

				vPv := &corev1.PersistentVolume{}
				err := syncContext.VirtualClient.Get(ctx.Context, types.NamespacedName{Name: baseVPv.Name}, vPv)
				assert.NilError(t, err)

				pPv := &corev1.PersistentVolume{}
				err = syncContext.PhysicalClient.Get(ctx.Context, types.NamespacedName{Name: basePPv.Name}, pPv)
				assert.NilError(t, err)

				_, err = syncer.Sync(syncContext, pPv, vPv)
				assert.NilError(t, err)
			},
		},
	})
}
