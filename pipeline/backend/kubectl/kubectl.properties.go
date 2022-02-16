package kubectl

// The run id
func (this *KubeCtlBackend) ID() string {
	return "wp-" + this.RunID
}

// The run namespace
func (this *KubeCtlBackend) Namespace() string {
	return this.Client.CoreArgs.Namespace
}
