package v1beta1

import (
	"time"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (m *MTLSSpec) Validate() (admission.Warnings, field.ErrorList) {
	var warns admission.Warnings
	var errs field.ErrorList

	if m == nil || m.Provider != CertManagerMTLSProvider {
		return nil, nil
	}

	if m.RenewBefore != nil {
		if m.RenewBefore.Duration < 5*time.Minute {
			errs = append(errs, field.Invalid(field.NewPath("spec.mTLS.renewBefore"), m.RenewBefore, "must be at least 5 minutes"))
		}
	}

	return warns, errs
}
