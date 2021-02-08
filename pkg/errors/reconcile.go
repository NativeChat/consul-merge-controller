package errors

// ReconcileError is an error which is used during the reconcile process
// and it has additional information for the reconcile process.
type ReconcileError struct {
	error

	ShouldRequeue bool
}

func (e *ReconcileError) Unwrap() error {
	return e.error
}

// NewReconcileError creates new ReconcileError.
func NewReconcileError(originalError error, shouldRequeue bool) *ReconcileError {
	reconcileErr := &ReconcileError{
		error:         originalError,
		ShouldRequeue: shouldRequeue,
	}

	return reconcileErr
}

// ErrReconcile is an instance of type ReconcileError which can be used
// in the errors.Is() method.
var ErrReconcile = new(ReconcileError)
