package config

type workspace struct{}

var Workspace = workspace{}

func (*workspace) GetClientId() (string, error) {
	return GetStringConfigOrPrintErr(
		"workspace.client-id",
		"workspace.client-id not found",
	)
}
