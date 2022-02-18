package kubectl

// The run id
func (backend *KubeCtlBackend) ID() string {
	return toKuberenetesValidName("wp-"+backend.RunID, 30)
}

// The run namespace
func (backend *KubeCtlBackend) Namespace() string {
	return backend.Client.CoreArgs.Namespace
}

func (backend *KubeCtlBackend) Context() string {
	return backend.Client.CoreArgs.Context
}
