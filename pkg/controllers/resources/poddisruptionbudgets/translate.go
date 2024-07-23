package poddisruptionbudgets

import (
	"github.com/loft-sh/vcluster/pkg/syncer/synccontext"
	"github.com/loft-sh/vcluster/pkg/util/translate"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (s *pdbSyncer) translate(ctx *synccontext.SyncContext, vObj *policyv1.PodDisruptionBudget) *policyv1.PodDisruptionBudget {
	newPDB := translate.HostMetadata(ctx, vObj, s.VirtualToHost(ctx, types.NamespacedName{Name: vObj.GetName(), Namespace: vObj.GetNamespace()}, vObj))
	newPDB.Spec.Selector = translate.HostLabelSelector(ctx, newPDB.Spec.Selector)
	return newPDB
}

func (s *pdbSyncer) translateUpdate(ctx *synccontext.SyncContext, pObj, vObj *policyv1.PodDisruptionBudget) {
	pObj.Annotations = translate.HostAnnotations(vObj, pObj)
	pObj.Labels = translate.HostLabels(ctx, vObj, pObj)
	pObj.Spec.MaxUnavailable = vObj.Spec.MaxUnavailable
	pObj.Spec.MinAvailable = vObj.Spec.MinAvailable
	pObj.Spec.Selector = translate.HostLabelSelector(ctx, vObj.Spec.Selector)
}
